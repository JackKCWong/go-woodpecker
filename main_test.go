package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"go-pacidae/internal/gitop"
	"go-pacidae/internal/maven"
	"os"
	"testing"
	"time"
)

var testRepo string

func init() {
	testRepo = os.Getenv("PACIDAE_REPO")
	if testRepo == "" {
		panic("need to specify a test repo with env var PACIDAE_REPO")
	}
}

func TestUpdateDependencies(t *testing.T) {
	mvn := maven.Maven{
		Pom: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stdout, errors := mvn.MvnDependencyUpdate(ctx)
	lines := drainStdout(t, stdout)

	err := <-errors
	require.Nilf(t, err, "failed to update deps: %q", err)
	require.Contains(t, lines, "[INFO] BUILD SUCCESS")

	stdout, errors = mvn.MvnVerify(ctx)
	_ = drainStdout(t, stdout)
	err = <-errors
	require.Nilf(t, err, "failed to verify: %q", err)
	require.Contains(t, lines, "[INFO] BUILD SUCCESS")
}

func TestGitPullRequest(t *testing.T) {
	gitClient := gitop.GitClient{Dir: testRepo}
	commit, err := gitClient.CommitAndPush("update-deps", "auto update dependencies")
	require.NotEmptyf(t, commit, "failed to commit: %q", err)
}

func drainStdout(t *testing.T, stdout <-chan string) []string {
	var lines []string
	for line := range stdout {
		t.Log(line)
		lines = append(lines, line)
	}

	return lines
}
