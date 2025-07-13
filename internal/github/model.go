package github

import (
	"fmt"
	"time"

	"github.com/fchimpan/gh-workflow-stats/internal/errors"
	"github.com/fchimpan/gh-workflow-stats/internal/logger"
	"github.com/google/go-github/v60/github"
)

// Enhanced RateLimitError with more context
type RateLimitError struct {
	Err        error
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
	Resource   string
	Operation  string
}

func (e RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("GitHub API rate limit exceeded for %s (limit: %d, remaining: %d, reset at: %s, retry after: %s): %v",
			e.Resource, e.Limit, e.Remaining, e.ResetAt.Format(time.RFC3339), e.RetryAfter, e.Err)
	}
	return fmt.Sprintf("GitHub API rate limit exceeded for %s (limit: %d, remaining: %d, reset at: %s): %v",
		e.Resource, e.Limit, e.Remaining, e.ResetAt.Format(time.RFC3339), e.Err)
}

func (e RateLimitError) Unwrap() error {
	return e.Err
}

// NewRateLimitError creates a RateLimitError from github.RateLimitError
func NewRateLimitError(gitHubErr *github.RateLimitError, resource, operation string) *RateLimitError {
	rateLimitErr := &RateLimitError{
		Err:       gitHubErr,
		Resource:  resource,
		Operation: operation,
	}

	if gitHubErr != nil {
		rateLimitErr.Limit = gitHubErr.Rate.Limit
		rateLimitErr.Remaining = gitHubErr.Rate.Remaining
		rateLimitErr.ResetAt = gitHubErr.Rate.Reset.UTC()

		// Calculate retry after duration
		now := time.Now()
		if gitHubErr.Rate.Reset.After(now) {
			rateLimitErr.RetryAfter = gitHubErr.Rate.Reset.Sub(now)
		}
	}

	return rateLimitErr
}

// APIError represents a structured GitHub API error
type APIError struct {
	*errors.AppError
	StatusCode int
	RequestID  string
	Resource   string
	Operation  string
}

func NewAPIError(message string, statusCode int, cause error) *APIError {
	return &APIError{
		AppError:   errors.NewGitHubAPIError(message, cause),
		StatusCode: statusCode,
	}
}

func (e *APIError) WithResource(resource string) *APIError {
	e.Resource = resource
	return e
}

func (e *APIError) WithOperation(operation string) *APIError {
	e.Operation = operation
	return e
}

func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// LogError logs an error with structured context
func LogError(logger logger.Logger, err error, operation string, context map[string]interface{}) {
	if err == nil {
		return
	}

	// Convert context to key-value pairs for slog
	var kvPairs []interface{}
	kvPairs = append(kvPairs, "operation", operation)

	for k, v := range context {
		kvPairs = append(kvPairs, k, v)
	}

	// Add error-specific context
	if rateLimitErr, ok := err.(*RateLimitError); ok {
		kvPairs = append(kvPairs,
			"error_type", "rate_limit",
			"limit", rateLimitErr.Limit,
			"remaining", rateLimitErr.Remaining,
			"reset_at", rateLimitErr.ResetAt,
			"retry_after", rateLimitErr.RetryAfter,
			"resource", rateLimitErr.Resource,
		)
	} else if apiErr, ok := err.(*APIError); ok {
		kvPairs = append(kvPairs,
			"error_type", "api_error",
			"status_code", apiErr.StatusCode,
			"request_id", apiErr.RequestID,
			"resource", apiErr.Resource,
		)
	} else if appErr, ok := err.(*errors.AppError); ok {
		kvPairs = append(kvPairs,
			"error_type", appErr.Type,
			"retryable", appErr.Retryable,
		)
		if appErr.Code != "" {
			kvPairs = append(kvPairs, "error_code", appErr.Code)
		}
		if len(appErr.Context) > 0 {
			for k, v := range appErr.Context {
				kvPairs = append(kvPairs, "context_"+k, v)
			}
		}
	}

	logger.Error(err.Error(), kvPairs...)
}
