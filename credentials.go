package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/realestate-com-au/goamz/aws"
	"github.com/realestate-com-au/goamz/iam"

	"code.google.com/p/go.crypto/ssh"
)

const FORMAT_VERSION string = "2014-06-12"

// How long to retry after rotating credentials for
// new credentials to become active (in seconds)
const ROTATE_TIMEOUT int = 30

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

func parseCredential(data []byte, keyfile string) (*Credentials, error) {
	var creds Credentials
	err := json.Unmarshal(data, &creds)
	if err != nil {
		return nil, err
	}

	privKey, err := loadPrivateKey(keyfile)
	if err != nil {
		return nil, err
	}

	fp, err := SSHPrivateFingerprint(*privKey)
	if err != nil {
		return nil, err
	}

	var offset int = -1
	for i, enc := range creds.Encryptions {
		if enc.Fingerprint == fp {
			offset = i
			break
		}
	}

	if offset < 0 {
		err := errors.New("The SSH key specified cannot decrypt those credentials")
		return nil, err
	}

	var tmp string
	switch {
	case creds.Version == "2014-05-31":
		log.Print("INFO: These credentials are in the old format; re-run 'credulous save' now to remove this warning")
		tmp, err = CredulousDecodePureRSA(creds.Encryptions[offset].Ciphertext, privKey)
	case creds.Version == "2014-06-12":
		tmp, err = CredulousDecodeAES(creds.Encryptions[offset].Ciphertext, privKey)
	}

	if err != nil {
		return nil, err
	}

	var cred Credential
	err = json.Unmarshal([]byte(tmp), &cred)
	if err != nil {
		return nil, err
	}

	creds.Encryptions[0].decoded = cred
	return &creds, nil
}

func readCredentialFile(fileName string, keyfile string) (*Credentials, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(b), "Version") {
		log.Print("INFO: These credentials are in the old format; re-run 'credulous save' now to remove this warning")
		creds, err := parseOldCredential(b, keyfile)
		if err != nil {
			return nil, err
		}
		return creds, nil
	}

	creds, err := parseCredential(b, keyfile)
	if err != nil {
		return nil, err
	}

	return creds, nil
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
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=\"%v\"\nexport AWS_SECRET_ACCESS_KEY=\"%v\"\n", cred.KeyId, cred.SecretKey)
}

func (cred Credentials) Display(output io.Writer) {
	fmt.Fprintf(output, "export AWS_ACCESS_KEY_ID=\"%v\"\nexport AWS_SECRET_ACCESS_KEY=\"%v\"\n",
		cred.Encryptions[0].decoded.KeyId, cred.Encryptions[0].decoded.SecretKey)
	for key, val := range cred.Encryptions[0].decoded.EnvVars {
		fmt.Fprintf(output, "export %s=\"%s\"\n", key, val)
	}
}

func (creds Credentials) verifyUserAndAccount() error {
	// need to check both the username and the account alias for the
	// supplied creds match the passed-in username and account alias
	auth := aws.Auth{
		AccessKey: creds.Encryptions[0].decoded.KeyId,
		SecretKey: creds.Encryptions[0].decoded.SecretKey,
	}
	// Note: the region is irrelevant for IAM
	instance := iam.New(auth, aws.APSoutheast2)

	// Make sure the account is who we expect
	err := verify_account(creds.AccountAliasOrId, instance)
	if err != nil {
		return err
	}

	// Make sure the user is who we expect
	// If the username is the same as the account name, then it's the root user
	// and there's actually no username at all (oddly)
	if creds.IamUsername == creds.AccountAliasOrId {
		err = verify_user("", instance)
	} else {
		err = verify_user(creds.IamUsername, instance)
	}
	if err != nil {
		return err
	}

	return nil
}

// Only delete the oldest key *if* the new key is valid; otherwise,
// delete the newest key
func (cred *Credential) deleteOneKey(username string) (err error) {
	auth := aws.Auth{
		AccessKey: cred.KeyId,
		SecretKey: cred.SecretKey,
	}
	instance := iam.New(auth, aws.APSoutheast2)

	allKeys, err := instance.AccessKeys(username)
	if err != nil {
		return err
	}

	// wtf?
	if len(allKeys.AccessKeys) == 0 {
		err = errors.New("Zero access keys found for this account -- cannot rotate")
		return err
	}

	// only one key
	if len(allKeys.AccessKeys) == 1 {
		return nil
	}

	// Find out which key to delete.
	var oldestId string
	var oldest int64

	for _, key := range allKeys.AccessKeys {
		t, err := time.Parse("2006-01-02T15:04:05Z", key.CreateDate)
		key_create_date := t.Unix()
		if err != nil {
			return err
		}
		// If we find an inactive one, just delete it
		if key.Status == "Inactive" {
			oldestId = key.Id
			break
		}
		if oldest == 0 || key_create_date < oldest {
			oldest = key_create_date
			oldestId = key.Id
		}
	}

	if oldestId == "" {
		err = errors.New("Cannot find oldest key for this account, will not rotate")
		return err
	}

	_, err = instance.DeleteAccessKey(oldestId, username)
	if err != nil {
		return err
	}

	return nil
}

func (cred *Credential) createNewAccessKey(username string) (err error) {
	auth := aws.Auth{
		AccessKey: cred.KeyId,
		SecretKey: cred.SecretKey,
	}
	instance := iam.New(auth, aws.APSoutheast2)

	resp, err := instance.CreateAccessKey(username)
	if err != nil {
		return err
	}

	cred.KeyId = resp.AccessKey.Id
	cred.SecretKey = resp.AccessKey.Secret
	return nil
}

// Potential conditions to handle here:
// * AWS has one key
//     * only generate a new key, do not delete the old one
// * AWS has two keys
//     * both are active and valid
//     * new one is inactive
//     * old one is inactive
// * We successfully delete the oldest key, but fail in creating the new key (eg network, permission issues)
func (cred *Credential) rotateCredentials(username string) (err error) {
	err = cred.deleteOneKey(username)
	if err != nil {
		return err
	}
	err = cred.createNewAccessKey(username)
	if err != nil {
		return err
	}
	// Loop until the credentials are active
	count := 0
	for _, _, err = getAWSUsernameAndAlias(*cred); err != nil && count < ROTATE_TIMEOUT; _, _, err = getAWSUsernameAndAlias(*cred) {
		time.Sleep(1 * time.Second)
		count += 1
	}
	if err != nil {
		err = errors.New("Timed out waiting for new credentials to become active")
		return err
	}
	return nil
}

func SaveCredentials(cred Credential, username, alias string, pubkeys []ssh.PublicKey, lifetime int, force bool) (err error) {

	var key_create_date int64

	if force {
		key_create_date = time.Now().Unix()
	} else {
		auth := aws.Auth{AccessKey: cred.KeyId, SecretKey: cred.SecretKey}
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

		date, _ := getKeyCreateDate(instance)
		t, err := time.Parse("2006-01-02T15:04:05Z", date)
		key_create_date = t.Unix()
		if err != nil {
			return err
		}
	}

	fmt.Printf("saving credentials for %s@%s\n", username, alias)
	plaintext, err := json.Marshal(cred)
	if err != nil {
		return err
	}

	enc_slice := []Encryption{}
	for _, pubkey := range pubkeys {
		encoded, err := CredulousEncode(string(plaintext), pubkey)
		if err != nil {
			return err
		}

		enc_slice = append(enc_slice, Encryption{
			Ciphertext:  encoded,
			Fingerprint: SSHFingerprint(pubkey),
		})
	}
	creds := Credentials{
		Version:          FORMAT_VERSION,
		AccountAliasOrId: alias,
		IamUsername:      username,
		CreateTime:       fmt.Sprintf("%d", key_create_date),
		Encryptions:      enc_slice,
		LifeTime:         lifetime,
	}

	creds.WriteToDisk(fmt.Sprintf("%v-%v.json", key_create_date, cred.KeyId[12:]))
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
