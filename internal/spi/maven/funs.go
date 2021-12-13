package maven

import (
	"bufio"
	"bytes"
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"regexp"
	"sort"
	"strings"
)

// FindVulnerabilities returns Dependency with CVSS score greater or equal to 7.0 (HIGH-CRITICAL)
func FindVulnerabilities(vr *VulnerabilityReport) []Dependency {
	deps := vr.Dependencies

	vulnerables := filterDeps(deps, func(d Dependency) bool {
		return len(d.Vulnerabilities) > 0
	})

	findCoH := func(v []Vulnerability) []Vulnerability {
		return filterVuls(v, func(v Vulnerability) bool {
			return v.Cvssv3.BaseScore >= 7.0 || v.Cvssv2.Score >= 7.0
		})
	}

	highOrCritical := filterDeps(vulnerables, func(d Dependency) bool {
		coh := findCoH(d.Vulnerabilities)
		return len(coh) > 0
	})

	sort.SliceStable(highOrCritical, func(i, j int) bool {
		return len(findCoH(highOrCritical[i].Vulnerabilities)) > len(findCoH(highOrCritical[j].Vulnerabilities))
	})

	return highOrCritical
}

func parseDepTree(content string) api.DependencyTree {
	prefixPattern := regexp.MustCompile("^\\W+")
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	scanner.Scan() // skip 1st line

	nodes := make([]api.DependencyTreeNode, 0, 0)
	for scanner.Scan() {
		line := scanner.Text()
		prefix := prefixPattern.FindString(line)
		depth := len(prefix) / 3
		raw := string(prefixPattern.ReplaceAll([]byte(line), []byte("")))
		parts := strings.Split(raw, ":")
		nodes = append(nodes, api.DependencyTreeNode{
			ID:      strings.Join([]string{parts[0], parts[1], parts[3]}, ":"),
			Type:    parts[2],
			Scope:   parts[4],
			Version: parts[3],
			Depth:   depth,
			Raw:     raw,
		})
	}

	return api.DependencyTree{Nodes: nodes}
}

func filterVuls(vuls []Vulnerability, f func(v Vulnerability) bool) []Vulnerability {
	result := make([]Vulnerability, 0, len(vuls))

	for _, v := range vuls {
		if f(v) {
			result = append(result, v)
		}
	}

	return result
}

func filterDeps(deps []Dependency, f func(d Dependency) bool) []Dependency {
	result := make([]Dependency, 0, len(deps))

	for _, d := range deps {
		if f(d) {
			result = append(result, d)
		}
	}

	return result
}
