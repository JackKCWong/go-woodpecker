package util

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"strings"
)

func Printfln(out io.Writer, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(out, format+"\n", args...)
	return err
}

func DrainLines(out io.Writer, lines <-chan string) {
	replacer := strings.NewReplacer(
		"[INFO]", color.HiBlueString("[INFO]"),
		"[WARNING]", color.HiYellowString("[WARNING]"),
		"[ERROR]", color.HiRedString("[ERROR]"),
		"BUILD SUCCESS", color.HiGreenString("BUILD SUCCESS"),
		"BUILD FAILURE", color.HiRedString("BUILD FAILURE"))

	for l := range lines {
		_, _ = fmt.Fprintln(out, replacer.Replace(l))
	}
}
