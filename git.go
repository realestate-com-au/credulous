package main

import (
	"bytes"
	"os/exec"
	"strings"
)

func run_git_command(git_dir string, args ...string) (bytes.Buffer, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = git_dir

	// We want to capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	return out, err
}

func run_git_command_or_fail(git_dir string, args ...string) bytes.Buffer {
	out, err := run_git_command(git_dir, args...)
	panic_the_err(err)
	return out
}

func resolve(git_dir string) {
	run_git_command_or_fail(git_dir, "add", "--all", ".")

	out := run_git_command_or_fail(git_dir, "status", "-s")
	if len(out.String()) > 0 {
		run_git_command(git_dir, "commit", "-m", "Totes made some changes or something...")
	}

	out = run_git_command_or_fail(git_dir, "remote", "-v", "show")
	if len(out.String()) > 0 {
		run_git_command_or_fail(git_dir, "fetch", "origin")
		_, err := run_git_command(git_dir, "rebase", "origin/master")
		if err != nil {
			out = run_git_command_or_fail(git_dir, "status", "-s")
			lines := strings.Split(strings.TrimSpace(out.String()), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "UU") {
					out, err = run_git_command(git_dir, "checkout", "--theirs", "--", line[3:])
				}
			}
			run_git_command_or_fail(git_dir, "add", "--all", ".")
			run_git_command_or_fail(git_dir, "rebase", "--continue")
		}

		run_git_command(git_dir, "push")
	}
}
