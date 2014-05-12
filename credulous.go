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

func decryptPEM(pemblock *pem.Block) ([]byte, error) {
	passwd, _ := gopass.GetPass("Enter passphrase for ~/.ssh/id_rsa: ")
	decryptedBytes, err := x509.DecryptPEMBlock(pemblock, []byte(passwd))
	panic_the_err(err)
	pemBytes := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: decryptedBytes,
	}
	decryptedPEM := pem.EncodeToMemory(&pemBytes)
	return decryptedPEM, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Credulous"
	app.Usage = "Use it!"
	app.Version = "0.1.2"

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
				var privkeyFile string
				if c.String("key") == "" {
					privkeyFile = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
				} else {
					privkeyFile = c.String("key")
				}
				tmp, err := ioutil.ReadFile(privkeyFile)
				panic_the_err(err)
				pemblock, _ := pem.Decode([]byte(tmp))
				if x509.IsEncryptedPEMBlock(pemblock) {
					tmp, err = decryptPEM(pemblock)
					panic_the_err(err)
				} else {
					log.Print("WARNING: Your private SSH key has no passphrase!")
				}
				key, err := ssh.ParseRawPrivateKey(tmp)
				panic_the_err(err)
				privateKey := key.(*rsa.PrivateKey)
				cred := RetrieveCredentials(c.String("account"), c.String("username"), privateKey)
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
	}

	app.Run(os.Args)
}
