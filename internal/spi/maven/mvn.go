package maven

import (
	"context"
	"fmt"
	"github.com/go-cmd/cmd"
	"path"
	"strings"
)

type Mvn struct {
	Bin string
	POM string
}

func (m Mvn) DependencyUpdate(ctx context.Context, includes ...string) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "versions:use-next-releases", "-Dincludes="+strings.Join(includes, ","))
}

func (m Mvn) DependencyTree(ctx context.Context, outFile string) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "dependency:tree", "-DoutputFile="+outFile)
}

func (m Mvn) VersionCommit(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "versions:commit")
}

func (m Mvn) VulnerabilityReport(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "org.owasp:dependency-check-maven:check", "-DretireJsAnalyzerEnabled=false", "-DprettyPrint=true", "-Dformat=json")
}

func (m Mvn) Verify(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "verify")
}

func (m Mvn) mvnRun(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	errCh := make(chan error, 2)
	goalRun := cmd.NewCmdOptions(cmd.Options{
		Buffered:  false,
		Streaming: true,
	}, m.mvn(), append([]string{"-B"}, args...)...)

	goalRun.Dir = m.wd()
	done := goalRun.Start()
	go func() {
		defer close(errCh)
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
		case status := <-done:
			if status.Error != nil {
				errCh <- fmt.Errorf("err: %q, exit code: %q", status.Error, status.Exit)
			}

			if status.Exit != 0 {
				errCh <- fmt.Errorf("exit code: %q", status.Exit)
			}
		}
	}()

	return goalRun.Stdout, errCh
}

func (m Mvn) mvn() string {
	if m.Bin != "" {
		return m.Bin
	}

	return "mvn"
}

func (m Mvn) wd() string {
	if m.POM == "" {
		panic("pom.xml not specified")
	}

	dir, file := path.Split(m.POM)
	if file != "pom.xml" {
		panic(m.POM + " is not a valid pom.xml")
	}

	return dir
}
