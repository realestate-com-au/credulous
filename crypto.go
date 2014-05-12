package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"math/big"
	"reflect"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

// returns a base64 encoded ciphertext. The salt is generated internally
func CredulousEncode(plaintext string, pubkey ssh.PublicKey) (cipher string, salt string, err error) {
	salt = "pepper"
	s := reflect.ValueOf(pubkey).Elem()
	rsaKey := rsa.PublicKey{
		N: s.Field(0).Interface().(*big.Int),
		E: s.Field(1).Interface().(int),
	}
	out, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &rsaKey, []byte(salt+plaintext), []byte("Credulous"))
	panic_the_err(err)
	cipher = base64.StdEncoding.EncodeToString(out)
	return cipher, salt, nil
}

func CredulousDecode(ciphertext string, salt string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, in, []byte("Credulous"))
	panic_the_err(err)
	plaintext = strings.Replace(string(out), salt, "", 1)
	return plaintext, nil
}
