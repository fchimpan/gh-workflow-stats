package actions

import "time"

type ConclusionName string

type WorkflowRuns struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	WorkflowRunsCount      int    `json:"workflow_runs_count"`
	TotalWorkflowRunsCount int    `json:"total_workflow_runs_count"`
	*WorkflowRunsConclusions
}

type WorkflowRunsConclusions map[ConclusionName]*RunsConclusion

type RunsConclusion struct {
	Count        int64 `json:"count"`
	WorkflowRuns []WorkflowRun
}

type WorkflowRun struct {
	Conclusion string    `json:"conclusion"`
	Status     string    `json:"status"`
	RunAttempt int64     `json:"run_attempt"`
	URL        string    `json:"url"`
	HTMLURL    string    `json:"html_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	JobsURL    string    `json:"jobs_url"`
	Actor      string    `json:"actor"`
}

type WorkflowJob struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Conclusions JobsConclusions
}

type JobsConclusions struct {
	TotalCount int64 `json:"total_count"`
	Success    JobsConclusion
	Failure    JobsConclusion
	Cancelled  JobsConclusion
	Skipped    JobsConclusion
	Others     JobsConclusion
}

type JobsConclusion struct {
	Count int64 `json:"count"`
	Jobs  []struct {
		Name        string    `json:"name"`
		Conclusion  string    `json:"conclusion"`
		Status      string    `json:"status"`
		RunAttempt  int64     `json:"run_attempt"`
		URL         string    `json:"url"`
		HTMLURL     string    `json:"html_url"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Actor       string    `json:"actor"`
		StartedAt   time.Time `json:"started_at"`
		CompletedAt time.Time `json:"completed_at"`
		Steps       []struct {
			Name        string    `json:"name"`
			Status      string    `json:"status"`
			Conclusion  string    `json:"conclusion"`
			Number      int64     `json:"number"`
			StartedAt   time.Time `json:"started_at"`
			CompletedAt time.Time `json:"completed_at"`
		} `json:"steps"`
	} `json:"jobs"`
}
