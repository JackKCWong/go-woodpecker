package cmdutil

import (
	"bytes"
	"fmt"
)

type StdoutReader struct {
	Out <-chan string
	buf bytes.Buffer
}

func (r StdoutReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if line, ok := <-r.Out; ok {
		fmt.Fprintln(&r.buf, line)
	}

	return r.buf.Read(p)
}
