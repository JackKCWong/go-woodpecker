package gitop

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestGitClient_CommitAndPush(t *testing.T) {
	repo := os.Getenv("WOODPECKER_REPO")
	require.NotEmpty(t, repo, "WOODPECKER_REPO is not defined")

	gitClient := GitClient{Dir: repo}
	commit, err := gitClient.CommitAndPush("update-deps", "auto update dependencies")
	require.NotEmptyf(t, commit, "failed to commit: %q", err)
}

func TestGitHub_CreatePullRequest(t *testing.T) {
	token := os.Getenv("WOODPECKER_GITHUB_TOKEN")
	require.NotEmpty(t, token, "WOODPECKER_TOKEN is not defined")

	gh := GitHub{AccessToken: token}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := gh.CreatePullRequest(ctx, "git@github.com:JackKCWong/app-runner.git", "update-deps", "master")
	require.Nilf(t, err, "failed to create PR: %q", err)
}
