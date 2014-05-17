package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"crypto/rsa"
	"crypto/x509"

	"code.google.com/p/go.crypto/ssh"

	"code.google.com/p/gopass"
	"github.com/codegangsta/cli"
)

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

func getPrivateKey(c *cli.Context) *rsa.PrivateKey {
	var privkeyFile string
	var tmp []byte
	var err error

	if c.String("key") == "" {
		privkeyFile = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
	} else {
		privkeyFile = c.String("key")
	}

	if tmp, err = ioutil.ReadFile(privkeyFile); err != nil {
		panic_the_err(err)
	}

	pemblock, _ := pem.Decode([]byte(tmp))
	if x509.IsEncryptedPEMBlock(pemblock) {
		if tmp, err = decryptPEM(pemblock, privkeyFile); err != nil {
			panic_the_err(err)
		}
	} else {
		log.Print("WARNING: Your private SSH key has no passphrase!")
	}

	key, err := ssh.ParseRawPrivateKey(tmp)
	if err != nil {
		panic_the_err(err)
	}
	privateKey := key.(*rsa.PrivateKey)
	return privateKey
}

func main() {
	app := cli.NewApp()
	app.Name = "credulous"
	app.Usage = "Use it!"
	app.Version = "0.1.3"

	app.Commands = []cli.Command{
		{
			Name:  "save",
			Usage: "Save AWS credentials for a file.",
			Flags: []cli.Flag{
				cli.StringFlag{"key, k", "", "SSH public key"},
			},
			Action: func(c *cli.Context) {
				var pubkeyFile string
				if c.String("key") == "" {
					pubkeyFile = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa.pub")
				} else {
					pubkeyFile = c.String("key")
				}

				AWSAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
				AWSSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				if AWSAccessKeyId == "" || AWSSecretAccessKey == "" {
					fmt.Println("Can't save, no credentials in the environment")
					os.Exit(1)
				}
				username, _ := getAWSUsername(AWSAccessKeyId, AWSSecretAccessKey)
				alias, _ := getAWSAccountAlias(AWSAccessKeyId, AWSSecretAccessKey)
				fmt.Printf("saving credentials for %s@%s\n", username, alias)
				pubkeyString, err := ioutil.ReadFile(pubkeyFile)
				panic_the_err(err)
				pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkeyString))
				SaveCredentials(username, alias, AWSAccessKeyId, AWSSecretAccessKey, pubkey)
			},
		},
		{
			Name:  "source",
			Usage: "Source AWS credentials from a file.",
			Flags: []cli.Flag{
				cli.StringFlag{"account, a", "", "AWS Account alias or id"},
				cli.StringFlag{"key, k", "", "SSH private key"},
				cli.StringFlag{"username, u", "", "IAM User"},
			},
			Action: func(c *cli.Context) {
				
				privateKey := getPrivateKey(c) 
				cred, err := RetrieveCredentials(c.String("account"), c.String("username"), privateKey)
				if err != nil {
					panic_the_err(err)
				}
				err = ValidateCredentials(c.String("account"), c.String("username"), cred)
				if err != nil {
					panic_the_err(err)
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
			Usage: "List the sets of available credentials",
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
