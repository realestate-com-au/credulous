package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

type FileOperator interface {
	ReadFromDisk(path string) (*Credential, error)
	WriteToDisk(cred *Credential, path string) error
}

type ConcreteFileOperator struct {
}

func (c *ConcreteFileOperator) ReadFromDisk(path string) (*Credential, error) {
	return nil, nil
}

func (c *ConcreteFileOperator) WriteToDisk(cred *Credential, path string) error {
	b, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

func SaveCredential(cred *Credential, op FileOperator) error {
	path := buildPath(cred.AccountAliasOrId, cred.IamUsername)
	keyTime, err := time.Parse("2006-01-02T15:04:05Z", cred.CreateTime)
	if err != nil {
		return err
	}
	filename := buildCredentialFileName(keyTime, cred.KeyId)
	err = op.WriteToDisk(cred, filepath.Join(path, filename))
	if err != nil {
		return err
	}
	return nil
}

func RetrieveCredential(account, username string, op FileOperator) (*Credential, error) {
	path := buildPath(account, username)
	cred, err := op.ReadFromDisk(path)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func buildPath(account, username string) string {
	home := os.Getenv("HOME")
	rootPath := filepath.Join(home, ".credulous", "local")
	return filepath.Join(rootPath, account, username)
}

func buildCredentialFileName(keyTime time.Time, apiKey string) string {
	return fmt.Sprintf("%v_%v.json", keyTime.Unix(), apiKey[12:])
}

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

func NewCredential(keyId, secret string, aws AwsCaller) (*Credential, error) {
	cred := &Credential{
		KeyId:     keyId,
		SecretKey: secret,
	}

	username, err := aws.GetAwsUsername(keyId, secret)
	if err != nil {
		return nil, err
	}
	cred.IamUsername = username

	account, err := aws.GetAwsAccountAlias(keyId, secret)
	if err != nil {
		return nil, err
	}
	cred.AccountAliasOrId = account

	createDate, err := aws.GetKeyCreateDate(keyId, secret)
	if err != nil {
		return nil, err
	}
	cred.CreateTime = createDate

	return cred, nil
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

func RetrieveCredentials(alias string, username string, privkey *rsa.PrivateKey) Credential {
	fullPath := filepath.Join(getRootPath(), "local", alias, username)
	filePath := filepath.Join(fullPath, latestFileInDir(fullPath).Name())
	return *readCredentialFile(filePath, privkey)
}

func latestFileInDir(dir string) os.FileInfo {
	entries, err := ioutil.ReadDir(dir)
	panic_the_err(err)
	return entries[len(entries)-1]
}
