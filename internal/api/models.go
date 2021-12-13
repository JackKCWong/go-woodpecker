package api

type DependencyUpdater interface {
	CanContinueUpdate() bool
	UpdateDependency() error
	Verify() error
	DependencyTree() (DependencyTree, error)
}

type DependencyTree struct {
	Nodes []DependencyTreeNode
}

type DependencyTreeNode struct {
	ID              string
	Type            string
	Scope           string
	Version         string
	Depth           int
	Raw             string
	PackageUrl      string
	ShouldUpdate    bool
	Vulnerabilities []Vulnerability
}

type Vulnerability struct {
	ID          string
	Descrption  string
	Source      string
	Severity    string
	CVEUrl      string
	CVSSv2Score float64
	CVSSv3Score float64
}

func (t DependencyTree) Find(depID string) (DependencyTreeNode, bool) {
	for _, n := range t.Nodes {
		if n.ID == depID {
			return n, true
		}
	}

	return DependencyTreeNode{}, false
}
