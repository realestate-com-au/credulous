package main

import (
	"os"
	"errors"
)

func panic_the_err(err error) {
	if err != nil {
		panic(err)
	}
}

func ensure_directory(path string) {
	if path == "" {
		panic(errors.New("Can't ensure empty string as a directory!"))
	}

	err := os.MkdirAll(path, 0755)
	panic_the_err(err)
}
