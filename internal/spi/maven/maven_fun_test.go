package maven

import (
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ParseDepTree(t *testing.T) {
	tree := `hello-world
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
	depTree := parseDepTree(tree)

	require.Equal(t, 26, len(depTree.Nodes))
	require.Equal(t, depTree.Nodes[0], api.DependencyTreeNode{
		ID:      "org.slf4j:slf4j-api:1.7.32",
		Type:    "jar",
		Scope:   "compile",
		Version: "1.7.32",
		Depth:   1,
		Raw:     "org.slf4j:slf4j-api:jar:1.7.32:compile",
	})
	require.Equal(t, depTree.Nodes[20], api.DependencyTreeNode{
		ID:      "org.jetbrains:annotations:13.0",
		Type:    "jar",
		Scope:   "test",
		Version: "13.0",
		Depth:   3,
		Raw:     "org.jetbrains:annotations:jar:13.0:test",
	})
}
