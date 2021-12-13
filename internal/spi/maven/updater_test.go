package maven

import (
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ParseDepTree(t *testing.T) {
	depTree := parseDepTree(testDepTree)

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
