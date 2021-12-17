package maven

import (
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ParseDepTree(t *testing.T) {
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

func Test_Verify(t *testing.T) {
	mvn := NewRunner(testRepo+"/pom.xml", Opts{
		Output: util.Discard,
	})

	result, err := mvn.Verify()
	require.Nil(t, err)

	require.True(t, result.Passed)
	require.Contains(t, "[INFO] Results:", result.Report)
	require.Contains(t, "[INFO] Tests run:", result.Report)
}
