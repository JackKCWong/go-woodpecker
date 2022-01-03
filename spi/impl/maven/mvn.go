package maven

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/util/cmdutil"
	"github.com/go-cmd/cmd"
	"io"
	"path"
)

type mvn struct {
	Bin string
	POM string
}

func newMvn() mvn {
	return mvn{
		Bin: "mvn",
		POM: "pom.xml",
	}
}

func (m mvn) Run(ctx context.Context, task string, args ...string) (io.Reader, error) {
	goalRun := cmd.NewCmdOptions(cmd.Options{
		Buffered:  false,
		Streaming: true,
	}, m.bin(), mkargs(task, args)...)

	goalRun.Dir = m.Wd()
	errReader := cmdutil.NewErrReader()

	go func() {
		done := goalRun.Start()
		select {
		case <-ctx.Done():
			_ = goalRun.Stop()
			errReader.CloseWithError(fmt.Errorf("mvn exited with timeout: %w", ctx.Err()))
		case s := <-done:
			if s.Error != nil {
				errReader.CloseWithError(fmt.Errorf("mvn exited with err: %w", s.Error))
			}

			if s.Exit != 0 {
				errReader.CloseWithError(fmt.Errorf("mvn exited with code: %d", s.Exit))
			}

			errReader.Close()
		}
	}()

	return io.MultiReader(cmdutil.StdoutReader{Out: goalRun.Stdout}, errReader), nil
}

func mkargs(goal string, args []string) []string {
	cmdargs := []string{"-B"}
	cmdargs = append(cmdargs, goal)
	cmdargs = append(cmdargs, args...)

	return cmdargs
}

func (m mvn) bin() string {
	if m.Bin != "" {
		return m.Bin
	}

	return "mvn"
}

func (m mvn) Wd() string {
	if m.POM == "" {
		panic("pom.xml not specified")
	}

	dir, file := path.Split(m.POM)
	if file != "pom.xml" {
		panic(m.POM + " is not a valid pom.xml")
	}

	if dir == "" {
		return "."
	}

	return dir
}
