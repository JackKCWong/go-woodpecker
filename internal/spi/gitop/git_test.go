package gitop

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var testRepo string

func init() {
	testRepo = os.Getenv("WOODPECKER_REPO")
	if testRepo == "" {
		testRepo = os.ExpandEnv("$HOME/Workspace/repos/mu-server")
	}
}
func TestGitClient_CommitAndPush(t *testing.T) {
	gitClient := GitClient{RepoDir: testRepo}
	commit, err := gitClient.Commit("auto update dependencies")
	require.NotEmptyf(t, commit, "failed to commit: %q", err)
}

func TestGitClient_Origin(t *testing.T) {
	gitClient := GitClient{RepoDir: testRepo}
	origin, err := gitClient.Origin()
	require.Nil(t, err)
	require.Equal(t, "git@github.com:3redronin/mu-server.git", origin)
}

func TestGitHub_CreatePullRequest(t *testing.T) {
	token := os.Getenv("WOODPECKER_GITHUB_TOKEN")
	require.NotEmpty(t, token, "WOODPECKER_TOKEN is not defined")

	gh := GitHub{AccessToken: token}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := gh.CreatePullRequest(ctx, "git@github.com:JackKCWong/app-runner.git", "update-deps", "master",
		"DO NOT MERGE", "unit test")
	require.Nilf(t, err, "failed to create PR: %q", err)
}
