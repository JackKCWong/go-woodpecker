package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

type GitHub struct {
	ApiURL      *url.URL
	AccessToken string
}

func (c GitHub) CreatePullRequest(ctx context.Context,
	remoteURL, fromBranch, toBranch string,
	title, body string) (string, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.AccessToken},
	)

	tc := oauth2.NewClient(ctx, ts)

	githubClient := github.NewClient(tc)
	githubClient.BaseURL = c.ApiURL
	user, resp, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}

	if resp.Status != "200 OK" {
		return "", fmt.Errorf("error authenticating user: %v", user)
	}

	npr := github.NewPullRequest{
		Title:               github.String(title),
		Body:                github.String(body),
		Head:                github.String(fromBranch),
		Base:                github.String(toBranch),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}

	owner, repoName := getRepo(remoteURL)

	pr, resp, err := githubClient.PullRequests.Create(ctx, owner, repoName, &npr)
	if err != nil {
		return "", err
	}

	if resp.Status != "201 Created" {
		return "", fmt.Errorf("failed to create PR: %v", resp)
	}

	return *pr.HTMLURL, nil
}

func getRepo(repoURL string) (string, string) {
	urlSplits := strings.Split(repoURL, ":")
	repo := strings.TrimSuffix(urlSplits[1], ".git")
	repoSplits := strings.Split(repo, "/")

	return repoSplits[0], repoSplits[1]
}