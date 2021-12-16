package main

import (
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"os"
)

func newProgressOutput(verbose, noProgress bool) io.Writer {
	var progressOut = ioutil.Discard
	if verbose {
		progressOut = os.Stdout
	} else if !noProgress {
		progressOut = progressbar.DefaultBytes(-1, "working hard...")
	}

	return progressOut
}
