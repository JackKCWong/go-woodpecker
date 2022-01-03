package maven

import (
	"context"
	"encoding/json"
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestRunner_VulnerabilityReport(t *testing.T) {
	report, err := ioutil.ReadFile("testdata/target/dependency-check-report.json")
	if err != nil {
		t.Fatal(err)
	}

	vr := &VulnerabilityReport{}
	err = json.Unmarshal(report, vr)
	require.Nil(t, err)

	require.Greater(t, len(vr.Dependencies), 0, "expecting non-zero dependencies")

	hoc := vr.HighOrCritical()
	require.Greater(t, len(hoc), 0, "expecting non-zero Critical or High vulnerabilities")
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

func TestRunner_ParseDepTree(t *testing.T) {
	depTree := parseDepTree(testDepTree)

	require.Equal(t, 27, len(depTree.Nodes()))
	require.Equal(t, depTree.Get(1), api.DependencyTreeNode{
		ID:      "org.slf4j:slf4j-api:1.7.32",
		Type:    "jar",
		Scope:   "compile",
		Version: "1.7.32",
		Depth:   1,
		Raw:     "org.slf4j:slf4j-api:jar:1.7.32:compile",
	})
	require.Equal(t, depTree.Get(21), api.DependencyTreeNode{
		ID:      "org.jetbrains:annotations:13.0",
		Type:    "jar",
		Scope:   "test",
		Version: "13.0",
		Depth:   3,
		Raw:     "org.jetbrains:annotations:jar:13.0:test",
	})
}

type mockMvn struct {
}

func TestRunner_DependencyTree(t *testing.T) {
	runner := Runner{
		POM: "pom.xml",
		mvn: mockMvn{},
		opts: Opts{
			Output: os.Stdout,
		},
	}

	depTree, err := runner.DependencyTree()
	require.Nil(t, err)

	tcnative, found := depTree.Find("io.netty:netty-tcnative-classes:2.0.46.Final")
	require.True(t, found)
	require.Greater(t, len(tcnative.Vulnerabilities), 0)

	vulnerable, found := depTree.MostVulnerable()
	require.True(t, found)
	require.Len(t, vulnerable.Nodes(), 2)
	require.Equal(t, "io.netty:netty-handler:4.1.71.Final", vulnerable.Get(0).ID)
}

func (m mockMvn) Wd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return path.Join(wd, "testdata")
}

func (m mockMvn) Run(ctx context.Context, task string, args ...string) (io.Reader, error) {
	switch task {
	case "dependency:tree":
		outputFile := strings.Split(args[0], "=")[1]
		err := ioutil.WriteFile(outputFile, []byte(testDepTree), 0644)
		if err != nil {
			panic(err)
		}

		return strings.NewReader(""), nil
	case "org.owasp:dependency-check-maven:aggregate":
		return strings.NewReader(""), nil
	}

	panic("unexpected task")
}
