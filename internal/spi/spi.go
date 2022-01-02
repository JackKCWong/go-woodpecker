package spi

import (
	"context"
	"io"
)

type Git interface {
	Clone(ctx context.Context, url string) error
	Commit(msg string) error
	Push(ctx context.Context) error
	PullRequest(ctx context.Context) error
}

type TaskRunner interface {
	Run(ctx context.Context, task string, args ...string) (io.Reader, error)
	Wd() string
}
