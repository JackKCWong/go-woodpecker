package config

import (
	"embed"
	"errors"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/JackKCWong/go-woodpecker/spi"
	"github.com/JackKCWong/go-woodpecker/spi/impl/gitcmd"
	"github.com/JackKCWong/go-woodpecker/spi/impl/github"
	"github.com/JackKCWong/go-woodpecker/spi/impl/ossindex"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed template.yaml
var YamlTemplate embed.FS

func NewProgressOutput() io.WriteCloser {
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

func ReadConfigFile() error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			return fmt.Errorf("%s/.woodpecker.yaml not found, use `woodpecker config` to show you how to config", os.Getenv("HOME"))
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

func NewGitClient() (spi.GitClient, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dir, err := gitcmd.FindGitDir(wd)
	if err != nil {
		return nil, err
	}

	return &gitcmd.GitClient{
		RepoDir: dir,
	}, nil
}

func NewGitHub() (spi.GitServer, error) {
	apiUrlStr := viper.GetString("github.api-url")
	if apiUrlStr == "" {
		return nil, errors.New("github.api-url is not set")
	}

	if !strings.HasSuffix(apiUrlStr, "/") {
		apiUrlStr += "/"
	}

	apiURL, err := url.Parse(apiUrlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse github.api-url: %w", err)
	}

	accessToken := viper.GetString("github.access-token")
	if accessToken == "" {
		return nil, errors.New("github.access-token is not set")
	}

	return github.New(github.GitHub{
		ApiURL:      apiURL,
		AccessToken: accessToken,
	}), nil
}

func NewOSSIndexClient() (spi.OSSIndex, error) {
	apiUrl := viper.GetString("ossindex.api-url")
	if apiUrl == "" {
		return nil, errors.New("ossindex.api-url is not set")
	}

	token := viper.GetString("ossindex.token")
	if token == "" {
		return nil, errors.New("ossindex.token is not set")
	}

	timeout := viper.GetInt("ossindex.timeout")
	if timeout == 0 {
		timeout = 30
	}

	return ossindex.Sonatype{
		APIBaseURL:   apiUrl,
		APIBasicAuth: token,
		HttpClient:   http.Client{Timeout: time.Duration(timeout) * time.Second},
	}, nil
}
