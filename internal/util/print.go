package util

import (
	"fmt"
	"io"
)

func Printfln(out io.Writer, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(out, format+"\n", args...)
	return err
}

func DrainLines(out io.Writer, lines <-chan string) {
	for l := range lines {
		_, _ = fmt.Fprintln(out, l)
	}
}
