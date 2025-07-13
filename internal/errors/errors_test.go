package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "Error without cause",
			appError: &AppError{
				Type:    "TestError",
				Message: "test message",
			},
			expected: "TestError: test message",
		},
		{
			name: "Error with cause",
			appError: &AppError{
				Type:    "TestError",
				Message: "test message",
				Cause:   errors.New("underlying error"),
			},
			expected: "TestError: test message (caused by: underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	appError := &AppError{
		Type:    "TestError",
		Message: "test message",
		Cause:   cause,
	}

	assert.Equal(t, cause, appError.Unwrap())
}

func TestAppError_WithContext(t *testing.T) {
	appError := &AppError{
		Type:    "TestError",
		Message: "test message",
	}

	result := appError.WithContext("key", "value")
	assert.Equal(t, appError, result) // Should return same instance
	assert.Equal(t, "value", appError.Context["key"])

	// Add another context
	appError.WithContext("another", "context")
	assert.Equal(t, "context", appError.Context["another"])
	assert.Equal(t, "value", appError.Context["key"]) // Previous context should remain
}

func TestAppError_WithOperation(t *testing.T) {
	appError := &AppError{
		Type:    "TestError",
		Message: "test message",
	}

	result := appError.WithOperation("test_operation")
	assert.Equal(t, appError, result) // Should return same instance
	assert.Equal(t, "test_operation", appError.Operation)
}

func TestNewConfigurationError(t *testing.T) {
	cause := errors.New("config error")
	err := NewConfigurationError("invalid config", cause)

	assert.Equal(t, "ConfigurationError", err.Type)
	assert.Equal(t, "invalid config", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.False(t, err.Retryable)
}

func TestNewGitHubAPIError(t *testing.T) {
	cause := errors.New("api error")
	err := NewGitHubAPIError("API request failed", cause)

	assert.Equal(t, "GitHubAPIError", err.Type)
	assert.Equal(t, "API request failed", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, err.Retryable)
}

func TestNewRateLimitError(t *testing.T) {
	cause := errors.New("rate limit exceeded")
	err := NewRateLimitError("rate limited", cause)

	assert.Equal(t, "RateLimitError", err.Type)
	assert.Equal(t, "rate limited", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, err.Retryable)
}

func TestNewWorkflowError(t *testing.T) {
	cause := errors.New("workflow error")
	err := NewWorkflowError("workflow not found", cause)

	assert.Equal(t, "WorkflowError", err.Type)
	assert.Equal(t, "workflow not found", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.False(t, err.Retryable)
}

func TestNewDataProcessingError(t *testing.T) {
	cause := errors.New("processing error")
	err := NewDataProcessingError("failed to process data", cause)

	assert.Equal(t, "DataProcessingError", err.Type)
	assert.Equal(t, "failed to process data", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.False(t, err.Retryable)
}

func TestNewSystemError(t *testing.T) {
	cause := errors.New("system error")
	err := NewSystemError("system failure", cause)

	assert.Equal(t, "SystemError", err.Type)
	assert.Equal(t, "system failure", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, err.Retryable)
}

func TestIsConfigurationError(t *testing.T) {
	configErr := NewConfigurationError("config error", nil)
	apiErr := NewGitHubAPIError("api error", nil)
	stdErr := errors.New("standard error")

	assert.True(t, IsConfigurationError(configErr))
	assert.False(t, IsConfigurationError(apiErr))
	assert.False(t, IsConfigurationError(stdErr))
}

func TestIsGitHubAPIError(t *testing.T) {
	configErr := NewConfigurationError("config error", nil)
	apiErr := NewGitHubAPIError("api error", nil)
	stdErr := errors.New("standard error")

	assert.False(t, IsGitHubAPIError(configErr))
	assert.True(t, IsGitHubAPIError(apiErr))
	assert.False(t, IsGitHubAPIError(stdErr))
}

func TestIsRateLimitError(t *testing.T) {
	rateLimitErr := NewRateLimitError("rate limit", nil)
	apiErr := NewGitHubAPIError("api error", nil)
	stdErr := errors.New("standard error")

	assert.True(t, IsRateLimitError(rateLimitErr))
	assert.False(t, IsRateLimitError(apiErr))
	assert.False(t, IsRateLimitError(stdErr))
}

func TestIsRetryableError(t *testing.T) {
	retryableErr := NewGitHubAPIError("api error", nil)
	nonRetryableErr := NewConfigurationError("config error", nil)
	stdErr := errors.New("standard error")

	assert.True(t, IsRetryableError(retryableErr))
	assert.False(t, IsRetryableError(nonRetryableErr))
	assert.False(t, IsRetryableError(stdErr))
}

func TestWrapWithContext(t *testing.T) {
	originalErr := errors.New("original error")
	context := map[string]string{
		"operation": "test",
		"resource":  "workflow",
	}

	wrappedErr := WrapWithContext(originalErr, "operation failed", context)

	assert.NotNil(t, wrappedErr)

	var appErr *AppError
	assert.True(t, errors.As(wrappedErr, &appErr))
	assert.Equal(t, "WrappedError", appErr.Type)
	assert.Equal(t, "operation failed", appErr.Message)
	assert.Equal(t, originalErr, appErr.Cause)
	assert.Equal(t, context, appErr.Context)
	assert.False(t, appErr.Retryable)
}

func TestWrapWithContext_RetryableError(t *testing.T) {
	retryableErr := NewGitHubAPIError("api error", nil)
	context := map[string]string{"resource": "workflow"}

	wrappedErr := WrapWithContext(retryableErr, "operation failed", context)

	var appErr *AppError
	assert.True(t, errors.As(wrappedErr, &appErr))
	assert.True(t, appErr.Retryable) // Should inherit retryable status
}

func TestWrapWithContext_NilError(t *testing.T) {
	result := WrapWithContext(nil, "message", nil)
	assert.Nil(t, result)
}

func TestErrorConstants(t *testing.T) {
	// Just verify the constants exist and are not empty
	assert.NotEmpty(t, ErrMissingConfiguration.Error())
	assert.NotEmpty(t, ErrInvalidConfiguration.Error())
	assert.NotEmpty(t, ErrGitHubAPIAccess.Error())
	assert.NotEmpty(t, ErrGitHubRateLimit.Error())
	assert.NotEmpty(t, ErrGitHubAuthentication.Error())
	assert.NotEmpty(t, ErrGitHubResourceNotFound.Error())
	assert.NotEmpty(t, ErrWorkflowNotFound.Error())
	assert.NotEmpty(t, ErrWorkflowRunsEmpty.Error())
	assert.NotEmpty(t, ErrWorkflowJobsEmpty.Error())
	assert.NotEmpty(t, ErrDataProcessing.Error())
	assert.NotEmpty(t, ErrDataParsing.Error())
	assert.NotEmpty(t, ErrDataValidation.Error())
	assert.NotEmpty(t, ErrFileSystem.Error())
	assert.NotEmpty(t, ErrNetwork.Error())
	assert.NotEmpty(t, ErrTimeout.Error())
}
