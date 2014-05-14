package main

import (
	"errors"
	"os"
	"fmt"
)

func panic_the_err(err error) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func ensure_directory(path string) {
	if path == "" {
		panic(errors.New("Can't ensure empty string as a directory!"))
	}

	err := os.MkdirAll(path, 0755)
	panic_the_err(err)
}
