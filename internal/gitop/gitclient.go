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
	"time"
)

type GitClient struct {
	Dir         string
	AccessToken string
}

func (c GitClient) CommitAndPush(branchName, msg string) (string, error) {
	repo, err := git.PlainOpen(c.Dir)
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

	err = wtree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		Create: true,
		Keep:   true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to checkout branch: %w", err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	auth, err := gitAuth()
	if err != nil {
		return "", fmt.Errorf("failed to get auth: %w", err)
	}

	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})

	if err != nil {
		return "", fmt.Errorf("failed to push: %w", err)
	}

	return h.String(), nil
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
