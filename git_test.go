package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/libgit2/git2go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGitAdd(t *testing.T) {
	Convey("Testing gitAdd", t, func() {
		Convey("Test add to non-existent repository", func() {
			_, err := gitAddCommitFile("/no/such/repo", "testdata/newcreds.json", "message")
			So(err, ShouldNotEqual, nil)
		})

		// need to create a new repo first
		repopath := path.Join("testrepo." + fmt.Sprintf("%d", os.Getpid()))
		repo, err := git.InitRepository(repopath, false)
		panic_the_err(err)

		Convey("Test add non-existent file to a repo", func() {
			_, err := gitAddCommitFile(repo.Path(), "/no/such/file", "message")
			// fmt.Println("gitAdd returned " + fmt.Sprintf("%s", err))
			So(err, ShouldNotEqual, nil)
		})

		Convey("Test add a file to the repo", func() {
			fp, _ := os.Create(path.Join(repopath, "testfile"))
			_, _ = fp.WriteString("A test string")
			_ = fp.Close()
			commitId, err := gitAddCommitFile(repo.Path(), "testfile", "message")
			So(err, ShouldEqual, nil)
			So(commitId, ShouldNotEqual, nil)
			So(commitId, ShouldNotBeBlank)
		})

		Convey("Test checking whether a repo is a repo", func() {
			fullpath, _ := filepath.Abs(repopath)
			isrepo, err := isRepo(fullpath)
			So(err, ShouldEqual, nil)
			So(isrepo, ShouldEqual, true)
		})

		// os.RemoveAll(path.Clean(path.Join(repo.Path(), "..")))

	})
}
