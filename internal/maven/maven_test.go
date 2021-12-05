package maven

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
		panic("need to specify a test repo with env var WOODPECKER_REPO")
	}
}

func TestUpdateDependencies(t *testing.T) {
	mvn := Maven{
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

func drainStdout(t *testing.T, stdout <-chan string) []string {
	var lines []string
	for line := range stdout {
		t.Log(line)
		lines = append(lines, line)
	}

	return lines
}
