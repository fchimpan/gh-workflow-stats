package parser

import (
	"strconv"
	"time"

	"github.com/google/go-github/v60/github"
)

const (
	ConclusionSuccess = "success"
	ConclusionFailure = "failure"
	ConclusionOthers  = "others"
	StatusCompleted   = "completed"

	// Maximum duration of a workflow run is 35 days in Self-hosted runners.
	// ref: https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/about-self-hosted-runners#usage-limits
	MaxWorkflowDurationSeconds = 35 * 24 * 60 * 60 // 35 days in seconds
	MaxWorkflowDurationCapped  = 3024000           // Capped duration value
)

type WorkflowRunsStatsSummary struct {
	TotalRunsCount         int                                `json:"total_runs_count"`
	Name                   string                             `json:"name"`
	Rate                   Rate                               `json:"rate"`
	ExecutionDurationStats ExecutionDurationStats             `json:"execution_duration_stats"`
	Conclusions            map[string]*WorkflowRunsConclusion `json:"conclusions"`
}

type Rate struct {
	SuccesRate  float64 `json:"success_rate"`
	FailureRate float64 `json:"failure_rate"`
	OthersRate  float64 `json:"others_rate"`
}

type WorkflowRunsConclusion struct {
	RunsCount    int            `json:"runs_count"`
	WorkflowRuns []*WorkflowRun `json:"workflow_runs"`
}

type WorkflowRun struct {
	ID           int64     `json:"id,omitempty"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	Actor        string    `json:"actor"`
	RunAttempt   int       `json:"run_attempt"`
	HTMLURL      string    `json:"html_url"`
	JobsURL      string    `json:"jobs_url"`
	LogsURL      string    `json:"logs_url"`
	RunStartedAt time.Time `json:"run_started_at"`
	UpdateAt     time.Time `json:"update_at"`
	CreatedAt    time.Time `json:"created_at"`
	Duration     float64   `json:"duration"`
}

func WorkflowRunsParse(wrs []*github.WorkflowRun) *WorkflowRunsStatsSummary {
	wfrss := &WorkflowRunsStatsSummary{
		TotalRunsCount: 0,
		Conclusions: map[string]*WorkflowRunsConclusion{
			ConclusionSuccess: {
				RunsCount:    0,
				WorkflowRuns: []*WorkflowRun{},
			},
			ConclusionFailure: {
				RunsCount:    0,
				WorkflowRuns: []*WorkflowRun{},
			},
			ConclusionOthers: {
				RunsCount:    0,
				WorkflowRuns: []*WorkflowRun{},
			},
		},
	}
	if len(wrs) == 0 {
		return wfrss
	}
	wfrss.Name = wrs[0].GetName()

	durations := make([]float64, 0, len(wrs))
	for _, wr := range wrs {
		c := wr.GetConclusion()
		if c != ConclusionSuccess && c != ConclusionFailure {
			c = ConclusionOthers
		}

		wfrss.TotalRunsCount++
		wfrss.Conclusions[c].RunsCount++

		w := WorkflowRun{
			ID:           wr.GetID(),
			Status:       wr.GetStatus(),
			Conclusion:   wr.GetConclusion(),
			Actor:        wr.GetActor().GetLogin(),
			RunAttempt:   wr.GetRunAttempt(),
			HTMLURL:      wr.GetHTMLURL() + "/attempts/" + strconv.Itoa(wr.GetRunAttempt()),
			JobsURL:      wr.GetJobsURL(),
			LogsURL:      wr.GetLogsURL(),
			RunStartedAt: wr.GetRunStartedAt().Time.UTC(),
			UpdateAt:     wr.GetUpdatedAt().Time.UTC(),
			CreatedAt:    wr.GetCreatedAt().Time.UTC(),
		}
		// TODO: This is not the correct way to calculate the duration. https://github.com/fchimpan/gh-workflow-stats/issues/11
		d := wr.GetUpdatedAt().Sub(wr.GetRunStartedAt().Time).Seconds()
		if d > MaxWorkflowDurationSeconds {
			d = MaxWorkflowDurationCapped
		}
		w.Duration = d
		if c == ConclusionSuccess && d > 0 && wr.GetStatus() == StatusCompleted {
			durations = append(durations, d)
		}
		wfrss.Conclusions[c].WorkflowRuns = append(wfrss.Conclusions[c].WorkflowRuns, &w)
	}

	wfrss.ExecutionDurationStats = calcStats(durations)
	wfrss.Rate.SuccesRate = float64(wfrss.Conclusions[ConclusionSuccess].RunsCount) / max(float64(wfrss.TotalRunsCount), 1)
	wfrss.Rate.FailureRate = float64(wfrss.Conclusions[ConclusionFailure].RunsCount) / max(float64(wfrss.TotalRunsCount), 1)
	wfrss.Rate.OthersRate = float64(1 - wfrss.Rate.SuccesRate - wfrss.Rate.FailureRate)

	return wfrss
}
