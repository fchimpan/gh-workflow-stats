package github

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/fchimpan/gh-workflow-stats/internal/errors"
	"github.com/fchimpan/gh-workflow-stats/internal/logger"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v60/github"
)

type Authenticator interface {
	AuthTokenForHost(host string) (string, error)
}

type WorkflowStatsClient struct {
	client        *github.Client
	authenticator Authenticator
	logger        logger.Logger
}

type GitHubAuthenticator struct{}

func (ga *GitHubAuthenticator) AuthTokenForHost(host string) (string, error) {
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", errors.NewConfigurationError(
			fmt.Sprintf("GitHub authentication token not found for host %s", host),
			nil,
		).WithContext("host", host).WithContext("help", "Run 'gh auth login' to authenticate")
	}
	return token, nil
}

func NewClient(host string, authenticator Authenticator, log logger.Logger) (*WorkflowStatsClient, error) {
	if log == nil {
		log = logger.NewNoOpLogger()
	}

	token, err := authenticator.AuthTokenForHost(host)
	if err != nil {
		LogError(log, err, "authentication", map[string]interface{}{
			"host": host,
		})
		return nil, err
	}

	r, err := github_ratelimit.NewRateLimitWaiterClient(nil)
	if err != nil {
		wrappedErr := errors.NewSystemError("failed to create rate limit client", err)
		LogError(log, wrappedErr, "client_creation", map[string]interface{}{
			"host": host,
		})
		return nil, wrappedErr
	}

	client := github.NewClient(r).WithAuthToken(token)
	if host != "github.com" {
		client.BaseURL.Host = host
		client.BaseURL.Path = "/api/v3/"
		log.Debug("configured GitHub Enterprise Server client", "host", host, "base_url", client.BaseURL.String())
	} else {
		log.Debug("configured GitHub.com client")
	}

	log.Info("GitHub client created successfully", "host", host)
	return &WorkflowStatsClient{
		client:        client,
		authenticator: authenticator,
		logger:        log,
	}, nil
}

// WithLogger returns a new client with the specified logger
func (c *WorkflowStatsClient) WithLogger(log logger.Logger) *WorkflowStatsClient {
	return &WorkflowStatsClient{
		client:        c.client,
		authenticator: c.authenticator,
		logger:        log,
	}
}

// handleHTTPError converts HTTP errors to structured errors
func (c *WorkflowStatsClient) handleHTTPError(resp *github.Response, err error, operation, resource string) error {
	if err == nil {
		return nil
	}

	// Handle rate limit errors
	if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		enhancedErr := NewRateLimitError(rateLimitErr, resource, operation)
		LogError(c.logger, enhancedErr, operation, map[string]interface{}{
			"resource":    resource,
			"limit":       enhancedErr.Limit,
			"remaining":   enhancedErr.Remaining,
			"reset_at":    enhancedErr.ResetAt,
			"retry_after": enhancedErr.RetryAfter,
		})
		return enhancedErr
	}

	// Handle other GitHub errors
	if resp != nil && resp.Response != nil {
		apiErr := NewAPIError(
			fmt.Sprintf("GitHub API request failed for %s", resource),
			resp.Response.StatusCode,
			err,
		).WithResource(resource).WithOperation(operation)

		if resp.Header.Get("X-GitHub-Request-Id") != "" {
			apiErr = apiErr.WithRequestID(resp.Header.Get("X-GitHub-Request-Id"))
		}

		LogError(c.logger, apiErr, operation, map[string]interface{}{
			"resource":    resource,
			"status_code": resp.StatusCode,
			"request_id":  resp.Header.Get("X-GitHub-Request-Id"),
		})
		return apiErr
	}

	// Handle generic errors
	wrappedErr := errors.NewSystemError(
		fmt.Sprintf("failed to %s for %s", operation, resource),
		err,
	)
	LogError(c.logger, wrappedErr, operation, map[string]interface{}{
		"resource": resource,
	})
	return wrappedErr
}
