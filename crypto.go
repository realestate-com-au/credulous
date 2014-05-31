package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"reflect"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

type Salter interface {
	GenerateSalt() (string, error)
}

type RandomSaltGenerator struct {
}

type StaticSaltGenerator struct {
	salt string
}

func (ssg *StaticSaltGenerator) GenerateSalt() (string, error) {
	return ssg.salt, nil
}

func (sg *RandomSaltGenerator) GenerateSalt() (string, error) {
	const SALT_LENGTH = 8
	b := make([]byte, SALT_LENGTH)
	_, err := rand.Read(b)
	encoder := base64.StdEncoding
	encoded := make([]byte, encoder.EncodedLen(len(b)))
	encoder.Encode(encoded, b)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func sshPubkeyToRsaPubkey(pubkey ssh.PublicKey) rsa.PublicKey {
	s := reflect.ValueOf(pubkey).Elem()
	rsaKey := rsa.PublicKey{
		N: s.Field(0).Interface().(*big.Int),
		E: s.Field(1).Interface().(int),
	}
	return rsaKey
}

func rsaPubkeyToSSHPubkey(rsakey rsa.PublicKey) (sshkey ssh.PublicKey, err error) {
	sshkey, err = ssh.NewPublicKey(&rsakey)
	if err != nil {
		return nil, err
	}
	return sshkey, nil
}

// returns a base64 encoded ciphertext. The salt is generated internally
func CredulousEncode(plaintext string, pubkey ssh.PublicKey) (cipher string, err error) {
	rsaKey := sshPubkeyToRsaPubkey(pubkey)
	out, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &rsaKey, []byte(plaintext), []byte("Credulous"))
	if err != nil {
		return "", err
	}
	cipher = base64.StdEncoding.EncodeToString(out)
	return cipher, nil
}

func CredulousDecode(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, in, []byte("Credulous"))
	if err != nil {
		return "", err
	}
	plaintext = string(out)
	return plaintext, nil
}

func CredulousDecodeWithSalt(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, in, []byte("Credulous"))
	if err != nil {
		return "", err
	}
	plaintext = strings.Replace(string(out), salt, "", 1)
	return plaintext, nil
}

func loadPrivateKey(filename string) (privateKey *rsa.PrivateKey, err error) {
	var tmp []byte

	if tmp, err = ioutil.ReadFile(filename); err != nil {
		return &rsa.PrivateKey{}, err
	}

	pemblock, _ := pem.Decode([]byte(tmp))
	if x509.IsEncryptedPEMBlock(pemblock) {
		if tmp, err = decryptPEM(pemblock, filename); err != nil {
			return &rsa.PrivateKey{}, err
		}
	} else {
		log.Print("WARNING: Your private SSH key has no passphrase!")
	}

	key, err := ssh.ParseRawPrivateKey(tmp)
	if err != nil {
		return &rsa.PrivateKey{}, err
	}
	privateKey = key.(*rsa.PrivateKey)
	return privateKey, nil
}

func SSHFingerprint(pubkey ssh.PublicKey) (fingerprint string) {
	binary := pubkey.Marshal()
	hash := md5.Sum(binary)
	// now add the colons
	fingerprint = fmt.Sprintf("%02x", (hash[0]))
	for i := 1; i < len(hash); i += 1 {
		fingerprint += ":" + fmt.Sprintf("%02x", (hash[i]))
	}
	return fingerprint
}

func SSHPrivateFingerprint(privkey rsa.PrivateKey) (fingerprint string, err error) {
	sshPubkey, err := rsaPubkeyToSSHPubkey(privkey.PublicKey)
	if err != nil {
		return "", err
	}
	fingerprint = SSHFingerprint(sshPubkey)
	return fingerprint, nil
}
