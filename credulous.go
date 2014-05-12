package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"crypto/rsa"

	"code.google.com/p/go.crypto/ssh"

	"github.com/codegangsta/cli"
)

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
					pubkey_file = filepath.Join(os.Getenv("HOME"), "/.ssh/id_rsa_nopasswd.pub")
				} else {
					pubkey_file = c.String("key")
				}

				aws_access_key_id := os.Getenv("AWS_ACCESS_KEY_ID")
				aws_secret_access_key := os.Getenv("AWS_SECRET_ACCESS_KEY")
				username, _ := getAWSUsername(aws_access_key_id, aws_secret_access_key)
				alias, _ := getAWSAccountAlias(aws_access_key_id, aws_secret_access_key)
				fmt.Printf("saving credentials for %s@%s\n", username, alias)
				pubkey_str, err := ioutil.ReadFile(pubkey_file)
				panic_the_err(err)
				pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkey_str))
				SaveCredentials(username, alias, aws_access_key_id, aws_secret_access_key, pubkey)
			},
		},
		{
			Name:  "source",
			Usage: "Source AWS credentials from a file.",
			Flags: []cli.Flag{
				cli.StringFlag{"account, a", "", "AWS Account alias or id"},
				cli.StringFlag{"username, u", "", "IAM User"},
			},
			Action: func(c *cli.Context) {

				home := os.Getenv("HOME")
				tmp, err := ioutil.ReadFile(home + "/.ssh/id_rsa_nopasswd")
				panic_the_err(err)
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
