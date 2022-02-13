package github

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/spi"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

type GitHub struct {
	ApiURL      *url.URL
	AccessToken string
	restClient  *github.Client
}

func New(opts GitHub) *GitHub {
	i := GitHub{
		ApiURL:      opts.ApiURL,
		AccessToken: opts.AccessToken,
	}

	i.init()
	return &i
}

func (c *GitHub) init() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.AccessToken},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	githubClient := github.NewClient(tc)
	githubClient.BaseURL = c.ApiURL

	c.restClient = githubClient
}

func (c GitHub) CreatePullRequest(ctx context.Context,
	remoteURL, fromBranch, toBranch string,
	title, body string) (string, error) {

	user, resp, err := c.restClient.Users.Get(ctx, "")
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

	pr, resp, err := c.restClient.PullRequests.Create(ctx, owner, repoName, &npr)
	if err != nil {
		return "", err
	}

	if resp.Status != "201 Created" {
		return "", fmt.Errorf("failed to create PR: %v", resp)
	}

	return *pr.HTMLURL, nil
}

func (c GitHub) ListPullRequests(ctx context.Context, remoteURL string, user string) ([]spi.PullRequest, error) {
	owner, repoName := getRepo(remoteURL)
	prs, resp, err := c.restClient.PullRequests.List(ctx, owner, repoName, nil)
	if err != nil {
		return nil, err
	}

	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("failed to list PRs: %v", resp)
	}

	result := make([]spi.PullRequest, len(prs))
	for _, pr := range prs {
		if *pr.User.Name == user {
			result = append(result, spi.PullRequest{
				Repo:       repoName,
				Owner:      owner,
				URL:        *pr.HTMLURL,
				FromBranch: *pr.Head.Ref,
				ToBranch:   *pr.Base.Ref,
			})
		}
	}

	return result, nil
}

func getRepo(repoURL string) (string, string) {
	urlSplits := strings.Split(repoURL, ":")
	repo := strings.TrimSuffix(urlSplits[1], ".git")
	repoSplits := strings.Split(repo, "/")

	return repoSplits[0], repoSplits[1]
}
