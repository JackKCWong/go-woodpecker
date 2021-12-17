package maven

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	POM  string
	mvn  *mvn
	opts Opts
}

type Opts struct {
	Output               io.WriteCloser
	DependencyCheckProps map[string]string
}

func NewRunner(pom string, opts Opts) *Runner {
	return &Runner{
		POM:  pom,
		mvn:  &mvn{POM: pom},
		opts: opts,
	}
}

func (u Runner) CanContinueUpdate() bool {
	//TODO implement me
	panic("implement me")
}

func (u Runner) UpdateDependency(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := u.drainStdout(u.mvn.DependencyUpdate(ctx, id))
	if err != nil {
		return err
	}

	return nil
}

func (u Runner) Verify() (api.VerificationResult, error) {
	stdoutR, stdoutW := io.Pipe()
	report := bytes.Buffer{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := collectStdout(stdoutR, new(testResultCollector))
		report.WriteString(strings.Join(r, "\n"))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	err := u.mvn.Verify(ctx, util.MultiWriteCloser(stdoutW, u.opts.Output))
	wg.Wait()

	if err != nil {
		return api.VerificationResult{
			Passed: false,
			Report: report.String(),
		}, err
	}

	return api.VerificationResult{
		Passed: true,
		Report: report.String(),
	}, nil
}

func (u Runner) DependencyTree() (api.DependencyTree, error) {
	var tree api.DependencyTree
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	props := make([]string, 0, len(u.opts.DependencyCheckProps)+1)
	props = append(props, "-Dformat=json")
	for k, v := range u.opts.DependencyCheckProps {
		props = append(props, fmt.Sprintf("-D%s=%s", k, v))
	}

	err := u.drainStdout(u.mvn.DependencyCheck(ctx, props...))
	if err != nil {
		return tree, err
	}

	temp, err := os.CreateTemp(os.TempDir(), "dtree")
	if err != nil {
		return tree, err
	}

	err = u.drainStdout(u.mvn.DependencyTree(ctx, temp.Name()))
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

func (u Runner) drainStdout(stdout <-chan string, errors <-chan error) error {
	go util.DrainLines(u.opts.Output, stdout)

	err := <-errors
	if err != nil {
		return err
	}

	return nil
}

func (u Runner) loadVulnerabilityReport() (*VulnerabilityReport, error) {
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

func (u Runner) StageUpdate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := u.drainStdout(u.mvn.VersionCommit(ctx))
	if err != nil {
		return err
	}

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
