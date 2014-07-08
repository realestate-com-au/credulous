package main

import (
	"errors"
	"path"
	"time"

	"github.com/libgit2/git2go"
)

type RepoConfig struct {
	Name  string
	Email string
}

func isGitRepo(checkpath string) (bool, error) {
	ceiling := []string{checkpath}

	repopath, err := git.Discover(checkpath, false, ceiling)
	nonRepoErr := errors.New("Could not find repository from '" + checkpath + "'")
	if err != nil && err.Error() != nonRepoErr.Error() {
		return false, err
	}
	if err.Error() == nonRepoErr.Error() {
		return false, nil
	}
	// the path is the parent of the repo, which appends '.git'
	// to the path
	dirpath := path.Dir(path.Clean(repopath))
	if dirpath == checkpath {
		return true, nil
	}
	return false, nil
}

func getRepoConfig(repo *git.Repository) (RepoConfig, error) {
	config, err := repo.Config()
	if err != nil {
		return RepoConfig{}, err
	}
	name, err := config.LookupString("user.name")
	if err != nil {
		return RepoConfig{}, err
	}
	email, err := config.LookupString("user.email")
	if err != nil {
		return RepoConfig{}, err
	}
	repoconf := RepoConfig{
		Name:  name,
		Email: email,
	}
	return repoconf, nil
}

func gitAddCommitFile(repopath, filename, message string) (commitId string, err error) {
	repo, err := git.OpenRepository(repopath)
	if err != nil {
		return "", err
	}

	config, err := getRepoConfig(repo)
	if err != nil {
		return "", err
	}

	index, err := repo.Index()
	if err != nil {
		return "", err
	}

	err = index.AddByPath(filename)
	if err != nil {
		return "", err
	}

	err = index.Write()
	if err != nil {
		return "", err
	}

	treeId, err := index.WriteTree()
	if err != nil {
		return "", err
	}

	// new file is now staged, so we have to create a commit
	sig := &git.Signature{
		Name:  config.Name,
		Email: config.Email,
		When:  time.Now(),
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return "", err
	}

	var commit *git.Oid
	haslog, err := repo.HasLog("HEAD")
	if err != nil {
		return "", err
	}
	if !haslog {
		// In this case, the repo has been initialized, but nothing has ever been committed
		commit, err = repo.CreateCommit("HEAD", sig, sig, message, tree)
		if err != nil {
			return "", err
		}
	} else {
		// In this case, the repo has commits
		currentBranch, err := repo.Head()
		if err != nil {
			return "", err
		}
		currentTip, err := repo.LookupCommit(currentBranch.Target())
		if err != nil {
			return "", err
		}
		commit, err = repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			return "", err
		}
	}

	return commit.String(), nil
}
