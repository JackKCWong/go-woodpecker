package gitop

import "context"

type GitServer interface {
	CreatePullRequest(ctx context.Context, remoteURL, fromBranch, toBranch string) (string, error)
}
