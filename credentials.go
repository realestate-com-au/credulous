package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

type Credential struct {
	CreateTime       string
	LifeTime         int
	KeyId            string
	SecretKey        string
	Salt             string
	AccountAliasOrId string
	IamUsername      string
	FingerPrint      string
}

func readCredentialFile(fileName string, privkey *rsa.PrivateKey) *Credential {
	b, err := ioutil.ReadFile(fileName)
	panic_the_err(err)

	var credential Credential
	err = json.Unmarshal(b, &credential)
	panic_the_err(err)

	decoded, err := CredulousDecode(credential.KeyId, credential.Salt, privkey)
	panic_the_err(err)
	credential.KeyId = decoded

	decoded, err = CredulousDecode(credential.SecretKey, credential.Salt, privkey)
	panic_the_err(err)
	credential.SecretKey = decoded

	return &credential
}

func (cred Credential) WriteToDisk(filename string) {
	b, err := json.Marshal(cred)
	panic_the_err(err)
	path := filepath.Join(getRootPath(), "local", cred.AccountAliasOrId, cred.IamUsername)
	os.MkdirAll(path, 0700)
	err = ioutil.WriteFile(filepath.Join(path, filename), b, 0600)
	panic_the_err(err)
}

func (cred Credential) Display(output io.Writer) {
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\n", cred.KeyId, cred.SecretKey)
}

func SaveCredentials(username, alias, id, secret string, pubkey ssh.PublicKey) {
	random_salt := RandomSaltGenerator{}
	id_encoded, generated_salt, err := CredulousEncode(id, &random_salt, pubkey)
	static_salt := StaticSaltGenerator{salt: generated_salt}
	panic_the_err(err)
	secret_encoded, generated_salt, err := CredulousEncode(secret, &static_salt, pubkey)
	panic_the_err(err)
	creds := Credential{
		KeyId:            id_encoded,
		SecretKey:        secret_encoded,
		AccountAliasOrId: alias,
		IamUsername:      username,
		Salt:             generated_salt,
	}
	key_create_date, _ := getKeyCreateDate(id, secret)
	t, err := time.Parse("2006-01-02T15:04:05Z", key_create_date)
	panic_the_err(err)
	creds.WriteToDisk(fmt.Sprintf("%v-%v.json", t.Unix(), id[12:]))
}

func getRootPath() string {
	home := os.Getenv("HOME")
	rootPath := home + "/.credulous"
	os.MkdirAll(rootPath, 0700)
	return rootPath
}

type FileLister interface {
	Readdir(int) ([]os.FileInfo, error)
}

func findDefaultDir(fl FileLister) (string, error) {
	dirents, err := fl.Readdir(0)
	if err != nil {
		return "", err
	}

	dirs := []os.FileInfo{}
	for _, dirent := range dirents {
		if dirent.IsDir() {
			dirs = append(dirs, dirent)
		}
	}

	switch {
	case len(dirs) == 0:
		return "", errors.New("No saved credentials found; please run 'credulous save' first")
	case len(dirs) > 1:
		return "", errors.New("More than one account found; please specify account and user")
	}

	return dirs[0].Name(), nil
}

func RetrieveCredentials(alias string, username string, privkey *rsa.PrivateKey) Credential {
	rootPath := filepath.Join(getRootPath(), "local")
	rootDir, err := os.Open(rootPath)
	if err != nil {
		panic_the_err(err)
	}

	if alias == "" {
		if alias, err = findDefaultDir(rootDir); err != nil {
			panic_the_err(err)
		}
	}

	if username == "" {
		aliasDir, err := os.Open(filepath.Join(rootPath, alias))
		if err != nil {
			panic_the_err(err)
		}
		username, err = findDefaultDir(aliasDir)
		if err != nil {
			panic_the_err(err)
		}
	}

	fullPath := filepath.Join(rootPath, alias, username)
	filePath := filepath.Join(fullPath, latestFileInDir(fullPath).Name())
	return *readCredentialFile(filePath, privkey)
}

func latestFileInDir(dir string) os.FileInfo {
	entries, err := ioutil.ReadDir(dir)
	panic_the_err(err)
	return entries[len(entries)-1]
}
