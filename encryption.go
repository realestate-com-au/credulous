package main

import (
	"io"
	"io/ioutil"

	"crypto/rsa"

	"code.google.com/p/go.crypto/ssh"
)

type SshController interface {
	EncryptCredential(cred *Credential, key ssh.PublicKey) error
	DecryptCredential(cred *Credential, key rsa.PrivateKey) error
}

func RetrievePublicKey(file io.Reader) (ssh.PublicKey, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	pubkey, _, _, _, err := ssh.ParseAuthorizedKey(data)

	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

func EncryptCredential(cred *Credential, key ssh.PublicKey, encryptor SshController) error {
	err := encryptor.EncryptCredential(cred, key)
	if err != nil {
		return err
	}

	return nil
}
