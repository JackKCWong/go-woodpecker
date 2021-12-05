package gitop

import (
	"context"
	"fmt"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	"strings"
)

type GitHub struct {
	AccessToken string
}

func (c GitHub) CreatePullRequest(ctx context.Context, remoteURL, fromBranch, toBranch string) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.AccessToken},
	)

	tc := oauth2.NewClient(ctx, ts)

	githubClient := github.NewClient(tc)
	user, resp, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		return err
	}

	if resp.Status != "200 OK" {
		return fmt.Errorf("error authenticating user: %v", user)
	}

	pr := github.NewPullRequest{
		Title:               github.String("auto update dependencies"),
		Body:                github.String("this request is created by Pacidae"),
		Head:                github.String(fromBranch),
		Base:                github.String(toBranch),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}

	owner, repoName := getRepo(remoteURL)

	_, resp, err = githubClient.PullRequests.Create(ctx, owner, repoName, &pr)
	if err != nil {
		return err
	}

	if resp.Status != "201 Created" {
		return fmt.Errorf("failed to create PR: %v", resp)
	}

	return nil
}

func getRepo(repoURL string) (string, string) {
	urlSplits := strings.Split(repoURL, ":")
	repo := strings.TrimSuffix(urlSplits[1], ".git")
	repoSplits := strings.Split(repo, "/")

	return repoSplits[0], repoSplits[1]
}