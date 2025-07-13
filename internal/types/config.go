package types

import "time"

// WorkflowConfig represents configuration for workflow operations
type WorkflowConfig struct {
	Host             string `json:"host"`
	Org              string `json:"org"`
	Repo             string `json:"repo"`
	WorkflowFileName string `json:"workflow_file_name,omitempty"`
	WorkflowID       int64  `json:"workflow_id,omitempty"`
}

// IsValid returns true if the configuration has the required fields
func (c *WorkflowConfig) IsValid() bool {
	return c.Org != "" && c.Repo != "" && (c.WorkflowFileName != "" || c.WorkflowID != 0)
}

// WorkflowFetchOptions represents options for fetching workflow data
type WorkflowFetchOptions struct {
	// Filter options
	Actor               string    `json:"actor,omitempty"`
	Branch              string    `json:"branch,omitempty"`
	Event               string    `json:"event,omitempty"`
	Status              []string  `json:"status,omitempty"`
	Created             string    `json:"created,omitempty"`
	HeadSHA             string    `json:"head_sha,omitempty"`
	ExcludePullRequests bool      `json:"exclude_pull_requests"`
	CheckSuiteID        int64     `json:"check_suite_id,omitempty"`
	All                 bool      `json:"all"`
	
	// Display options
	OutputJSON bool `json:"output_json"`
	JobCount   int  `json:"job_count"`
}

// OutputOptions represents options for output formatting
type OutputOptions struct {
	Format   OutputFormat `json:"format"`
	JobCount int          `json:"job_count"`
}

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
)

// APIRequestOptions represents options for GitHub API requests
type APIRequestOptions struct {
	Actor               string `json:"actor,omitempty"`
	Branch              string `json:"branch,omitempty"`
	Event               string `json:"event,omitempty"`
	Status              string `json:"status,omitempty"` // Single status for API
	Created             string `json:"created,omitempty"`
	HeadSHA             string `json:"head_sha,omitempty"`
	ExcludePullRequests bool   `json:"exclude_pull_requests"`
	CheckSuiteID        int64  `json:"check_suite_id,omitempty"`
	All                 bool   `json:"all"`
}

// ToAPIRequestOptions converts WorkflowFetchOptions to APIRequestOptions
func (o *WorkflowFetchOptions) ToAPIRequestOptions() *APIRequestOptions {
	var status string
	if len(o.Status) > 0 {
		status = o.Status[0] // API accepts only single status
	}
	
	return &APIRequestOptions{
		Actor:               o.Actor,
		Branch:              o.Branch,
		Event:               o.Event,
		Status:              status,
		Created:             o.Created,
		HeadSHA:             o.HeadSHA,
		ExcludePullRequests: o.ExcludePullRequests,
		CheckSuiteID:        o.CheckSuiteID,
		All:                 o.All,
	}
}

// TimestampRange represents a time range for filtering
type TimestampRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// Validate returns an error if the time range is invalid
func (tr *TimestampRange) Validate() error {
	if tr.Start != nil && tr.End != nil && tr.Start.After(*tr.End) {
		return ErrInvalidTimeRange
	}
	return nil
}