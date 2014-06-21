package main

import "github.com/libgit2/git2go"

func gitAdd(repository, filename string) (err error) {
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	index, err := repo.Index()
	if err != nil {
		return err
	}

	err = index.AddByPath(filename)
	if err != nil {
		return err
	}

	err = index.Write()
	if err != nil {
		return err
	}
	return nil
}

func gitCommit(repo, comment string) (err error) {
	return nil
}
