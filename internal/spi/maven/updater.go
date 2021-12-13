package maven

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Updater struct {
	POM  string
	mvn  *Mvn
	opts UpdaterOpts
}

type UpdaterOpts struct {
	Verbose bool
}

func NewUpdater(pom string, opts UpdaterOpts) *Updater {
	return &Updater{
		POM:  pom,
		mvn:  &Mvn{POM: pom},
		opts: opts,
	}
}

func (u Updater) CanContinueUpdate() bool {
	//TODO implement me
	panic("implement me")
}

func (u Updater) UpdateDependency() error {
	//TODO implement me
	panic("implement me")
}

func (u Updater) Verify() error {
	//TODO implement me
	panic("implement me")
}

func (u Updater) DependencyTree() (api.DependencyTree, error) {
	var tree api.DependencyTree
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err := u.checkStdout(u.mvn.VulnerabilityReport(ctx))
	if err != nil {
		return tree, err
	}

	temp, err := os.CreateTemp(os.TempDir(), "dtree")
	if err != nil {
		return tree, err
	}

	err = u.checkStdout(u.mvn.DependencyTree(ctx, temp.Name()))
	if err != nil {
		return tree, err
	}

	treeInBytes, err := ioutil.ReadFile(temp.Name())
	if err != nil {
		return tree, err
	}

	vr, err := u.loadVulnerabilityReport()
	if err != nil {
		return tree, err
	}

	tree = parseDepTree(string(treeInBytes))
	vr.FillIn(&tree)

	return tree, nil
}

func (u Updater) checkStdout(stdout <-chan string, errors <-chan error) error {
	if u.opts.Verbose {
		go util.DrainLines(os.Stdout, stdout)
	}

	err := <-errors
	if err != nil {
		return err
	}

	return nil
}

func (u Updater) loadVulnerabilityReport() (*VulnerabilityReport, error) {
	dir, _ := path.Split(u.POM)
	if dir == "" {
		dir = "."
	}

	report, err := ioutil.ReadFile(dir + "/target/dependency-check-report.json")
	if err != nil {
		return nil, err
	}

	vr := &VulnerabilityReport{}
	err = json.Unmarshal(report, vr)
	if err != nil {
		return nil, err
	}

	return vr, nil
}

func parseDepTree(content string) api.DependencyTree {
	prefixPattern := regexp.MustCompile("^\\W+")
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	nodes := make([]api.DependencyTreeNode, 0, 0)

	scanner.Scan()
	proj := scanner.Text()
	id, typ, ver, scope := parseTreeNode(proj)
	nodes = append(nodes, api.DependencyTreeNode{
		ID:      id,
		Type:    typ,
		Scope:   scope,
		Version: ver,
		Depth:   0,
		Raw:     proj,
	})

	for scanner.Scan() {
		line := scanner.Text()
		prefix := prefixPattern.FindString(line)
		depth := len(prefix) / 3
		raw := string(prefixPattern.ReplaceAll([]byte(line), []byte("")))
		id, typ, ver, scope := parseTreeNode(raw)
		nodes = append(nodes, api.DependencyTreeNode{
			ID:      id,
			Type:    typ,
			Scope:   scope,
			Version: ver,
			Depth:   depth,
			Raw:     raw,
		})
	}

	return api.DependencyTree{Nodes: nodes}
}

func parseTreeNode(line string) (string, string, string, string) {
	parts := strings.Split(line, ":")

	scope := ""
	if len(parts) > 4 {
		scope = parts[4]
	}

	return strings.Join([]string{parts[0], parts[1], parts[3]}, ":"),
		parts[2], parts[3], scope
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