package maven

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/spi"
	"io"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"time"
)

type Runner struct {
	POM  string
	mvn  spi.TaskRunner
	opts Opts
}

type Opts struct {
	Output               io.WriteCloser
	DependencyCheckProps []string
}

func New(pom string, opts Opts) api.DependencyManager {
	return &Runner{
		POM:  pom,
		mvn:  mvn{POM: pom},
		opts: opts,
	}
}

func (u Runner) UpdateDependency(dep api.DependencyTreeNode) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	out, err := u.mvn.Run(ctx, "versions:use-next-releases", "-Dincludes="+getGroupArtifact(dep.ID))
	if err != nil {
		return "", err
	}

	prefix := "[INFO] Updated " + fmt.Sprintf("%s:%s:%s", getGroupArtifact(dep.ID), dep.Type, dep.Version) + " to version "
	captured := collectStdout(io.TeeReader(out, u.opts.Output), func(line string) bool {
		return strings.HasPrefix(line, prefix)
	})

	if len(captured) == 0 {
		return "", nil
	}

	newVersion := strings.Split(captured[0], " to version ")[1]
	return strings.ReplaceAll(dep.ID, dep.Version, newVersion), nil
}

func (u Runner) Verify() (api.TestReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	stdout, err := u.mvn.Run(ctx, "verify")
	report := collectStdout(io.TeeReader(stdout, u.opts.Output), new(testResultCollector).IsTestResult)

	if err != nil {
		return api.TestReport{
			Passed:  false,
			Summary: strings.Join(report, "\n"),
		}, err
	}

	return api.TestReport{
		Passed:  true,
		Summary: strings.Join(report, "\n"),
	}, nil
}

func (u Runner) DependencyTree() (api.DependencyTree, error) {
	var tree api.DependencyTree
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	props := make([]string, 0, len(u.opts.DependencyCheckProps)+1)
	props = append(props, "-Dformats=json,html")
	for _, v := range u.opts.DependencyCheckProps {
		props = append(props, fmt.Sprintf("-D%s", v))
	}

	stdout, err := u.mvn.Run(ctx, "org.owasp:dependency-check-maven:aggregate", props...)
	if err != nil {
		return tree, err
	}

	_, _ = io.Copy(u.opts.Output, stdout)

	tempFile, err := ioutil.TempFile(path.Join(u.mvn.Wd(), "target"), "woodpecker-maven-dependency-tree")
	if err != nil {
		return tree, err
	}

	stdout, err = u.mvn.Run(ctx, "dependency:tree", "-DoutputFile="+tempFile.Name(), "-DappendOutput=true")
	if err != nil {
		return tree, err
	}

	_, _ = io.Copy(u.opts.Output, stdout)

	treeInBytes, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return tree, err
	}

	vr, err := u.loadVulnerabilityReport()
	if err != nil {
		return tree, err
	}

	tree = parseDepTree(string(treeInBytes))
	vr.fillIn(&tree)

	return tree, nil
}

func (u Runner) loadVulnerabilityReport() (*VulnerabilityReport, error) {
	report, err := ioutil.ReadFile(u.mvn.Wd() + "/target/dependency-check-report.json")
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

func (u Runner) StageUpdate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stdout, err := u.mvn.Run(ctx, "versions:commit")
	if err != nil {
		return err
	}

	_, _ = io.Copy(u.opts.Output, stdout)

	return nil
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

	return api.NewDependencyTree(nodes)
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

func getGroupArtifact(id string) string {
	split := strings.Split(id, ":")
	return split[0] + ":" + split[1]
}

type collector func(string) bool

func collectStdout(stdout io.Reader, collectors ...collector) []string {
	var out []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		for _, c := range collectors {
			if c(line) {
				out = append(out, line)
				break
			}
		}
	}

	return out
}

type testResultCollector struct {
	started bool
}

func (c *testResultCollector) IsTestResult(line string) bool {
	if strings.HasPrefix(line, "[INFO] Results:") {
		c.started = true
		return false
	}

	if c.started {
		return strings.HasPrefix(line, "[INFO] Tests run:")
	}

	return false
}