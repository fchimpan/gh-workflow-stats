package github

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v60/github"
)

type Authenticator interface {
	AuthTokenForHost(host string) (string, error)
}

type WorkflowStatsClient struct {
	client        *github.Client
	authenticator Authenticator
}

type GitHubAuthenticator struct{}

func (ga *GitHubAuthenticator) AuthTokenForHost(host string) (string, error) {
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", fmt.Errorf("gh auth token not found for host %s", host)
	}
	return token, nil
}

func NewClient(host string, authenticator Authenticator) (*WorkflowStatsClient, error) {
	token, err := authenticator.AuthTokenForHost(host)
	if err != nil {
		return nil, err
	}
	r, err := github_ratelimit.NewRateLimitWaiterClient(nil)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(r).WithAuthToken(token)
	if host != "github.com" {
		client.BaseURL.Host = host
		client.BaseURL.Path = "/api/v3/"
	}
	return &WorkflowStatsClient{client: client, authenticator: authenticator}, nil
}
