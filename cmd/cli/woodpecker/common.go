package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/gitop"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	"io"
	"net/url"
	"os"
	"time"
)

func newProgressOutput() io.WriteCloser {
	var progressOut io.WriteCloser
	if viper.GetBool("verbose") {
		progressOut = os.Stdout
	} else if !viper.GetBool("noprogress") {
		bar := progressbar.NewOptions64(-1,
			progressbar.OptionSetDescription("working hard..."),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(10),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionOnCompletion(func() {
				_, _ = fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		)
		_ = bar.RenderBlank()
		progressOut = bar
	} else {
		progressOut = util.Discard
	}

	return progressOut
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

func newGitHubClient() (*gitop.GitHub, error) {
	baseURL, err := url.Parse(viper.GetString("github.url"))
	if err != nil {
		return nil, err
	}

	return &gitop.GitHub{
		BaseURL:     baseURL,
		AccessToken: viper.GetString("github.accesstoken"),
	}, nil
}
