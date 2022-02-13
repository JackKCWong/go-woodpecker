package api

import "fmt"

type DependencyManager interface {
	UpdateDependency(dep DependencyTreeNode) (string, error)
	Verify() (TestReport, error)
	StageUpdate() error
	DependencyTree() (DependencyTree, error)
	IsMultiModules() (bool, error)
	SubModule(module string) (DependencyManager, error)
}

type TestReport struct {
	Passed  bool
	Summary string
	Error   string
}

type DependencyTree struct {
	nodes []DependencyTreeNode
}

func NewDependencyTree(nodes []DependencyTreeNode) DependencyTree {
	return DependencyTree{nodes: nodes}
}

type DependencyTreeNode struct {
	ID              string
	Type            string
	Group           string
	Artifact        string
	Version         string
	Scope           string
	Depth           int
	Raw             string
	PackageUrl      string
	Vulnerabilities []Vulnerability
}

type Vulnerability struct {
	Reference          string   `json:"reference"`
	CvssVector         string   `json:"cvssVector"`
	ExternalReferences []string `json:"externalReferences"`
	Cve                string   `json:"cve"`
	CvssScore          float64  `json:"cvssScore"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description"`
	Title              string   `json:"title"`
}

func (v Vulnerability) NVDUrl() string {
	return fmt.Sprintf("https://nvd.nist.gov/vuln/detail/%s", v.Cve)
}

func (v Vulnerability) Severity() string {
	switch {
	case v.CvssScore >= 9.0:
		return "Critical"
	case v.CvssScore >= 7.0:
		return "High"
	case v.CvssScore >= 4.0:
		return "Medium"
	default:
		return "Low"
	}
}

type ComponentReport struct {
	Reference       string          `json:"reference"`
	Coordinates     string          `json:"coordinates"`
	Description     string          `json:"description"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

func (t DependencyTree) Nodes() []DependencyTreeNode {
	return t.nodes
}

func (t DependencyTree) Get(i int) DependencyTreeNode {
	return t.nodes[i]
}

func (t DependencyTree) Set(i int, node DependencyTreeNode) {
	t.nodes[i] = node
}

func (t DependencyTree) Find(depID string) (DependencyTreeNode, bool) {
	for _, n := range t.Nodes() {
		if n.ID == depID {
			return n, true
		}
	}

	return DependencyTreeNode{}, false
}

func (t DependencyTree) AllVulnerabilities() []Vulnerability {
	var all []Vulnerability

	for _, n := range t.Nodes() {
		all = append(all, n.Vulnerabilities...)
	}

	return all
}

func (t DependencyTree) Subtree(i int, rootID string) (DependencyTree, bool) {
	var found bool
	var subtree []DependencyTreeNode
	for j := i; j < len(t.Nodes()); j++ {
		n := t.Get(j)
		if n.ID == rootID && n.Depth == 1 {
			found = true
			subtree = append(subtree, n)
			continue
		}

		if found && n.Depth == 1 {
			// stop at next root node
			break
		}

		if found {
			subtree = append(subtree, n)
		}
	}

	if found {
		return DependencyTree{subtree}, found
	}

	return DependencyTree{}, false
}

func (t DependencyTree) CriticalOrHigh() (DependencyTree, Vulnerability, bool) {
	for i, n := range t.Nodes() {
		if n.Depth <= 1 {
			subtree, _ := t.Subtree(i, n.ID)
			for _, v := range subtree.AllVulnerabilities() {
				if v.Severity() == "Critical" || v.Severity() == "High" {
					return subtree, v, true
				}
			}
		}
	}

	return DependencyTree{}, Vulnerability{}, false
}

// MostVulnerable returns the subtree with the highest CVSS score, if any
func (t DependencyTree) MostVulnerable() (DependencyTree, bool) {
	subTrees := make(map[string][]DependencyTreeNode)

	var currentTopNode string
	for i := 1; i < len(t.Nodes()); i++ {
		if t.Get(i).Depth == 1 {
			currentTopNode = t.Get(i).ID
		}

		subTrees[currentTopNode] = append(subTrees[currentTopNode], t.Get(i))
	}

	sum := func(nodes []DependencyTreeNode) float64 {
		s := 0.0
		for _, n := range nodes {
			for _, v := range n.Vulnerabilities {
				s += v.CvssScore
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
		return DependencyTree{nodes: subTrees[maxID]}, true
	}

	return DependencyTree{}, false
}

func (t DependencyTree) Root() DependencyTreeNode {
	return t.Get(0)
}

func (t DependencyTree) VulnerabilityCount() int {
	count := 0
	for _, n := range t.Nodes() {
		count += len(n.Vulnerabilities)
	}

	return count
}

func (t DependencyTree) FirstChildWithCVE(cveID string) (DependencyTree, bool) {
	for i, n := range t.Nodes() {
		if n.Depth == 1 {
			subtree, _ := t.Subtree(i, n.ID)
			for _, v := range subtree.AllVulnerabilities() {
				if v.Cve == cveID {
					return subtree, true
				}
			}
		}
	}

	return DependencyTree{}, false
}

func (t DependencyTree) FindCVE(cveID string) (Vulnerability, bool) {
	for _, n := range t.Nodes() {
		for _, v := range n.Vulnerabilities {
			if v.Cve == cveID {
				return v, true
			}
		}
	}

	return Vulnerability{}, false
}
