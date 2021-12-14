package api

type DependencyUpdater interface {
	CanContinueUpdate() bool
	UpdateDependency(depID string) error
	Verify() error
	StageUpdate() error
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

func (t DependencyTree) AllVulnerabilities() []Vulnerability {
	var all []Vulnerability

	for _, n := range t.Nodes {
		all = append(all, n.Vulnerabilities...)
	}

	return all
}

// MostVulnerable returns the sub tree with the highest CVSS score, if any
func (t DependencyTree) MostVulnerable() (DependencyTree, bool) {
	subTrees := make(map[string][]DependencyTreeNode)

	var currentTopNode string
	for i := 1; i < len(t.Nodes); i++ {
		if t.Nodes[i].Depth == 1 {
			currentTopNode = t.Nodes[i].ID
		}

		subTrees[currentTopNode] = append(subTrees[currentTopNode], t.Nodes[i])
	}

	sum := func(nodes []DependencyTreeNode) float64 {
		s := 0.0
		for _, n := range nodes {
			for _, v := range n.Vulnerabilities {
				s += v.CVSSv3Score + v.CVSSv2Score
			}
		}

		return s
	}

	var maxID string
	var maxScore float64
	for id, nodes := range subTrees {
		score := sum(nodes)
		if score > maxScore {
			maxID = id
			maxScore = score
		}
	}

	if maxScore > 0 {
		return DependencyTree{Nodes: subTrees[maxID]}, true
	}

	return DependencyTree{}, false
}

func (t DependencyTree) Root() DependencyTreeNode {
	return t.Nodes[0]
}
