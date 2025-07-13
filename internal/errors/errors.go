package errors

import (
	"errors"
	"fmt"
)

// Error types for consistent error handling
var (
	// Configuration errors
	ErrMissingConfiguration = errors.New("missing required configuration")
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// GitHub API errors
	ErrGitHubAPIAccess        = errors.New("GitHub API access error")
	ErrGitHubRateLimit        = errors.New("GitHub API rate limit exceeded")
	ErrGitHubAuthentication   = errors.New("GitHub authentication failed")
	ErrGitHubResourceNotFound = errors.New("GitHub resource not found")

	// Workflow errors
	ErrWorkflowNotFound  = errors.New("workflow not found")
	ErrWorkflowRunsEmpty = errors.New("no workflow runs found")
	ErrWorkflowJobsEmpty = errors.New("no workflow jobs found")

	// Data processing errors
	ErrDataProcessing = errors.New("data processing error")
	ErrDataParsing    = errors.New("data parsing error")
	ErrDataValidation = errors.New("data validation error")

	// System errors
	ErrFileSystem = errors.New("file system error")
	ErrNetwork    = errors.New("network error")
	ErrTimeout    = errors.New("operation timeout")
)

// AppError represents a structured application error
type AppError struct {
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Operation string            `json:"operation,omitempty"`
	Code      string            `json:"code,omitempty"`
	Cause     error             `json:"cause,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
	Retryable bool              `json:"retryable"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithContext(key, value string) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// Error creation helpers
func NewConfigurationError(message string, cause error) *AppError {
	return &AppError{
		Type:      "ConfigurationError",
		Message:   message,
		Cause:     cause,
		Retryable: false,
	}
}

func NewGitHubAPIError(message string, cause error) *AppError {
	return &AppError{
		Type:      "GitHubAPIError",
		Message:   message,
		Cause:     cause,
		Retryable: true,
	}
}

func NewRateLimitError(message string, cause error) *AppError {
	return &AppError{
		Type:      "RateLimitError",
		Message:   message,
		Cause:     cause,
		Retryable: true,
	}
}

func NewWorkflowError(message string, cause error) *AppError {
	return &AppError{
		Type:      "WorkflowError",
		Message:   message,
		Cause:     cause,
		Retryable: false,
	}
}

func NewDataProcessingError(message string, cause error) *AppError {
	return &AppError{
		Type:      "DataProcessingError",
		Message:   message,
		Cause:     cause,
		Retryable: false,
	}
}

func NewSystemError(message string, cause error) *AppError {
	return &AppError{
		Type:      "SystemError",
		Message:   message,
		Cause:     cause,
		Retryable: true,
	}
}

// Error checking helpers
func IsConfigurationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == "ConfigurationError"
}

func IsGitHubAPIError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == "GitHubAPIError"
}

func IsRateLimitError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == "RateLimitError"
}

func IsRetryableError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Retryable
}

// Wrap standard errors with context
func WrapWithContext(err error, message string, context map[string]string) error {
	if err == nil {
		return nil
	}

	appErr := &AppError{
		Type:      "WrappedError",
		Message:   message,
		Cause:     err,
		Context:   context,
		Retryable: false,
	}

	// If the wrapped error is retryable, make this one retryable too
	if IsRetryableError(err) {
		appErr.Retryable = true
	}

	return appErr
}
