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

var testDepTree = `test-project:hello-world:jar:SNAPSHOT
+- org.slf4j:slf4j-api:jar:1.7.32:compile
+- org.slf4j:slf4j-simple:jar:1.7.32:test
+- io.netty:netty-transport:jar:4.1.71.Final:compile
|  \- io.netty:netty-resolver:jar:4.1.71.Final:compile
+- io.netty:netty-handler:jar:4.1.71.Final:compile
|  \- io.netty:netty-tcnative-classes:jar:2.0.46.Final:compile
+- io.netty:netty-codec-http:jar:4.1.71.Final:compile
+- io.netty:netty-buffer:jar:4.1.71.Final:compile
+- io.netty:netty-common:jar:4.1.71.Final:compile
+- io.netty:netty-codec:jar:4.1.71.Final:compile
+- io.netty:netty-codec-http2:jar:4.1.71.Final:compile
+- javax.ws.rs:javax.ws.rs-api:jar:2.1.1:compile
+- org.hamcrest:hamcrest:jar:2.2:test
+- junit:junit:jar:4.13.2:test
+- org.eclipse.jetty:jetty-client:jar:9.4.31.v20200723:test
|  +- org.eclipse.jetty:jetty-http:jar:9.4.31.v20200723:test
|  \- org.eclipse.jetty:jetty-io:jar:9.4.31.v20200723:test
+- com.squareup.okhttp3:okhttp:jar:4.9.2:test
|  \- org.jetbrains.kotlin:kotlin-stdlib:jar:1.4.10:test
|     +- org.jetbrains.kotlin:kotlin-stdlib-common:jar:1.4.10:test
|     \- org.jetbrains:annotations:jar:13.0:test
+- com.squareup.okio:okio:jar:2.8.0:test
+- org.eclipse.jetty:jetty-util:jar:9.4.31.v20200723:test
+- org.json:json:jar:20210307:test
+- org.webjars:jquery:jar:1.12.0:test
\- org.webjars:jquery-ui:jar:1.12.1:test
`

func TestMavenUpdateDependencies(t *testing.T) {
	mvn := mvn{
		POM: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stdout, errors := mvn.DependencyUpdate(ctx, "io.netty:netty-tcnative-classes:2.0.46.Final")
	lines := drainStdout(t, stdout)

	err := <-errors
	require.Nilf(t, err, "failed to update deps: %q", err)
	require.Contains(t, lines, "[INFO] BUILD SUCCESS")

	//stdout, errors = mvn.Verify(ctx)
	//_ = drainStdout(t, stdout)
	//err = <-errors
	//require.Nilf(t, err, "failed to verify: %q", err)
	//require.Contains(t, lines, "[INFO] BUILD SUCCESS")
}

func TestMavenVerify(t *testing.T) {
	mvn := mvn{
		POM: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := mvn.Verify(ctx, os.Stdout)
	require.Nil(t, err)
}

func TestMavenDependencyTree(t *testing.T) {
	mvn := mvn{
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

func TestMavenVulnerabilityReport(t *testing.T) {
	mvn := mvn{
		POM: testRepo + "/pom.xml",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stdout, errors := mvn.DependencyCheck(ctx, "-Dformat=json", "-DretireJsAnalyzerEnabled=false")
	_ = drainStdout(t, stdout)
	err := <-errors
	require.Nilf(t, err, "failed to run dependency:tree, error:%q", err)
}

func TestVulnerabilityReport(t *testing.T) {
	report, err := ioutil.ReadFile(testRepo + "/target/dependency-check-report.json")
	require.Nil(t, err)

	vr := &VulnerabilityReport{}
	err = json.Unmarshal(report, vr)
	require.Nil(t, err)

	require.Greater(t, len(vr.Dependencies), 0, "expecting non-zero dependencies")

	hoc := vr.HighOrCritical()
	require.Greater(t, len(hoc), 0, "expecting non-zero Critical or High vulnerabilities")

	depTree := parseDepTree(testDepTree)
	vr.FillIn(&depTree)
	tcnative, found := depTree.Find("io.netty:netty-tcnative-classes:2.0.46.Final")
	require.True(t, found)
	require.Greater(t, len(tcnative.Vulnerabilities), 0)

	vulnerable, found := depTree.MostVulnerable()
	require.True(t, found)
	require.Len(t, vulnerable.Nodes, 2)
	require.Equal(t, "io.netty:netty-handler:4.1.71.Final", vulnerable.Get(0).ID)
}

func drainStdout(t *testing.T, stdout <-chan string) []string {
	var lines []string
	for line := range stdout {
		t.Log(line)
		lines = append(lines, line)
	}

	return lines
}
