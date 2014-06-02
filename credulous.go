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
	if passwd, err = gopass.GetPass(""); err != nil {
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

func getAccountAndUserName(c *cli.Context) (string, string, error) {
	if len(c.Args()) > 0 {
		result := strings.Split(c.Args()[0], "@")
		if len(result) < 2 {
			err := errors.New("Invalid account format; please specify <username>@<account>")
			return "", "", err
		}
		return result[1], result[0], nil
	}
	if c.String("credentials") != "" {
		result := strings.Split(c.String("credentials"), "@")
		if len(result) < 2 {
			err := errors.New("Invalid account format; please specify <username>@<account>")
			return "", "", err
		}
		return result[1], result[0], nil
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
	if c.StringSlice("env") == nil {
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
		parts := strings.Split(arg, "=")
		envMap[parts[0]] = parts[1]
	}
	return envMap, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "credulous"
	app.Usage = "Secure AWS Credential Management"
	app.Version = "0.1.3"

	app.Commands = []cli.Command{
		{
			Name:  "save",
			Usage: "Save AWS credentials",
			Flags: []cli.Flag{
				cli.StringFlag{"key, k", "", "SSH public key"},
				cli.StringSliceFlag{"env, e", &cli.StringSlice{}, "Environment variables to set in the form VAR=value"},
				cli.BoolFlag{"force, f", "Force saving without validating username or account.\n" +
					"\tYou MUST specify -u username -a account"},
				cli.StringFlag{"username, u", "", "Username (for use with '--force')"},
				cli.StringFlag{"account, a", "", "Account alias (for use with '--force')"},
			},
			Action: func(c *cli.Context) {
				var pubkeyFile string
				if c.String("key") == "" {
					pubkeyFile = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa.pub")
				} else {
					pubkeyFile = c.String("key")
				}

				username, account, err := parseUserAndAccount(c)
				if err != nil {
					panic_the_err(err)
				}

				envmap, err := parseEnvironmentArgs(c)
				if err != nil {
					panic_the_err(err)
				}

				AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
				AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				if AWSAccessKeyId == "" || AWSSecretAccessKey == "" {
					err := errors.New("Can't save, no credentials in the environment")
					panic_the_err(err)
				}
				pubkeyString, err := ioutil.ReadFile(pubkeyFile)
				panic_the_err(err)
				pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkeyString))
				cred := Credential{
					KeyId:     AWSAccessKeyId,
					SecretKey: AWSSecretAccessKey,
					EnvVars:   envmap,
				}
				err = SaveCredentials(cred, username, account, pubkey, c.Bool("force"))
				panic_the_err(err)
			},
		},
		{
			Name:  "source",
			Usage: "Source AWS credentials",
			Flags: []cli.Flag{
				cli.StringFlag{"account, a", "", "AWS Account alias or id"},
				cli.StringFlag{"key, k", "", "SSH private key"},
				cli.StringFlag{"username, u", "", "IAM User"},
				cli.StringFlag{"credentials, c", "", "Credentials, for example username@account"},
				cli.BoolFlag{"force, f", "Force sourcing of credentials without validating username or account"},
			},
			Action: func(c *cli.Context) {
				keyfile := getPrivateKey(c)
				account, username, err := getAccountAndUserName(c)
				if err != nil {
					panic_the_err(err)
				}
				cred, err := RetrieveCredentials(account, username, keyfile)
				if err != nil {
					panic_the_err(err)
				}

				if !c.Bool("force") {
					err = cred.ValidateCredentials(account, username)
					if err != nil {
						panic_the_err(err)
					}
				}
				cred.Display(os.Stdout)
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
	}

	app.Run(os.Args)
}
