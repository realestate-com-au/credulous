package main

import (
	"crypto/rsa"
	"io/ioutil"
	"testing"

	"code.google.com/p/go.crypto/ssh"
	. "github.com/smartystreets/goconvey/convey"
	// "io/ioutil"
	// "fmt"
)

type TestWriter struct {
	Written []byte
}

func (t *TestWriter) Write(p []byte) (n int, err error) {
	t.Written = p
	return 0, nil
}

func TestReadFile(t *testing.T) {
	Convey("Test Read File", t, func() {
		Convey("Valid Json returns Credential", func() {

			tmp, err := ioutil.ReadFile("testkey")
			panic_the_err(err)
			key, err := ssh.ParseRawPrivateKey(tmp)
			panic_the_err(err)
			privkey := key.(*rsa.PrivateKey)

			cred := readCredentialFile("credential.json", privkey)
			So(cred.LifeTime, ShouldEqual, 22)
		})
		Convey("Credentials display correctly", func() {
			cred := Credential{KeyId: "ABC", SecretKey: "SECRET"}
			testWriter := TestWriter{}
			cred.Display(&testWriter)
			So(string(testWriter.Written), ShouldEqual, "export AWS_ACCESS_KEY_ID=ABC\nexport AWS_SECRET_ACCESS_KEY=SECRET\n")
		})

		Convey("Saving credentials", func() {
			// temp_dir, err := ioutil.TempDir("", "SavingCredentialsTest")
			// panic_the_err(err)
			//
			// cred := Credential{KeyId: "ABC", SecretKey: "SECRET"}
			// new_filename := Save(cred, temp_dir)
			//
			// new_cred := readCredentialFile(new_filename)
			// fmt.Println(new_cred)

		})
	})
}
