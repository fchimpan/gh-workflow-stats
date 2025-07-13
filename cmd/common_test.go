package cmd

import (
	"os"
	"testing"

	"github.com/fchimpan/gh-workflow-stats/internal/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name     string
		org      string
		repo     string
		fileName string
		id       int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid flags with file name",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "ci.yml",
			id:       -1,
			wantErr:  false,
		},
		{
			name:     "Valid flags with workflow ID",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "",
			id:       12345,
			wantErr:  false,
		},
		{
			name:     "Valid flags with both file name and ID",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "ci.yml",
			id:       12345,
			wantErr:  false,
		},
		{
			name:     "Missing org",
			org:      "",
			repo:     "test-repo",
			fileName: "ci.yml",
			id:       -1,
			wantErr:  true,
			errMsg:   "ConfigurationError: " + ErrMissingOrgRepo,
		},
		{
			name:     "Missing repo",
			org:      "test-org",
			repo:     "",
			fileName: "ci.yml",
			id:       -1,
			wantErr:  true,
			errMsg:   "ConfigurationError: " + ErrMissingOrgRepo,
		},
		{
			name:     "Missing both org and repo",
			org:      "",
			repo:     "",
			fileName: "ci.yml",
			id:       -1,
			wantErr:  true,
			errMsg:   "ConfigurationError: " + ErrMissingOrgRepo,
		},
		{
			name:     "Missing both file name and ID",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "",
			id:       -1,
			wantErr:  true,
			errMsg:   "ConfigurationError: " + ErrMissingWorkflow,
		},
		{
			name:     "Empty file name and zero ID",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "",
			id:       0,
			wantErr:  true,
			errMsg:   "ConfigurationError: " + ErrMissingWorkflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(tt.org, tt.repo, tt.fileName, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolveHost(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		initialHost  string
		flagChanged  bool
		expectedHost string
	}{
		{
			name:         "No env var, flag not changed",
			envValue:     "",
			initialHost:  "github.com",
			flagChanged:  false,
			expectedHost: "github.com",
		},
		{
			name:         "Env var set, flag not changed",
			envValue:     "enterprise.github.com",
			initialHost:  "github.com",
			flagChanged:  false,
			expectedHost: "enterprise.github.com",
		},
		{
			name:         "Env var set, flag changed",
			envValue:     "enterprise.github.com",
			initialHost:  "custom.github.com",
			flagChanged:  true,
			expectedHost: "custom.github.com",
		},
		{
			name:         "No env var, flag changed",
			envValue:     "",
			initialHost:  "custom.github.com",
			flagChanged:  true,
			expectedHost: "custom.github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				_ = os.Setenv("GH_HOST", tt.envValue)
				defer func() { _ = os.Unsetenv("GH_HOST") }()
			} else {
				_ = os.Unsetenv("GH_HOST")
			}

			// Create a mock cobra command
			cmd := &cobra.Command{}
			cmd.Flags().String("host", "github.com", "GitHub host")
			if tt.flagChanged {
				_ = cmd.Flags().Set("host", tt.initialHost)
			}

			host := tt.initialHost
			resolveHost(cmd, &host)

			assert.Equal(t, tt.expectedHost, host)
		})
	}
}

func TestCreateConfig(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		org      string
		repo     string
		fileName string
		id       int64
		expected config
	}{
		{
			name:     "Config with file name",
			host:     "github.com",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "ci.yml",
			id:       -1,
			expected: config{
				host:             "github.com",
				org:              "test-org",
				repo:             "test-repo",
				workflowFileName: "ci.yml",
				workflowID:       -1,
			},
		},
		{
			name:     "Config with workflow ID",
			host:     "enterprise.github.com",
			org:      "test-org",
			repo:     "test-repo",
			fileName: "",
			id:       12345,
			expected: config{
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
			result := createConfig(tt.host, tt.org, tt.repo, tt.fileName, tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateOptions(t *testing.T) {
	tests := []struct {
		name                string
		actor               string
		branch              string
		event               string
		status              []string
		created             string
		headSHA             string
		excludePullRequests bool
		all                 bool
		js                  bool
		checkSuiteID        int64
		jobNum              int
		expected            options
	}{
		{
			name:                "Full options",
			actor:               "test-actor",
			branch:              "main",
			event:               "push",
			status:              []string{"completed", "in_progress"},
			created:             "2023-01-01",
			headSHA:             "abc123",
			excludePullRequests: true,
			all:                 true,
			js:                  true,
			checkSuiteID:        12345,
			jobNum:              5,
			expected: options{
				actor:               "test-actor",
				branch:              "main",
				event:               "push",
				status:              []string{"completed", "in_progress"},
				created:             "2023-01-01",
				headSHA:             "abc123",
				excludePullRequests: true,
				all:                 true,
				js:                  true,
				checkSuiteID:        12345,
				jobNum:              5,
			},
		},
		{
			name:     "Empty options",
			expected: options{jobNum: types.DefaultJobCount},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createOptions(tt.actor, tt.branch, tt.event, tt.status,
				tt.created, tt.headSHA, tt.excludePullRequests, tt.all, tt.js,
				tt.checkSuiteID, tt.jobNum)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "--org and --repo flag must be specified. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host with --host flag", ErrMissingOrgRepo)
	assert.Equal(t, "--file or --id flag must be specified", ErrMissingWorkflow)
	assert.Equal(t, 3, types.DefaultJobCount)
}
