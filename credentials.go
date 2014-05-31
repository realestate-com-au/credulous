package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/realestate-com-au/goamz/aws"
	"github.com/realestate-com-au/goamz/iam"

	"code.google.com/p/go.crypto/ssh"
)

const FORMAT_VERSION string = "2014-05-31"

type Credentials struct {
	Version          string
	IamUsername      string
	AccountAliasOrId string
	CreateTime       string
	LifeTime         int
	Encryptions      []Encryption
}

type Encryption struct {
	Fingerprint string
	Ciphertext  string
	// we can do this because the field isn't exported
	// and so won't be included when we call Marshal to
	// save the encrypted credentials
	decoded Credential
}

type Credential struct {
	KeyId     string
	SecretKey string
	EnvVars   map[string]string
}

type OldCredential struct {
	CreateTime       string
	LifeTime         int
	KeyId            string
	SecretKey        string
	Salt             string
	AccountAliasOrId string
	IamUsername      string
	FingerPrint      string
}

func decodeOldCredential(data []byte, keyfile string) (*OldCredential, error) {
	var credential OldCredential
	err := json.Unmarshal(data, &credential)
	if err != nil {
		return nil, err
	}

	privKey, err := loadPrivateKey(keyfile)
	if err != nil {
		return nil, err
	}

	decoded, err := CredulousDecodeWithSalt(credential.KeyId, credential.Salt, privKey)
	if err != nil {
		return nil, err
	}
	credential.KeyId = decoded

	decoded, err = CredulousDecodeWithSalt(credential.SecretKey, credential.Salt, privKey)
	if err != nil {
		return nil, err
	}
	credential.SecretKey = decoded

	return &credential, nil
}

func parseOldCredential(data []byte, keyfile string) (*Credentials, error) {
	oldCred, err := decodeOldCredential(data, keyfile)
	if err != nil {
		return nil, err
	}
	// build a new Credentials structure out of the old
	cred := Credential{
		KeyId:     oldCred.KeyId,
		SecretKey: oldCred.SecretKey,
	}
	enc := []Encryption{}
	enc = append(enc, Encryption{
		decoded: cred,
	})
	creds := Credentials{
		Version:          "noversion",
		IamUsername:      oldCred.IamUsername,
		AccountAliasOrId: oldCred.AccountAliasOrId,
		CreateTime:       oldCred.CreateTime,
		LifeTime:         oldCred.LifeTime,
		Encryptions:      enc,
	}

	return &creds, nil
}

func readCredentialFile(fileName string, keyfile string) (*Credentials, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(b), "Version:") {
		creds, err := parseOldCredential(b, keyfile)
		if err != nil {
			return nil, err
		}
		return creds, nil
	}

	return &Credentials{}, nil
}

func (cred Credentials) WriteToDisk(filename string) (err error) {
	b, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	path := filepath.Join(getRootPath(), "local", cred.AccountAliasOrId, cred.IamUsername)
	os.MkdirAll(path, 0700)
	err = ioutil.WriteFile(filepath.Join(path, filename), b, 0600)
	return err
}

func (cred OldCredential) Display(output io.Writer) {
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\n", cred.KeyId, cred.SecretKey)
}

func (cred Credentials) Display(output io.Writer) {
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\n",
		cred.Encryptions[0].decoded.KeyId, cred.Encryptions[0].decoded.SecretKey)
}

func SaveCredentials(id, secret, username, alias string, pubkey ssh.PublicKey) (err error) {
	auth := aws.Auth{AccessKey: id, SecretKey: secret}
	instance := iam.New(auth, aws.APSoutheast2)
	if username == "" {
		username, err = getAWSUsername(instance)
		if err != nil {
			return err
		}
	}
	if alias == "" {
		alias, err = getAWSAccountAlias(instance)
		if err != nil {
			return err
		}
	}
	fmt.Printf("saving credentials for %s@%s\n", username, alias)
	secrets := Credential{
		KeyId:     id,
		SecretKey: secret,
	}
	plaintext, err := json.Marshal(secrets)
	if err != nil {
		return err
	}
	encoded, err := CredulousEncode(string(plaintext), pubkey)
	if err != nil {
		return err
	}
	key_create_date, _ := getKeyCreateDate(instance)
	t, err := time.Parse("2006-01-02T15:04:05Z", key_create_date)
	if err != nil {
		return err
	}
	enc_slice := []Encryption{}
	enc_slice = append(enc_slice, Encryption{
		Ciphertext:  encoded,
		Fingerprint: SSHFingerprint(pubkey),
	})
	creds := Credentials{
		Version:          FORMAT_VERSION,
		AccountAliasOrId: alias,
		IamUsername:      username,
		CreateTime:       key_create_date,
		Encryptions:      enc_slice,
	}

	creds.WriteToDisk(fmt.Sprintf("%v-%v.json", t.Unix(), id[12:]))
	return nil
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

func getDirs(fl FileLister) ([]os.FileInfo, error) {
	dirents, err := fl.Readdir(0) // get all the entries
	if err != nil {
		return nil, err
	}

	dirs := []os.FileInfo{}
	for _, dirent := range dirents {
		if dirent.IsDir() {
			dirs = append(dirs, dirent)
		}
	}

	return dirs, nil
}

func findDefaultDir(fl FileLister) (string, error) {
	dirs, err := getDirs(fl)
	if err != nil {
		return "", err
	}

	switch {
	case len(dirs) == 0:
		return "", errors.New("No saved credentials found; please run 'credulous save' first")
	case len(dirs) > 1:
		return "", errors.New("More than one account found; please specify account and user")
	}

	return dirs[0].Name(), nil
}

func (cred Credentials) ValidateCredentials(alias string, username string) error {
	if cred.IamUsername != username {
		err := errors.New("FATAL: username in credential does not match requested username")
		return err
	}
	if cred.AccountAliasOrId != alias {
		err := errors.New("FATAL: account alias in credential does not match requested alias")
		return err
	}

	err := cred.verifyUserAndAccount()
	if err != nil {
		return err
	}
	return nil
}

func RetrieveCredentials(alias string, username string, keyfile string) (Credentials, error) {
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
	cred, err := readCredentialFile(filePath, keyfile)
	if err != nil {
		return Credentials{}, err
	}

	return *cred, nil
}

func latestFileInDir(dir string) os.FileInfo {
	entries, err := ioutil.ReadDir(dir)
	panic_the_err(err)
	return entries[len(entries)-1]
}

func listAvailableCredentials(rootDir FileLister) ([]string, error) {
	var creds []string

	repo_dirs, err := getDirs(rootDir) // get just the directories
	if err != nil {
		return creds, err
	}

	if len(repo_dirs) == 0 {
		return creds, errors.New("No saved credentials found; please run 'credulous save' first")
	}

	for _, repo_dirent := range repo_dirs {
		repo_path := filepath.Join(getRootPath(), repo_dirent.Name())
		repo_dir, err := os.Open(repo_path)
		if err != nil {
			return creds, err
		}

		alias_dirs, err := getDirs(repo_dir)
		if err != nil {
			return creds, err
		}

		for _, alias_dirent := range alias_dirs {
			alias_path := filepath.Join(repo_path, alias_dirent.Name())
			alias_dir, err := os.Open(alias_path)
			if err != nil {
				return creds, err
			}

			user_dirs, err := getDirs(alias_dir)
			if err != nil {
				return creds, err
			}

			for _, user_dirent := range user_dirs {
				user_path := filepath.Join(alias_path, user_dirent.Name())
				if latest := latestFileInDir(user_path); latest.Name() != "" {
					creds = append(creds, user_dirent.Name()+"@"+alias_dirent.Name())
				}
			}
		}
	}
	return creds, nil
}
