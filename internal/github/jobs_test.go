package github

import (
	"context"
	"testing"

	"github.com/fchimpan/gh-workflow-stats/internal/logger"
	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
)

func TestFetchWorkflowJobsAttempts_EmptyRuns(t *testing.T) {
	// Test with empty runs list
	ctx := context.Background()
	log := logger.NewNoOpLogger()
	
	client := &WorkflowStatsClient{
		logger: log,
		client: github.NewClient(nil), // Initialize with nil http client for testing
	}
	
	cfg := &WorkflowRunsConfig{
		Org:  "test-org",
		Repo: "test-repo",
	}
	
	// Execute with empty runs
	jobs, err := client.FetchWorkflowJobsAttempts(ctx, []*github.WorkflowRun{}, cfg)
	
	// Assert
	assert.NoError(t, err)
	assert.Empty(t, jobs)
}

func TestFetchWorkflowJobsAttempts_NilHandling(t *testing.T) {
	// Test that nil values are handled properly
	ctx := context.Background()
	log := logger.NewNoOpLogger()
	
	client := &WorkflowStatsClient{
		logger: log,
		client: github.NewClient(nil),
	}
	
	// Test with nil config
	_, err := client.FetchWorkflowJobsAttempts(ctx, []*github.WorkflowRun{{ID: github.Int64(1)}}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
	
	// Test with nil context
	cfg := &WorkflowRunsConfig{
		Org:  "test-org",
		Repo: "test-repo",
	}
	_, err = client.FetchWorkflowJobsAttempts(nil, []*github.WorkflowRun{{ID: github.Int64(1)}}, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

// Test helper functions
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{
			name:     "a is smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b is smaller",
			a:        15,
			b:        10,
			expected: 10,
		},
		{
			name:     "equal values",
			a:        10,
			b:        10,
			expected: 10,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -10,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFetchWorkflowJobsAttempts_NilRunsInSlice(t *testing.T) {
	// Test that nil runs within the slice are handled properly
	ctx := context.Background()
	log := logger.NewNoOpLogger()
	
	client := &WorkflowStatsClient{
		logger: log,
		client: github.NewClient(nil),
	}
	
	cfg := &WorkflowRunsConfig{
		Org:  "test-org",
		Repo: "test-repo",
	}
	
	// Create slice with nil runs mixed in
	runs := []*github.WorkflowRun{
		{ID: github.Int64(1), RunAttempt: github.Int(1)},
		nil, // This should be handled gracefully
		{ID: github.Int64(2), RunAttempt: github.Int(1)},
		nil, // Another nil run
	}
	
	// We mainly want to ensure no panic occurs when processing nil runs
	assert.NotPanics(t, func() {
		client.FetchWorkflowJobsAttempts(ctx, runs, cfg)
	})
}

// Integration test that would require actual GitHub API or proper mocking
// This is a placeholder showing the structure of more comprehensive tests
func TestFetchWorkflowJobsAttempts_Integration(t *testing.T) {
	t.Skip("Integration test - requires GitHub API mock or test server")
	
	// This test would verify:
	// 1. Concurrent execution with semaphore
	// 2. Error handling for rate limits
	// 3. Proper aggregation of jobs from multiple runs
	// 4. Handling of 404 errors
	// 5. Context cancellation
}