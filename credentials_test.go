package main

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

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

type TestFileOperator struct {
	Cred         *Credential
	FilePath     string
	ErrorOnWrite bool
}

func (t *TestFileOperator) WriteToDisk(cred *Credential, path string) error {
	t.Cred = cred
	t.FilePath = path
	if t.ErrorOnWrite {
		return errors.New("A test error")
	}
	return nil
}

func (t *TestFileOperator) ReadFromDisk(path string) (*Credential, error) {
	return nil, nil
}

func TestCredentialOperations(t *testing.T) {
	Convey("Test build Path", t, func() {
		Convey("it returns a full path to save file", func() {
			os.Setenv("HOME", "/home/user1")
			So(buildPath("account1", "username1"), ShouldEqual, "/home/user1/.credulous/local/account1/username1")
		})
	})

	Convey("Test build credential file name", t, func() {
		Convey("it returns a json filename with epoch time and last 8 chars of key", func() {
			keyTime, _ := time.Parse("2006-01-02", "2014-05-10")
			So(buildCredentialFileName(keyTime, "AKJJDOFIHKJD76DKHKSD"), ShouldEqual, "1399680000_76DKHKSD.json")
		})
	})

	Convey("Test Saving Credentials", t, func() {
		testOp := &TestFileOperator{}
		testCred := &Credential{
			CreateTime:       "2006-01-02T15:04:05Z",
			KeyId:            "AKHFJSKSVOSUFGHEIJHD",
			AccountAliasOrId: "account1",
			IamUsername:      "user1",
		}

		Convey("it passes the cred through to the file operator", func() {
			err := SaveCredential(testCred, testOp)
			So(err, ShouldEqual, nil)
			So(testOp.Cred, ShouldEqual, testCred)
		})

		Convey("it passes the correct file path", func() {
			os.Setenv("HOME", "/home/user1")
			err := SaveCredential(testCred, testOp)
			So(err, ShouldEqual, nil)
			So(testOp.FilePath, ShouldEqual, "/home/user1/.credulous/local/account1/user1/1136214245_FGHEIJHD.json")
		})

		Convey("it returns an error for a bad create time", func() {
			testCred.CreateTime = "this is not a date"
			err := SaveCredential(testCred, testOp)
			So(err, ShouldNotEqual, nil)
		})

		Convey("it returns an error if the file operator errors", func() {
			testOp.ErrorOnWrite = true
			err := SaveCredential(testCred, testOp)
			So(err, ShouldNotEqual, nil)
		})
	})
}
