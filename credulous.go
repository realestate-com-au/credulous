package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"code.google.com/p/go.crypto/ssh"

	"code.google.com/p/gopass"
	"github.com/codegangsta/cli"
)

const ENV_PATTERN string = "^[A-Za-z_][A-Za-z0-9_]*=.*"

func decryptPEM(pemblock *pem.Block, filename string) ([]byte, error) {
	var err error
	if _, err = fmt.Fprintf(os.Stderr, "Enter passphrase for %s: ", filename); err != nil {
		return []byte(""), err
	}

	// we already emit the prompt to stderr; GetPass only emits to stdout
	var passwd string
	passwd, err = gopass.GetPass("")
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return []byte(""), err
	}

	var decryptedBytes []byte
	if decryptedBytes, err = x509.DecryptPEMBlock(pemblock, []byte(passwd)); err != nil {
		return []byte(""), err
	}

	pemBytes := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: decryptedBytes,
	}
	decryptedPEM := pem.EncodeToMemory(&pemBytes)
	return decryptedPEM, nil
}

func getPrivateKey(c *cli.Context) (filename string) {
	if c.String("key") == "" {
		filename = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
	} else {
		filename = c.String("key")
	}
	return filename
}

func splitUserAndAccount(arg string) (string, string, error) {
	atpos := strings.LastIndex(arg, "@")
	if atpos < 1 {
		err := errors.New("Invalid account format; please specify <username>@<account>")
		return "", "", err
	}
	// pull off everything before the last '@'
	return arg[atpos+1:], arg[0:atpos], nil
}

func getAccountAndUserName(c *cli.Context) (string, string, error) {
	if len(c.Args()) > 0 {
		user, acct, err := splitUserAndAccount(c.Args()[0])
		if err != nil {
			return "", "", err
		}
		return user, acct, nil
	}
	if c.String("credentials") != "" {
		user, acct, err := splitUserAndAccount(c.String("credentials"))
		if err != nil {
			return "", "", err
		}
		return user, acct, nil
	} else {
		return c.String("account"), c.String("username"), nil
	}
}

func parseUserAndAccount(c *cli.Context) (username string, account string, err error) {
	if (c.String("username") == "" || c.String("account") == "") && c.Bool("force") {
		err = errors.New("Must specify both username and account with force")
		return "", "", err
	}

	// if username OR account were specified, but not both, complain
	if (c.String("username") != "" && c.String("account") == "") ||
		(c.String("username") == "" && c.String("account") != "") {
		if c.Bool("force") {
			err = errors.New("Must specify both username and account for force save")
		} else {
			err = errors.New("Must use force save when specifying username or account")
		}
		return "", "", err
	}

	// if username/account were specified, but force wasn't set, complain
	if c.String("username") != "" && c.String("account") != "" {
		if !c.Bool("force") {
			err = errors.New("Cannot specify username and/or account without force")
			return "", "", err
		} else {
			log.Print("WARNING: saving credentials without verifying username or account alias")
			username = c.String("username")
			account = c.String("account")
		}
	}
	return username, account, nil
}

func parseEnvironmentArgs(c *cli.Context) (map[string]string, error) {
	if len(c.StringSlice("env")) == 0 {
		return nil, nil
	}

	envMap := make(map[string]string)
	for _, arg := range c.StringSlice("env") {
		match, err := regexp.Match(ENV_PATTERN, []byte(arg))
		if err != nil {
			return nil, err
		}
		if !match {
			log.Print("WARNING: Skipping env argument " + arg + " -- not in NAME=value format")
			continue
		}
		parts := strings.SplitN(arg, "=", 2)
		envMap[parts[0]] = parts[1]
	}
	return envMap, nil
}

func readSSHPubkeyFile(filename string) (pubkey ssh.PublicKey, err error) {
	pubkeyString, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	pubkey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(pubkeyString))
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func parseKeyArgs(c *cli.Context) (pubkeys []ssh.PublicKey, err error) {
	// no args, so just use the default
	if len(c.StringSlice("key")) == 0 {
		pubkey, err := readSSHPubkeyFile(filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa.pub"))
		if err != nil {
			return nil, err
		}
		pubkeys = append(pubkeys, pubkey)
		return pubkeys, nil
	}

	for _, arg := range c.StringSlice("key") {
		pubkey, err := readSSHPubkeyFile(arg)
		if err != nil {
			return nil, err
		}
		pubkeys = append(pubkeys, pubkey)
	}
	return pubkeys, nil
}

// parseLifetimeArgs attempts to be a little clever in determining what credential
// lifetime you've chosen. It returns a number of hours and an error. It assumes that
// the argument was passed in as hours.
func parseLifetimeArgs(c *cli.Context) (lifetime int, err error) {
	// the default is zero, which is our default
	if c.Int("lifetime") < 0 {
		return 0, nil
	}

	return c.Int("lifetime"), nil
}

func parseSaveArgs(c *cli.Context) (cred Credential, username, account string, pubkeys []ssh.PublicKey, lifetime int, err error) {
	pubkeys, err = parseKeyArgs(c)
	if err != nil {
		return Credential{}, "", "", nil, 0, err
	}

	username, account, err = parseUserAndAccount(c)
	if err != nil {
		return Credential{}, "", "", nil, 0, err
	}

	envmap, err := parseEnvironmentArgs(c)
	if err != nil {
		return Credential{}, "", "", nil, 0, err
	}

	lifetime, err = parseLifetimeArgs(c)
	if err != nil {
		return Credential{}, "", "", nil, 0, err
	}

	AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if AWSAccessKeyId == "" || AWSSecretAccessKey == "" {
		err := errors.New("Can't save, no credentials in the environment")
		if err != nil {
			return Credential{}, "", "", nil, 0, err
		}
	}
	cred = Credential{
		KeyId:     AWSAccessKeyId,
		SecretKey: AWSSecretAccessKey,
		EnvVars:   envmap,
	}

	return cred, username, account, pubkeys, lifetime, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "credulous"
	app.Usage = "Secure AWS Credential Management"
	app.Version = "0.2.2"

	app.Commands = []cli.Command{
		{
			Name:  "save",
			Usage: "Save AWS credentials",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "key, k",
					Value: &cli.StringSlice{},
					Usage: "\n        SSH public keys for encryption",
				},
				cli.StringSliceFlag{
					Name:  "env, e",
					Value: &cli.StringSlice{},
					Usage: "\n        Environment variables to set in the form VAR=value",
				},
				cli.IntFlag{
					Name:  "lifetime, l",
					Value: 0,
					Usage: "\n        Credential lifetime in seconds (0 means forever)",
				},
				cli.BoolFlag{
					Name: "force, f",
					Usage: "\n        Force saving without validating username or account." +
						"\n        You MUST specify -u username -a account",
				},
				cli.StringFlag{
					Name:  "username, u",
					Value: "",
					Usage: "\n        Username (for use with '--force')",
				},
				cli.StringFlag{
					Name:  "account, a",
					Value: "",
					Usage: "\n        Account alias (for use with '--force')",
				},
			},
			Action: func(c *cli.Context) {
				cred, username, account, pubkeys, lifetime, err := parseSaveArgs(c)
				panic_the_err(err)
				err = SaveCredentials(cred, username, account, pubkeys, lifetime, c.Bool("force"))
				panic_the_err(err)
			},
		},

		{
			Name:  "source",
			Usage: "Source AWS credentials",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "account, a",
					Value: "",
					Usage: "\n        AWS Account alias or id",
				},
				cli.StringFlag{
					Name:  "key, k",
					Value: "",
					Usage: "\n        SSH private key",
				},
				cli.StringFlag{
					Name:  "username, u",
					Value: "",
					Usage: "\n        IAM User",
				},
				cli.StringFlag{
					Name:  "credentials, c",
					Value: "",
					Usage: "\n        Credentials, for example username@account",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "\n        Force sourcing of credentials without validating username or account",
				},
			},
			Action: func(c *cli.Context) {
				keyfile := getPrivateKey(c)
				account, username, err := getAccountAndUserName(c)
				if err != nil {
					panic_the_err(err)
				}
				creds, err := RetrieveCredentials(account, username, keyfile)
				if err != nil {
					panic_the_err(err)
				}

				if !c.Bool("force") {
					err = creds.ValidateCredentials(account, username)
					if err != nil {
						panic_the_err(err)
					}
				}
				creds.Display(os.Stdout)
			},
		},

		{
			Name:  "current",
			Usage: "Show the username and alias of the currently-loaded credentials",
			Action: func(c *cli.Context) {
				AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
				AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				if AWSAccessKeyId == "" || AWSSecretAccessKey == "" {
					err := errors.New("No amazon credentials are currently in your environment")
					panic_the_err(err)
				}
				cred := Credential{
					KeyId:     AWSAccessKeyId,
					SecretKey: AWSSecretAccessKey,
				}
				username, alias, err := getAWSUsernameAndAlias(cred)
				if err != nil {
					panic_the_err(err)
				}
				fmt.Printf("%s@%s\n", username, alias)
			},
		},

		{
			Name:  "display",
			Usage: "Display loaded AWS credentials",
			Action: func(c *cli.Context) {
				AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
				AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				fmt.Printf("AWS_ACCESS_KEY_ID: %s\n", AWSAccessKeyId)
				fmt.Printf("AWS_SECRET_ACCESS_KEY: %s\n", AWSSecretAccessKey)
			},
		},

		{
			Name:  "list",
			Usage: "List available AWS credentials",
			Action: func(c *cli.Context) {
				rootDir, err := os.Open(getRootPath())
				if err != nil {
					panic_the_err(err)
				}
				set, err := listAvailableCredentials(rootDir)
				if err != nil {
					panic_the_err(err)
				}
				for _, cred := range set {
					fmt.Println(cred)
				}
			},
		},

		{
			Name:  "rotate",
			Usage: "Rotate current AWS credentials, deleting the oldest",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "lifetime, l",
					Value: 0,
					Usage: "\n        New credential lifetime in seconds (0 means forever)",
				},
				cli.StringSliceFlag{
					Name:  "key, k",
					Value: &cli.StringSlice{},
					Usage: "\n        SSH public keys for encryption",
				},
				cli.StringSliceFlag{
					Name:  "env, e",
					Value: &cli.StringSlice{},
					Usage: "\n        Environment variables to set in the form VAR=value",
				},
			},
			Action: func(c *cli.Context) {
				cred, _, _, pubkeys, lifetime, err := parseSaveArgs(c)
				panic_the_err(err)
				username, account, err := getAWSUsernameAndAlias(cred)
				panic_the_err(err)
				err = (&cred).rotateCredentials(username)
				panic_the_err(err)
				err = SaveCredentials(cred, username, account, pubkeys, lifetime, c.Bool("force"))
				panic_the_err(err)
			},
		},
	}

	app.Run(os.Args)
}

func rotate(cred Credential) (err error) {
	return nil
}
