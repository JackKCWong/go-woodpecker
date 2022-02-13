package maven

import (
	"github.com/JackKCWong/go-woodpecker/api"
	"sort"
	"strings"
)

// HighOrCritical returns Dependency with CVSS score greater or equal to 7.0 (HIGH-CRITICAL)
func (vr *VulnerabilityReport) HighOrCritical() []Dependency {
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

func (vr *VulnerabilityReport) fillIn(tree *api.DependencyTree) {
	vuldb := make(map[string][]Vulnerability)
	pdb := make(map[string]string)

	for _, d := range vr.Dependencies {
		for _, v := range d.Vulnerabilities {
			gav := getMavenGAV(d.Packages[0].ID)
			vuldb[gav] = append(vuldb[gav], v)
			pdb[gav] = d.Packages[0].Url
		}
	}

	for i, n := range tree.Nodes() {
		n.Vulnerabilities = convertVul(vuldb[n.ID])
		n.PackageUrl = pdb[n.ID]
		tree.Set(i, n)
	}
}

func convertVul(vulnerabilities []Vulnerability) []api.Vulnerability {
	r := make([]api.Vulnerability, 0, len(vulnerabilities))

	for _, v := range vulnerabilities {
		r = append(r, api.Vulnerability{
			Cve:          v.Name,
			Description:  v.Description,
			Severity:     v.Severity,
			NVDReference: "https://nvd.nist.gov/vuln/detail/" + v.Name,
			CvssScore:    v.Cvssv3.BaseScore,
		})
	}

	sort.SliceStable(r, func(i, j int) bool {
		return r[i].CvssScore > r[j].CvssScore
	})

	return r
}

func getMavenGAV(packageId string) string {
	gav := strings.TrimPrefix(packageId, "pkg:maven/")

	gav = strings.ReplaceAll(gav, "/", ":")
	gav = strings.ReplaceAll(gav, "@", ":")

	return gav
}
