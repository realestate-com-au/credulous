package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"path"
	"strings"
	"syscall"
	"testing"
)

func touch_file(directory string, filename string, contents string) {
	location := path.Join(directory, filename)
	err := ioutil.WriteFile(location, []byte(contents), 0644)
	panic_the_err(err)
}

func TestGit(t *testing.T) {
	temp_dir, err := ioutil.TempDir("", "TestGit")
	panic_the_err(err)
	defer syscall.Rmdir(temp_dir)

	Convey("Resolving Git things", t, func() {
		Convey("When there is nothing to commit", func() {
			empty_repo_location := path.Join(temp_dir, "empty_repo")
			run_git_command_or_fail(temp_dir, "init", "empty_repo")
			So(func() { resolve(empty_repo_location) }, ShouldNotPanic)
		})
		Convey("When there are new files to add", func() {
			new_files_repo_location := path.Join(temp_dir, "new_files")
			run_git_command(temp_dir, "init", "new_files")

			touch_file(new_files_repo_location, "blah", "things")
			touch_file(new_files_repo_location, "stuff", "blah")
			So(func() { resolve(new_files_repo_location) }, ShouldNotPanic)

			out := run_git_command_or_fail(new_files_repo_location, "log", "--oneline")
			So(len(strings.Split(strings.TrimSpace(out.String()), "\n")), ShouldEqual, 1)
		})
		Convey("When there are modified files", func() {
			with_mod_repo_location := path.Join(temp_dir, "mod_files")
			run_git_command(temp_dir, "init", "mod_files")

			touch_file(with_mod_repo_location, "blah", "things")
			touch_file(with_mod_repo_location, "stuff", "blah")
			So(func() { resolve(with_mod_repo_location) }, ShouldNotPanic)

			// Add existing files
			out := run_git_command_or_fail(with_mod_repo_location, "log", "--oneline")
			So(len(strings.Split(strings.TrimSpace(out.String()), "\n")), ShouldEqual, 1)

			// Add some modifications
			touch_file(with_mod_repo_location, "blah", "some_change")
			out = run_git_command_or_fail(with_mod_repo_location, "status", "-s")
			So(len(out.String()), ShouldBeGreaterThan, 0)

			// Resolve the modifications
			So(func() { resolve(with_mod_repo_location) }, ShouldNotPanic)
			out = run_git_command_or_fail(with_mod_repo_location, "log", "--oneline")
			So(len(strings.Split(strings.TrimSpace(out.String()), "\n")), ShouldEqual, 2)
		})
		Convey("When there are non conflicting changes to pull", func() {
			repo1_location := path.Join(temp_dir, "repo1")
			run_git_command(temp_dir, "init", "repo1")
			touch_file(repo1_location, "blah", "some_change")
			So(func() { resolve(repo1_location) }, ShouldNotPanic)

			// Clone repo1
			repo2_location := path.Join(temp_dir, "repo2")
			run_git_command(temp_dir, "clone", repo1_location, repo2_location)

			// Now that repo1 is cloned, let's make a change that repo2 doesn't have
			touch_file(repo1_location, "blah", "third")
			So(func() { resolve(repo1_location) }, ShouldNotPanic)

			// Resolve repo2 so we get the new changes
			// Should have initial commit and the new changes commit
			So(func() { resolve(repo2_location) }, ShouldNotPanic)
			out := run_git_command_or_fail(repo2_location, "log", "--oneline")
			So(len(strings.Split(strings.TrimSpace(out.String()), "\n")), ShouldEqual, 2)
		})
		Convey("When there are conflicting changes to pull", func() {
			repo1_location := path.Join(temp_dir, "repo3")
			run_git_command(temp_dir, "init", "repo3")
			touch_file(repo1_location, "blah", "some_change")
			So(func() { resolve(repo1_location) }, ShouldNotPanic)

			// Clone repo1
			repo2_location := path.Join(temp_dir, "repo4")
			run_git_command(temp_dir, "clone", repo1_location, repo2_location)

			// Now that repo1 is cloned, let's make a change that repo2 doesn't have
			touch_file(repo1_location, "blah", "third")
			So(func() { resolve(repo1_location) }, ShouldNotPanic)

			// And make a conflicting change
			touch_file(repo2_location, "blah", "conflicting!")

			// Resolve repo2 so we get the new changes
			// Should have initial commit and the new changes commit
			// And our local change
			So(func() { resolve(repo2_location) }, ShouldNotPanic)
			out := run_git_command_or_fail(repo2_location, "log", "--oneline")
			So(len(strings.Split(strings.TrimSpace(out.String()), "\n")), ShouldEqual, 3)

			// And blah should have our contents
			b, err := ioutil.ReadFile(path.Join(repo2_location, "blah"))
			panic_the_err(err)
			So(string(b), ShouldEqual, "conflicting!")
		})
	})
}
