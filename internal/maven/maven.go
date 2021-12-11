package maven

import (
	"context"
	"fmt"
	"github.com/go-cmd/cmd"
	"path"
)

type Maven struct {
	Bin string
	Pom string
}

func (m Maven) DependencyUpdate(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "versions:use-next-releases")
}

func (m Maven) DependencyTree(ctx context.Context, outFile string) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "dependency:tree", "-DoutputFile="+outFile)
}

func (m Maven) Verify(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "verify")
}

func (m Maven) mvnRun(ctx context.Context, args ...string) (<-chan string, <-chan error) {
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

func (m Maven) mvn() string {
	if m.Bin != "" {
		return m.Bin
	}

	return "mvn"
}

func (m Maven) wd() string {
	if m.Pom == "" {
		panic("pom.xml not specified")
	}

	dir, file := path.Split(m.Pom)
	if file != "pom.xml" {
		panic(m.Pom + " is not a valid pom.xml")
	}

	return dir
}
