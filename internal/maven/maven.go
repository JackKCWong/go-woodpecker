package maven

import (
	"bufio"
	"context"
	"os/exec"
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
	errCh := make(chan error, 2)
	goalRun := exec.CommandContext(ctx, m.mvn(), goal)
	goalRun.Dir = m.wd()
	stdoutPipe, err := goalRun.StdoutPipe()
	if err != nil {
		errCh <- err
		return nil, errCh
	}

	err = goalRun.Start()
	if err != nil {
		errCh <- err
		return nil, errCh
	}

	stdout := make(chan string, 1000)
	go func() {
		defer close(errCh)
		defer close(stdout)

		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				stdout <- scanner.Text()
			}

			if scanner.Err() != nil {
				errCh <- scanner.Err()
			}
		}()

		err := goalRun.Wait()
		if err != nil {
			errCh <- err
		}
	}()

	return stdout, errCh
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
