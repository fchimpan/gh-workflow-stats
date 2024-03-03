package workflow

import (
	"fmt"

	"github.com/cli/go-gh/pkg/auth"
	"github.com/google/go-github/v57/github"
)

type workflowStatsClient struct {
	client *github.Client
}

func authTokenForHost(host string) (string, error) {
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", fmt.Errorf("gh auth token not found for host %s", host)
	}
	return token, nil
}

func newGithubClient(host string) (*workflowStatsClient, error) {
	token, err := authTokenForHost(host)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(token)
	if host != "github.com" {
		client.BaseURL.Host = host
		client.BaseURL.Path = "/api/v3/"
	}
	return &workflowStatsClient{client: client}, nil
}
