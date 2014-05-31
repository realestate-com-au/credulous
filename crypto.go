package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
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

// returns a base64 encoded ciphertext. The salt is generated internally
func CredulousEncode(plaintext string, pubkey ssh.PublicKey) (cipher string, err error) {
	s := reflect.ValueOf(pubkey).Elem()
	rsaKey := rsa.PublicKey{
		N: s.Field(0).Interface().(*big.Int),
		E: s.Field(1).Interface().(int),
	}
	out, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &rsaKey, []byte(plaintext), []byte("Credulous"))
	if err != nil {
		return "", err
	}
	cipher = base64.StdEncoding.EncodeToString(out)
	return cipher, nil
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
