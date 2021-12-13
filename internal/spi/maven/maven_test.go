package maven

import (
	"context"
	"encoding/json"
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
		testRepo = os.ExpandEnv("$HOME/Workspace/repos/mu-server")
	}
}

func TestMaven_UpdateDependencies(t *testing.T) {
	mvn := Mvn{
		POM: testRepo + "/pom.xml",
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
	mvn := Mvn{
		POM: testRepo + "/pom.xml",
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

func TestMvn_VulnerabilityReport(t *testing.T) {
	mvn := Mvn{
		POM: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stdout, errors := mvn.VulnerabilityReport(ctx)
	_ = drainStdout(t, stdout)
	err := <-errors
	require.Nilf(t, err, "failed to run dependency:tree, error:%q", err)
}

func Test_FindVulnerabilities(t *testing.T) {
	report, err := ioutil.ReadFile(testRepo + "/target/dependency-check-report.json")
	require.Nil(t, err)

	vr := &VulnerabilityReport{}
	err = json.Unmarshal(report, vr)
	require.Nil(t, err)

	require.Greater(t, len(vr.Dependencies), 0, "expecting non-zero dependencies")

	cohVuls := FindVulnerabilities(vr)
	require.Greater(t, len(cohVuls), 0, "expecting non-zero Critical or High vulnerabilities")
}

func drainStdout(t *testing.T, stdout <-chan string) []string {
	var lines []string
	for line := range stdout {
		t.Log(line)
		lines = append(lines, line)
	}

	return lines
}
