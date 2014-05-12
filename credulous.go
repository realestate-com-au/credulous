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
	decrypted_bytes, err := x509.DecryptPEMBlock(pemblock, []byte(passwd))
	panic_the_err(err)
	pem_bytes := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: decrypted_bytes,
	}
	decrypted_pem := pem.EncodeToMemory(&pem_bytes)
	return decrypted_pem, nil
}

func printErrAndExit(err error) {
	if err != nil {
		fmt.Printf("There was an error %v\n", err)
		os.Exit(1)
	}
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
				var pubkey_file string
				if c.String("key") == "" {
					pubkey_file = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa.pub")
				} else {
					pubkey_file = c.String("key")
				}

				pkFile, err := os.Open(pubkey_file)
				printErrAndExit(err)
				keyId := os.Getenv("AWS_ACCESS_KEY_ID")
				secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				cred, err := NewCredential(keyId, secretKey, &AwsRequestor{})
				printErrAndExit(err)
				fmt.Printf("saving credentials for %s@%s\n", cred.IamUsername, cred.AccountAliasOrId)
				pubKey, err := RetrievePublicKey(pkFile)
				printErrAndExit(err)
				err = EncryptCredential(cred, pubKey)
				printErrAndExit(err)
				SaveCredential(cred, &ConcreteFileOperator{})
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
				var privateKey string
				if c.String("key") == "" {
					privateKey = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa")
				} else {
					privateKey = c.String("key")
				}
				tmp, err := ioutil.ReadFile(privateKey)
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
				privkey := key.(*rsa.PrivateKey)
				cred := RetrieveCredentials(c.String("account"), c.String("username"), privkey)
				cred.Display(os.Stdout)
			},
		},
		{
			Name:  "display",
			Usage: "Display loaded AWS credentials",
			Action: func(c *cli.Context) {
				aws_access_key_id := os.Getenv("AWS_ACCESS_KEY_ID")
				aws_secret_access_key := os.Getenv("AWS_SECRET_ACCESS_KEY")
				fmt.Printf("AWS_ACCESS_KEY_ID: %s\n", aws_access_key_id)
				fmt.Printf("AWS_SECRET_ACCESS_KEY: %s\n", aws_secret_access_key)
			},
		},
	}

	app.Run(os.Args)
}
