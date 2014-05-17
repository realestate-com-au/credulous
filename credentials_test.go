package main

import (
	"crypto/rsa"
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

type TestFileList struct {
	testList []os.FileInfo
}

func (t *TestFileList) Readdir(n int) ([]os.FileInfo, error) {
	return t.testList, nil
}

type TestFileInfo struct {
	isDir bool
	name  string
}

func (t *TestFileInfo) IsDir() bool {
	return t.isDir
}

func (t *TestFileInfo) Name() string {
	return t.name
}

func (t *TestFileInfo) Size() int64 {
	return 0
}

func (t *TestFileInfo) Mode() os.FileMode {
	return 0
}

func (t *TestFileInfo) ModTime() time.Time {
	return time.Now()
}

func (t *TestFileInfo) Sys() interface{} {
	return nil
}

func TestGetDirs(t *testing.T) {
	Convey("Test finding all dirs", t, func() {
		Convey("Test with nothing", func() {
			t := TestFileList{}
			ents, err := getDirs(&t)
			So(err, ShouldEqual, nil)
			So(len(ents), ShouldEqual, 0)
		})
		Convey("Test with files only", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: false})
			i = append(i, &TestFileInfo{isDir: false})
			i = append(i, &TestFileInfo{isDir: false})
			t := TestFileList{testList: i}
			ents, err := getDirs(&t)
			So(err, ShouldEqual, nil)
			So(len(ents), ShouldEqual, 0)
		})
		Convey("Test with one dir", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: true})
			t := TestFileList{testList: i}
			ents, err := getDirs(&t)
			So(err, ShouldEqual, nil)
			So(len(ents), ShouldEqual, 1)
		})
		Convey("Test with multiple dirs", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: true})
			i = append(i, &TestFileInfo{isDir: true})
			i = append(i, &TestFileInfo{isDir: true})
			i = append(i, &TestFileInfo{isDir: true})
			t := TestFileList{testList: i}
			ents, err := getDirs(&t)
			So(err, ShouldEqual, nil)
			So(len(ents), ShouldEqual, 4)
		})
	})
}

func TestFindDefaultDir(t *testing.T) {
	Convey("Test Finding Default Dirs", t, func() {
		Convey("With no files or directories", func() {
			t := TestFileList{}
			_, err := findDefaultDir(&t)
			So(err.Error(), ShouldEqual, "No saved credentials found; please run 'credulous save' first")
		})
		Convey("With one file and no directories", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: false})
			t := TestFileList{testList: i}
			_, err := findDefaultDir(&t)
			So(err, ShouldNotEqual, nil)
			So(err.Error(), ShouldEqual, "No saved credentials found; please run 'credulous save' first")
		})
		Convey("With one file and one directory", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: false})
			i = append(i, &TestFileInfo{isDir: true, name: "foo"})
			t := TestFileList{testList: i}
			name, err := findDefaultDir(&t)
			So(err, ShouldEqual, nil)
			So(name, ShouldEqual, "foo")
		})
		Convey("With no files and more than one directory", func() {
			i := []os.FileInfo{}
			i = append(i, &TestFileInfo{isDir: true, name: "foo"})
			i = append(i, &TestFileInfo{isDir: true, name: "bar"})
			i = append(i, &TestFileInfo{isDir: true, name: "baz"})
			t := TestFileList{testList: i}
			_, err := findDefaultDir(&t)
			So(err, ShouldNotEqual, nil)
			So(err.Error(), ShouldEqual, "More than one account found; please specify account and user")
		})
	})
}

func TestValidateCredentials(t *testing.T) {
	Convey("Test credential validation", t, func() {
		// we can't really test ValidateCredentials directly,
		// because it calls verifyUserAndAccount, which
		// creates its own IAM connection. This is probably not
		// the best way to have implemented that function.
		// goamz provides an iamtest package, and we should
		// use that.
	})
}

func TestReadFile(t *testing.T) {
	Convey("Test Read File", t, func() {
		Convey("Valid Json returns Credential", func() {
			tmp, err := ioutil.ReadFile("testkey")
			panic_the_err(err)
			key, err := ssh.ParseRawPrivateKey(tmp)
			panic_the_err(err)
			privkey := key.(*rsa.PrivateKey)

			cred, _ := readCredentialFile("credential.json", privkey)
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

func TestListAvailableCreds(t *testing.T) {
	Convey("Test listing available credentials", t, func() {
		Convey("Test with no credentials", func() {
			tmp := TestFileList{}
			creds, err := listAvailableCredentials(&tmp)
			So(len(creds), ShouldEqual, 0)
			So(err.Error(), ShouldEqual, "No saved credentials found; please run 'credulous save' first")
		})
	})
}
