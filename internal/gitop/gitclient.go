package gitop

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitClient struct {
	Dir string
}

func (c GitClient) Commit(branchName, msg string) (string, error) {
	repo, err := git.PlainOpen(c.Dir)
	if err != nil {
		return "", err
	}

	branch, err := newBranch(repo, branchName)
	if err != nil {
		return "", err
	}

	wtree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	err = wtree.Checkout(&git.CheckoutOptions{
		Branch: branch.Name(),
		Keep:   true,
	})
	if err != nil {
		return "", err
	}

	err = wtree.AddGlob("**/pom.xml")

	if err != nil {
		return "", err
	}

	h, err := wtree.Commit(msg, &git.CommitOptions{
		All: true,
	})
	if err != nil {
		return "", err
	}

	return h.String(), nil
}

func newBranch(repo *git.Repository, name string) (*plumbing.Reference, error) {
	_, err := repo.Branch(name)
	if err == nil {
		// branch already exists
		return nil, fmt.Errorf("branch %s already exists", name)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+name), head.Hash())
	err = repo.Storer.SetReference(ref)
	if err != nil {
		return nil, err
	}

	return ref, nil
}
