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
)

type WorkflowRunsStatsSummary struct {
	TotalRunsCount         int                                `json:"total_runs_count"`
	Name                   string                             `json:"name"`
	Rate                   Rate                               `json:"rate"`
	ExecutionDurationStats executionDurationStats             `json:"execution_duration_stats"`
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
	Duration     float64   `json:"duration"`
}

func WorkflowRunsParse(wrs []*github.WorkflowRun) *WorkflowRunsStatsSummary {
	if len(wrs) == 0 {
		return &WorkflowRunsStatsSummary{
			TotalRunsCount: 0,
		}
	}
	wfrss := &WorkflowRunsStatsSummary{}
	durations := make([]float64, 0, len(wrs))
	for _, wr := range wrs {
		c := wr.GetConclusion()
		if c != ConclusionSuccess && c != ConclusionFailure {
			c = ConclusionOthers
		}
		if wfrss.Conclusions == nil {
			wfrss.Conclusions = map[string]*WorkflowRunsConclusion{
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
				}}
			wfrss.Name = wr.GetName()

		}
		if _, ok := wfrss.Conclusions[c]; !ok {
			wfrss.Conclusions[c] = &WorkflowRunsConclusion{
				RunsCount:    0,
				WorkflowRuns: []*WorkflowRun{},
			}
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
			RunStartedAt: wr.GetRunStartedAt().Time,
		}
		if c == ConclusionSuccess {
			d := wr.GetUpdatedAt().Sub(wr.GetRunStartedAt().Time).Seconds()
			w.Duration = d
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
