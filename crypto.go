package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
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

type AESEncryption struct {
	EncodedKey string
	Ciphertext string
}

func encodeAES(key []byte, plaintext string) (ciphertext string, err error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// We need an unique IV to go at the front of the ciphertext
	out := make([]byte, aes.BlockSize+len(plaintext))
	iv := out[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(cipherBlock, iv)
	stream.XORKeyStream(out[aes.BlockSize:], []byte(plaintext))
	encoded := base64.StdEncoding.EncodeToString(out)
	return encoded, nil
}

// takes a base64-encoded AES-encrypted ciphertext
func decodeAES(key []byte, ciphertext string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	decrypter, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := encrypted[:aes.BlockSize]
	msg := encrypted[aes.BlockSize:]
	aesDecrypter := cipher.NewCFBDecrypter(decrypter, iv)
	aesDecrypter.XORKeyStream(msg, msg)
	return string(msg), nil
}

// returns a base64 encoded ciphertext.
// OAEP can only encrypt plaintexts that are smaller than the key length; for
// a 1024-bit key, about 117 bytes. So instead, this function:
// * generates a random 32-byte symmetric key (randKey)
// * encrypts the plaintext with AES256 using that random symmetric key -> cipherText
// * encrypts the random symmetric key with the ssh PublicKey -> cipherKey
// * returns the base64-encoded marshalled JSON for the ciphertext and key
func CredulousEncode(plaintext string, pubkey ssh.PublicKey) (ciphertext string, err error) {
	rsaKey := sshPubkeyToRsaPubkey(pubkey)
	randKey := make([]byte, 32)
	_, err = rand.Read(randKey)
	if err != nil {
		return "", err
	}

	encoded, err := encodeAES(randKey, plaintext)
	if err != nil {
		return "", err
	}

	out, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &rsaKey, []byte(randKey), []byte("Credulous"))
	if err != nil {
		return "", err
	}
	cipherKey := base64.StdEncoding.EncodeToString(out)

	cipherStruct := AESEncryption{
		EncodedKey: cipherKey,
		Ciphertext: encoded,
	}

	tmp, err := json.Marshal(cipherStruct)
	if err != nil {
		return "", err
	}

	ciphertext = base64.StdEncoding.EncodeToString(tmp)

	return ciphertext, nil
}

func CredulousDecodeAES(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error) {
	in, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// pull apart the layers of base64-encoded JSON
	var encrypted AESEncryption
	err = json.Unmarshal(in, &encrypted)
	if err != nil {
		return "", err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(encrypted.EncodedKey)
	if err != nil {
		return "", err
	}

	// decrypt the AES key using the ssh private key
	aesKey, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privkey, encryptedKey, []byte("Credulous"))
	if err != nil {
		return "", err
	}

	plaintext, err = decodeAES(aesKey, encrypted.Ciphertext)

	return plaintext, nil
}

func CredulousDecodePureRSA(ciphertext string, privkey *rsa.PrivateKey) (plaintext string, err error) {
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
