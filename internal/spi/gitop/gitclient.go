package gitop

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os"
	"path"
	"path/filepath"
)

type GitClient struct {
	RepoDir string
}

func (c GitClient) Origin() (string, error) {
	repo, err := git.PlainOpen(c.RepoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	origin, err := repo.Remote("origin")
	if err != nil {
		return "", err
	}

	return origin.Config().URLs[0], nil
}

func (c GitClient) Branch(name string) error {
	repo, err := git.PlainOpen(c.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	wtree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open worktree: %w", err)
	}

	err = wtree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + name),
		Create: true,
		Keep:   true,
	})

	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

func (c GitClient) Commit(msg string) (string, error) {
	repo, err := git.PlainOpen(c.RepoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	wtree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to open worktree: %w", err)
	}

	status, err := wtree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree status: %w", err)
	}

	if status.IsClean() {
		return "", fmt.Errorf("nothing to commit")
	}

	err = wtree.AddGlob("**pom.xml")
	if err != nil {
		return "", fmt.Errorf("failed to track files: %w", err)
	}

	h, err := wtree.Commit(msg, &git.CommitOptions{
		All: true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to commit files: %w", err)
	}

	return h.String(), nil
}

func (c GitClient) Push(ctx context.Context) error {
	auth, err := gitAuth()
	if err != nil {
		return fmt.Errorf("failed to get auth: %w", err)
	}

	repo, err := git.PlainOpen(c.RepoDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})

	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func (c GitClient) Clone(ctx context.Context, url string) error {
	panic("not implemented")
}

func FindGitDir(dir string) (string, error) {
	if _, err := os.Stat(path.Join(dir, ".git")); err == nil {
		return dir, nil
	}

	if filepath.IsAbs(dir) {
		dir = filepath.Dir(dir)
	} else {
		dir = filepath.Join(dir, "..")
	}

	if _, err := os.Stat(path.Join(dir, ".git")); err == nil {
		return dir, nil
	}

	return "", fmt.Errorf("not a git repo")
}

func newSshPubKeyAuth() (*ssh.PublicKeys, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sshKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	_, err = os.Stat(sshKeyFile)
	if os.IsNotExist(err) {
		sshKeyFile = path.Join(homeDir, ".ssh", "id_ed25519")
	}

	_, err = os.Stat(sshKeyFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("cannot find id_rsa or id_ed25519 in ~/.ssh :%w", err)
	}

	return ssh.NewPublicKeysFromFile("git", sshKeyFile, "")
}

func gitAuth() (transport.AuthMethod, error) {
	var auth transport.AuthMethod
	var err error
	auth, err = newSshPubKeyAuth()
	if err != nil {
		// fallback to key auth
		auth, err = ssh.NewSSHAgentAuth("")
	}

	if err != nil {
		return nil, err
	}

	return auth, nil
}
