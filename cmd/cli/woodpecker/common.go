package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/gitop"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func newProgressOutput() io.Writer {
	var progressOut = ioutil.Discard
	if viper.GetBool("verbose") {
		progressOut = os.Stdout
	} else if !viper.GetBool("noprogress") {
		progressOut = progressbar.DefaultBytes(-1, "working hard...")
	}

	return progressOut
}

// bindCmdOptsToViperConf replace '-' with '.' before binding so it can bind to nested properties more naturally
// eg. foo-bar is bound to foo.bar
func bindCmdOptsToViperConf(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(strings.ReplaceAll(f.Name, "-", "."), f)
	})
}

func readViperConf() error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			util.Printfln(os.Stdout, "config not found, continue...")
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

func newGitClient() (*gitop.GitClient, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir, err := gitop.FindGitDir(wd)
	if err != nil {
		return nil, err
	}

	return &gitop.GitClient{
		RepoDir: dir,
	}, nil
}
