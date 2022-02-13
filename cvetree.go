package woodpecker

import (
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/api"
)

type TreeOpts struct {
	Opts
}

func (wp Woodpecker) Tree(opts TreeOpts) (api.DependencyTree, error) {
	tree, err := wp.DepMgr.DependencyTree()
	if err != nil {
		return api.DependencyTree{}, err
	}

	var coordinates []string
	idx := make(map[string]*api.DependencyTreeNode)
	nodes := tree.Nodes()
	for i := range nodes {
		coordinate := fmt.Sprintf("pkg:maven/%s/%s@%s", nodes[i].Group, nodes[i].Artifact, nodes[i].Version)
		coordinates = append(coordinates, coordinate)
		idx[coordinate] = &nodes[i]
	}

	reports, err := wp.OSSIndex.GetComponentReports(coordinates)
	if err != nil {
		return api.DependencyTree{}, err
	}

	if opts.Verbose {
		cves, _ := json.Marshal(reports)
		fmt.Printf("%s\n", string(cves))
	}

	for _, r := range reports {
		if len(r.Vulnerabilities) > 0 {
			idx[r.Coordinates].Vulnerabilities = r.Vulnerabilities
		}
	}

	return tree, nil
}
