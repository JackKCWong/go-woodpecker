package api

type DependencyUpdater interface {
	CanContinueUpdate() bool
	UpdateDependency() error
	Verify() error
	DependencyTree() DependencyTree
}

type DependencyTree struct {
	Nodes []DependencyTreeNode
}

type DependencyTreeNode struct {
	ID      string
	Type    string
	Scope   string
	Version string
	Depth   int
	Raw     string
}
