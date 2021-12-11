package maven

import (
	"context"
	"github.com/stretchr/testify/require"
	"io/ioutil"
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

	stdout, errors := mvn.DependencyUpdate(ctx)
	lines := drainStdout(t, stdout)

	err := <-errors
	require.Nilf(t, err, "failed to update deps: %q", err)
	require.Contains(t, lines, "[INFO] BUILD SUCCESS")

	stdout, errors = mvn.Verify(ctx)
	_ = drainStdout(t, stdout)
	err = <-errors
	require.Nilf(t, err, "failed to verify: %q", err)
	require.Contains(t, lines, "[INFO] BUILD SUCCESS")
}

func TestMaven_DependencyTree(t *testing.T) {
	mvn := Maven{
		Pom: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	outFile, err := ioutil.TempFile(os.TempDir(), "dtree")
	require.Nil(t, err)

	stdout, errors := mvn.DependencyTree(ctx, outFile.Name())
	_ = drainStdout(t, stdout)
	err = <-errors
	require.Nilf(t, err, "failed to run dependency:tree, error:%q", err)

	tree, err := ioutil.ReadFile(outFile.Name())
	require.Nil(t, err)
	require.NotEmpty(t, tree)
	t.Logf("dependency tree: %s", tree)
}

func drainStdout(t *testing.T, stdout <-chan string) []string {
	var lines []string
	for line := range stdout {
		t.Log(line)
		lines = append(lines, line)
	}

	return lines
}
