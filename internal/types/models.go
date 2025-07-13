package types

import (
	"time"
)

// WorkflowConclusion represents workflow execution conclusions
type WorkflowConclusion string

const (
	ConclusionSuccess WorkflowConclusion = "success"
	ConclusionFailure WorkflowConclusion = "failure"
	ConclusionOthers  WorkflowConclusion = "others"
)

// String returns the string representation of WorkflowConclusion
func (c WorkflowConclusion) String() string {
	return string(c)
}

// WorkflowStatus represents workflow execution status
type WorkflowStatus string

const (
	StatusCompleted   WorkflowStatus = "completed"
	StatusInProgress  WorkflowStatus = "in_progress"
	StatusQueued      WorkflowStatus = "queued"
	StatusRequested   WorkflowStatus = "requested"
	StatusWaiting     WorkflowStatus = "waiting"
	StatusPending     WorkflowStatus = "pending"
)

// ExecutionStats represents statistical data for execution durations
type ExecutionStats struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Mean    float64 `json:"mean"`
	Median  float64 `json:"median"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	StdDev  float64 `json:"std_dev"`
	Count   int     `json:"count"`
}

// SuccessRate represents success/failure rates
type SuccessRate struct {
	Success float64 `json:"success_rate"`
	Failure float64 `json:"failure_rate"`
	Others  float64 `json:"others_rate"`
}

// WorkflowRunSummary represents a processed workflow run
type WorkflowRunSummary struct {
	ID           int64              `json:"id"`
	Status       WorkflowStatus     `json:"status"`
	Conclusion   WorkflowConclusion `json:"conclusion"`
	Actor        string             `json:"actor"`
	RunAttempt   int                `json:"run_attempt"`
	HTMLURL      string             `json:"html_url"`
	JobsURL      string             `json:"jobs_url"`
	LogsURL      string             `json:"logs_url"`
	RunStartedAt time.Time          `json:"run_started_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	CreatedAt    time.Time          `json:"created_at"`
	Duration     float64            `json:"duration"`
}

// WorkflowRunStats represents aggregated statistics for workflow runs
type WorkflowRunStats struct {
	TotalCount     int                                       `json:"total_count"`
	Name           string                                    `json:"name"`
	SuccessRate    SuccessRate                               `json:"success_rate"`
	ExecutionStats ExecutionStats                            `json:"execution_stats"`
	Conclusions    map[WorkflowConclusion]*ConclusionSummary `json:"conclusions"`
}

// ConclusionSummary represents summary data for a specific conclusion
type ConclusionSummary struct {
	Count int                   `json:"count"`
	Runs  []*WorkflowRunSummary `json:"runs"`
}

// WorkflowJobStats represents aggregated statistics for workflow jobs
type WorkflowJobStats struct {
	Name           string                     `json:"name"`
	TotalCount     int                        `json:"total_count"`
	SuccessRate    SuccessRate                `json:"success_rate"`
	Conclusions    map[WorkflowConclusion]int `json:"conclusions"`
	ExecutionStats ExecutionStats             `json:"execution_stats"`
	Steps          []*StepStats               `json:"steps"`
}

// StepStats represents aggregated statistics for workflow steps
type StepStats struct {
	Name           string                     `json:"name"`
	Number         int64                      `json:"number"`
	RunCount       int                        `json:"run_count"`
	Conclusions    map[WorkflowConclusion]int `json:"conclusions"`
	SuccessRate    SuccessRate                `json:"success_rate"`
	ExecutionStats ExecutionStats             `json:"execution_stats"`
	FailureURLs    []string                   `json:"failure_urls,omitempty"`
}

// WorkflowAnalysisResult represents the complete analysis result
type WorkflowAnalysisResult struct {
	RunStats  *WorkflowRunStats   `json:"run_stats"`
	JobStats  []*WorkflowJobStats `json:"job_stats,omitempty"`
	Metadata  *AnalysisMetadata   `json:"metadata"`
}

// AnalysisMetadata contains metadata about the analysis
type AnalysisMetadata struct {
	GeneratedAt   time.Time          `json:"generated_at"`
	Config        *WorkflowConfig    `json:"config"`
	FetchOptions  *WorkflowFetchOptions `json:"fetch_options"`
	TotalFetched  int                `json:"total_fetched"`
	TotalFiltered int                `json:"total_filtered"`
	RateLimited   bool               `json:"rate_limited,omitempty"`
}

// IsValidConclusion returns true if the conclusion is valid
func IsValidConclusion(conclusion string) bool {
	switch WorkflowConclusion(conclusion) {
	case ConclusionSuccess, ConclusionFailure:
		return true
	default:
		return false
	}
}

// NormalizeConclusion converts a conclusion string to a standard conclusion
func NormalizeConclusion(conclusion string) WorkflowConclusion {
	switch WorkflowConclusion(conclusion) {
	case ConclusionSuccess:
		return ConclusionSuccess
	case ConclusionFailure:
		return ConclusionFailure
	default:
		return ConclusionOthers
	}
}

// IsValidStatus returns true if the status is valid
func IsValidStatus(status string) bool {
	switch WorkflowStatus(status) {
	case StatusCompleted, StatusInProgress, StatusQueued, StatusRequested, StatusWaiting, StatusPending:
		return true
	default:
		return false
	}
}