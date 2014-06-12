package main

import (
	"io/ioutil"
	"testing"

	"crypto/rsa"

	"code.google.com/p/go.crypto/ssh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEncode(t *testing.T) {
	Convey("Test Encode A String", t, func() {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXg9Vmhy9YSB8BcN3yHgQjdX9lN3j2KRpv7kVDXSiIana2WbKP7IiTS0uJcJWUM3vlHjdL9KOO0jCWWzVFIcmLhiVVG+Fy2tothBp/NhjR8WWG/6Jg/6tXvVkLG6bDgfbDaLWdE5xzjL0YG8TrIluqnu0J5GHKrQcXF650PlqkGo+whpXrS8wOG+eUmsHX9L1w/Z3TkQlMjQNJEoRbqqSrp7yGj4JqzbtLpsglPRlobD7LHp+5ZDxzpk9i+6hoMxp2muDFxnEtZyED6IMQlNNEGkc3sdmGPOo26oW2+ePkBcjpOpdVif/Iya/jDLuLFHAOol6G34Tr4IdTgaL0qCCr TEST KEY"))
		panic_the_err(err)
		plaintext := "some plaintext"
		ciphertext, err := CredulousEncode(plaintext, pubkey)
		So(err, ShouldEqual, nil)
		So(len(ciphertext), ShouldEqual, 556)
	})
}

func TestEncodeAES(t *testing.T) {
	Convey("Test encoding with AES", t, func() {
		plaintext := "some plaintext"
		key := "12345678901234567890123456789012"
		ciphertext, err := encodeAES([]byte(key), plaintext)
		So(err, ShouldEqual, nil)
		So(len(ciphertext), ShouldEqual, 40)
	})
}

func TestDecodeAES(t *testing.T) {
	Convey("Test encoding with AES", t, func() {
		ciphertext := "ghRydcBg7LR66v8pF6PvaXZ67gHk8toOtveDE+dP"
		key := "12345678901234567890123456789012"
		plaintext, err := decodeAES([]byte(key), ciphertext)
		So(err, ShouldEqual, nil)
		So(plaintext, ShouldEqual, "some plaintext")
	})
}

func TestDecodeAESCredential(t *testing.T) {
	Convey("Test decoding an AES-encrypted ciphertext", t, func() {
		ciphertext := "eyJFbmNvZGVkS2V5IjoicDI5R3NmSmhIVjYvRGd3cmd1d040aDhKTmErTGJkZ0VHcU5vaVB6c1Rnb3IrOEJsQnJTVW1rWGZQTlFvRnY4NHdlcGkvYmd4ZmNyYlpDWm5iMEx4bW9pVjhjMERZYlE5M3F1d0ptK2VBNVhSVlZzTFZodUk1RG9rOENMbkwxOEl5aXc4OENWMXR6ZkJOUWNnQVdBckpsNHBMdzZEbkVFS21NOHRabCtNRUVnTlFjVStybUprKytZbU1ubW44KzVEU1Q5TWtLQ0lxeHl2eVNCRGYxVGkrS2ZHNTlXajkybGQycGZ1Q3k5YWREYlQ2azc0ZG1MbFkvOTlZMWVDZkREMmJWZjNueWJrUkg2UTM3bXNQVHpnbGRaWE56cjBoeStTUERTZHozU0lBSmZGZGw1dy9ka3pYTms2TXcwaHMxbjhRR1BsdnBMOFI1MzF1Rit5a3c5STh3PT0iLCJDaXBoZXJ0ZXh0Ijoiem5Cc2ZxbmJwYTFtdEF6Q09GMVZpU3VsUlRQSGZIblE1UEREZzluYyJ9"
		tmp, err := ioutil.ReadFile("testkey")
		panic_the_err(err)
		key, err := ssh.ParseRawPrivateKey(tmp)
		privkey := key.(*rsa.PrivateKey)
		panic_the_err(err)
		plaintext, err := CredulousDecodeAES(ciphertext, privkey)
		So(err, ShouldEqual, nil)
		So(plaintext, ShouldEqual, "some plaintext")
	})
}

func TestDecodeWithSalt(t *testing.T) {
	Convey("Test Decode a string", t, func() {
		ciphertext := "sGhPCj9OCe0hv9PvWQvsu289sMsVNfqpyQDRCgXo+PwDMlXmRVXa5ErkkHNwyuYWFr9u1gkytiue7Dol4duvPycUYqpdeOOrfAMWkLWKGrO6tgTYtxMjVYBtp3negl2OeJqHFs6h/UwmNaO6IP2z2R8vPctmMmpwrkdzokiiPx6WKLDP17eoC+Q+zvDUqSTgqnSiwbjb+gFGFt7NTH65gHHHtwbm2wr45Oce4+LfddGo8V7A52ZjVlTHHdK+OiJzHmN8KMTAUi1d0ULI7oW+BfAX7iyA1SyvFx0oJHJ/dDidxPUm7i2vEeKtXU5BS8THv5dk01BwByJU+kl3qenCTA=="
		tmp, err := ioutil.ReadFile("testkey")
		panic_the_err(err)
		key, err := ssh.ParseRawPrivateKey(tmp)
		privkey := key.(*rsa.PrivateKey)
		panic_the_err(err)
		salt := "pepper"
		plaintext, err := CredulousDecodeWithSalt(ciphertext, salt, privkey)
		panic_the_err(err)
		So(plaintext, ShouldEqual, "some plaintext")
	})
}

func TestSSHFingerprint(t *testing.T) {
	Convey("Test generating SSH fingerprint", t, func() {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXg9Vmhy9YSB8BcN3yHgQjdX9lN3j2KRpv7kVDXSiIana2WbKP7IiTS0uJcJWUM3vlHjdL9KOO0jCWWzVFIcmLhiVVG+Fy2tothBp/NhjR8WWG/6Jg/6tXvVkLG6bDgfbDaLWdE5xzjL0YG8TrIluqnu0J5GHKrQcXF650PlqkGo+whpXrS8wOG+eUmsHX9L1w/Z3TkQlMjQNJEoRbqqSrp7yGj4JqzbtLpsglPRlobD7LHp+5ZDxzpk9i+6hoMxp2muDFxnEtZyED6IMQlNNEGkc3sdmGPOo26oW2+ePkBcjpOpdVif/Iya/jDLuLFHAOol6G34Tr4IdTgaL0qCCr TEST KEY"))
		panic_the_err(err)
		fingerprint := SSHFingerprint(pubkey)
		So(fingerprint, ShouldEqual, "c0:61:84:fc:e8:c9:52:dc:cd:a9:8e:82:a2:70:0a:30")
	})
}
