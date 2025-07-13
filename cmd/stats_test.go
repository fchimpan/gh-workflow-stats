package cmd

import (
	"testing"

	go_github "github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
)

func TestConfigStruct(t *testing.T) {
	tests := []struct {
		name   string
		config config
		want   config
	}{
		{
			name: "Config with file name",
			config: config{
				host:             "github.com",
				org:              "test-org",
				repo:             "test-repo",
				workflowFileName: "ci.yml",
				workflowID:       -1,
			},
			want: config{
				host:             "github.com",
				org:              "test-org",
				repo:             "test-repo",
				workflowFileName: "ci.yml",
				workflowID:       -1,
			},
		},
		{
			name: "Config with workflow ID",
			config: config{
				host:             "enterprise.github.com",
				org:              "test-org",
				repo:             "test-repo",
				workflowFileName: "",
				workflowID:       12345,
			},
			want: config{
				host:             "enterprise.github.com",
				org:              "test-org",
				repo:             "test-repo",
				workflowFileName: "",
				workflowID:       12345,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config)
		})
	}
}

func TestOptionsStruct(t *testing.T) {
	tests := []struct {
		name    string
		options options
		want    options
	}{
		{
			name: "Full options",
			options: options{
				actor:               "test-actor",
				branch:              "main",
				event:               "push",
				status:              []string{"completed", "in_progress"},
				created:             "2023-01-01",
				headSHA:             "abc123",
				excludePullRequests: true,
				checkSuiteID:        12345,
				all:                 true,
				js:                  true,
				jobNum:              5,
			},
			want: options{
				actor:               "test-actor",
				branch:              "main",
				event:               "push",
				status:              []string{"completed", "in_progress"},
				created:             "2023-01-01",
				headSHA:             "abc123",
				excludePullRequests: true,
				checkSuiteID:        12345,
				all:                 true,
				js:                  true,
				jobNum:              5,
			},
		},
		{
			name:    "Empty options",
			options: options{},
			want:    options{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.options)
		})
	}
}

func TestFilterRunAttemptsByStatus(t *testing.T) {
	// Create test WorkflowRuns
	statusCompleted := "completed"
	statusInProgress := "in_progress"
	conclusionSuccess := "success"
	conclusionFailure := "failure"

	runs := []*go_github.WorkflowRun{
		{Status: &statusCompleted, Conclusion: &conclusionSuccess},
		{Status: &statusInProgress, Conclusion: nil},
		{Status: &statusCompleted, Conclusion: &conclusionFailure},
	}

	tests := []struct {
		name          string
		runs          []*go_github.WorkflowRun
		status        []string
		expectedCount int
		description   string
	}{
		{
			name:          "Empty status slice returns all runs",
			runs:          runs,
			status:        []string{},
			expectedCount: 3,
			description:   "Empty status filter should return all runs",
		},
		{
			name:          "Single empty string returns all runs",
			runs:          runs,
			status:        []string{""},
			expectedCount: 3,
			description:   "Empty string status should return all runs",
		},
		{
			name:          "Filter by status completed",
			runs:          runs,
			status:        []string{"completed"},
			expectedCount: 2,
			description:   "Should return runs with completed status",
		},
		{
			name:          "Filter by conclusion success",
			runs:          runs,
			status:        []string{"success"},
			expectedCount: 1,
			description:   "Should return runs with success conclusion",
		},
		{
			name:          "Filter by multiple statuses",
			runs:          runs,
			status:        []string{"completed", "success"},
			expectedCount: 2,
			description:   "Should return runs matching either status or conclusion",
		},
		{
			name:          "Filter by non-existent status",
			runs:          runs,
			status:        []string{"queued"},
			expectedCount: 0,
			description:   "Should return no runs for non-matching status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterRunAttemptsByStatus(tt.runs, tt.status)
			assert.Len(t, result, tt.expectedCount, tt.description)
		})
	}
}

func TestStatsConstants(t *testing.T) {
	assert.Equal(t, "  fetching workflow runs...", workflowRunsText)
	assert.Equal(t, "  fetching workflow jobs...", workflowJobsText)
	assert.Equal(t, 14, charSize)
}
