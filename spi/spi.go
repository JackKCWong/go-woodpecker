package spi

import (
	"context"
	"io"
)

type GitClient interface {
	Clone(ctx context.Context, url string) error
	Origin() (string, error)
	Branch(name string) error
	Commit(msg string) (string, error)
	Push(ctx context.Context) error
}

type GitServer interface {
	CreatePullRequest(ctx context.Context, remoteURL, fromBranch, toBranch, title, body string) (string, error)
}

type PullRequest struct {
	URL        string
	Owner      string
	Repo       string
	FromBranch string
	ToBranch   string
}

type BuildTaskRunner interface {
	Run(ctx context.Context, task string, args ...string) (io.Reader, error)
	Wd() string
}
