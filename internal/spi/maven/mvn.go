package maven

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-cmd/cmd"
	"io"
	"path"
	"strings"
)

type mvn struct {
	Bin string
	POM string
}

func (m mvn) DependencyUpdate(ctx context.Context, includes ...string) (<-chan string, <-chan error) {
	var artifacts []string
	for _, include := range includes {
		artifacts = append(artifacts, m.getGroupArtifact(include))
	}

	return m.mvnRun(ctx, "versions:use-next-releases", "-Dincludes="+strings.Join(artifacts, ","))
}

func (m mvn) DependencyTree(ctx context.Context, outFile string) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "dependency:tree", "-DoutputFile="+outFile)
}

func (m mvn) VersionCommit(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "versions:commit")
}

func (m mvn) DependencyCheck(ctx context.Context, props ...string) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, append([]string{"org.owasp:dependency-check-maven:check"}, props...)...)
}

func (m mvn) Verify(ctx context.Context, output io.WriteCloser) error {
	status := m.run(ctx, output, "verify")
	if status.Error != nil {
		stderr := strings.Join(status.Stderr, "\n")
		return fmt.Errorf("mvn verify failed: err %w\n%s", status.Error, stderr)
	}

	if !status.Complete {
		return fmt.Errorf("mvn verify timeout")
	}

	if status.Exit != 0 {
		stderr := strings.Join(status.Stderr, "\n")
		return fmt.Errorf("mvn verify failed: exit %d\n%s", status.Exit, stderr)
	}

	return nil
}

func (m mvn) run(ctx context.Context, output io.WriteCloser, args ...string) cmd.Status {
	goal := cmd.NewCmdOptions(cmd.Options{
		Buffered:  false,
		Streaming: true,
	}, m.mvn(), append([]string{"-B"}, args...)...)

	goal.Dir = m.wd()
	done := goal.Start()

	go func() {
		for line := range goal.Stdout {
			fmt.Fprintln(output, line)
		}
		output.Close()
	}()

	select {
	case <-ctx.Done():
		return goal.Status()
	case status := <-done:
		return status
	}
}

func (m mvn) mvnRun(ctx context.Context, args ...string) (<-chan string, <-chan error) {
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

func (m mvn) mvn() string {
	if m.Bin != "" {
		return m.Bin
	}

	return "mvn"
}

func (m mvn) wd() string {
	if m.POM == "" {
		panic("pom.xml not specified")
	}

	dir, file := path.Split(m.POM)
	if file != "pom.xml" {
		panic(m.POM + " is not a valid pom.xml")
	}

	return dir
}

func (m mvn) getGroupArtifact(id string) string {
	split := strings.Split(id, ":")
	return split[0] + ":" + split[1]
}

type outputCollector interface {
	Include(string) bool
}

func collectStdout(stdout io.Reader, collectors ...outputCollector) []string {
	var out []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		for _, c := range collectors {
			if c.Include(line) {
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

func (c *testResultCollector) Include(line string) bool {
	if strings.HasPrefix(line, "[INFO] Results:") {
		c.started = true
		return true
	}

	if c.started {
		return strings.HasPrefix(line, "[INFO] Tests run:")
	}

	return false
}
