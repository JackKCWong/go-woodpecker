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

func (m Maven) MvnDependencyUpdate(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "versions:use-next-releases")
}

func (m Maven) MvnVerify(ctx context.Context) (<-chan string, <-chan error) {
	return m.mvnRun(ctx, "verify")
}

func (m Maven) mvnRun(ctx context.Context, goal string) (<-chan string, <-chan error) {
	goalRun := cmd.NewCmdOptions(cmd.Options{
		Streaming: true,
	}, m.mvn(), goal, "verify")

	goalRun.Dir = m.wd()
	done := goalRun.Start()
	errCh := make(chan error, 2)

	go func() {
		defer close(errCh)

		select {
		case <-ctx.Done():
			if ctx.Err() != nil {
				errCh <- ctx.Err()
			}

			err := goalRun.Stop()
			if err != nil {
				errCh <- err
			}
		case status := <-done:
			fmt.Printf("status: %v\n", status)
			if status.Error != nil {
				errCh <- status.Error
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
