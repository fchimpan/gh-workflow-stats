package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/briandowns/spinner"
)

type workflowStatsResult struct {
	WorkflowRunsSummary WorkflowRunsSummary  `json:"workflow_runs_summary"`
	JobSummary          []OutputStepsSummary `json:"job_summary"`
}

type Timestamp struct {
	time.Time
}

type ConclusionName string

type JobsName string

type JobSummary struct {
	Name               JobsName                  `json:"name"`
	TotalCount         int                       `json:"total_count"`
	Conclusions        map[ConclusionName]int    `json:"conclusion"`
	Rate               float64                   `json:"rate"`
	JobAverageDuration int64                     `json:"job_average_duration"`
	StepsSummary       map[JobsName]StepsSummary `json:"steps_summary"`
}

type OutputStepsSummary struct {
	Name               JobsName               `json:"name"`
	TotalCount         int                    `json:"total_count"`
	Conclusions        map[ConclusionName]int `json:"conclusion"`
	Rate               float64                `json:"rate"`
	StepsSummary       []StepsSummary         `json:"steps_summary"`
	JobAverageDuration int64                  `json:"job_average_duration"`
}

type StepsSummary struct {
	Name                string                 `json:"name"`
	Number              int64                  `json:"number"`
	Count               int                    `json:"count"`
	Conclusions         map[ConclusionName]int `json:"conclusion"`
	Rate                float64                `json:"rate"`
	StepAverageDuration float64                `json:"step_average_duration"`
	FailureHTMLURL      []string               `json:"failure_html_url"`
}

type WorkflowRunsSummary struct {
	ID          int64       `json:"id,omitempty"`
	WorkflowID  int64       `json:"workflow_id,omitempty"`
	Name        string      `json:"name,omitempty"`
	TotalCount  int         `json:"total_count"`
	Count       int         `json:"count"`
	Rate        float64     `json:"rate"`
	Conclusions Conclusions `json:"conclusions"`
}

type Conclusions map[ConclusionName]*RunsConclusion

type RunsConclusion struct {
	Count        int           `json:"count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

type WorkflowRun struct {
	ID           int64     `json:"id,omitempty"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	RunAttempt   int       `json:"run_attempt"`
	HTMLURL      string    `json:"html_url"`
	URL          string    `json:"url"`
	RunStartedAt time.Time `json:"run_started_at"`
	JobsURL      string    `json:"jobs_url"`
}

func workflowStats(cfg *workflowStatsConfig, s *spinner.Spinner) error {
	ctx := context.Background()
	client, err := newGithubClient(cfg.RepoInfo.Host)
	if err != nil {
		return err
	}

	s.Start()

	runs, err := client.fetchWorkflowRuns(ctx, cfg)
	if err != nil {
		return err
	}

	jobs, err := client.fetchWorkflowJobsAttempts(ctx, runs, cfg)
	if err != nil {
		return err
	}

	s.Stop()

	c := &Conclusions{}
	wrs := &WorkflowRunsSummary{
		ID:          runs[0].GetID(),
		WorkflowID:  runs[0].GetWorkflowID(),
		Name:        runs[0].GetName(),
		Count:       len(runs),
		Conclusions: *c,
	}

	for _, run := range runs {
		cn := ConclusionName(run.GetConclusion())
		if cn == "" {
			cn = "others"
		}
		rc := wrs.Conclusions[cn]
		if rc == nil {
			rc = &RunsConclusion{}
			wrs.Conclusions[cn] = rc
		}
		rc.Count++
		rc.WorkflowRuns = append(rc.WorkflowRuns, WorkflowRun{
			ID:           run.GetID(),
			Status:       run.GetStatus(),
			Conclusion:   run.GetConclusion(),
			RunAttempt:   run.GetRunAttempt(),
			HTMLURL:      run.GetHTMLURL(),
			URL:          run.GetURL(),
			RunStartedAt: run.GetCreatedAt().Time,
			JobsURL:      run.GetJobsURL(),
		})
		wrs.TotalCount += run.GetRunAttempt()
		wrs.Conclusions[ConclusionName(run.GetConclusion())] = rc
	}
	tc := wrs.TotalCount
	for k, v := range wrs.Conclusions {
		if k != "success" && k != "failure" {
			tc -= v.Count
		}
	}
	if wrs.Conclusions["success"] != nil {
		wrs.Rate = float64(wrs.Conclusions["success"].Count) / max(float64(tc), 1)
	}

	summary := make(map[JobsName]*JobSummary)
	for _, j := range jobs {
		sm := summary[JobsName(j.GetName())]
		if sm == nil {
			summary[JobsName(j.GetName())] = &JobSummary{
				Conclusions:  make(map[ConclusionName]int),
				StepsSummary: make(map[JobsName]StepsSummary),
			}
			sm = summary[JobsName(j.GetName())]
		}
		sm.TotalCount++
		cn := ConclusionName(j.GetConclusion())
		if cn == "" {
			cn = "others"
		}
		sm.Conclusions[cn]++
		if cn == "success" && j.GetStatus() == "completed" {
			sm.JobAverageDuration = (sm.JobAverageDuration + int64(math.Abs(j.GetCompletedAt().Time.Sub(j.GetStartedAt().Time).Seconds())))
		}

		for _, s := range j.Steps {
			ss, ok := sm.StepsSummary[JobsName(s.GetName())]
			if !ok {
				ss = StepsSummary{
					Conclusions:    make(map[ConclusionName]int),
					Name:           s.GetName(),
					Number:         s.GetNumber(),
					FailureHTMLURL: []string{},
				}
				sm.StepsSummary[JobsName(s.GetName())] = ss
			}
			c := s.GetConclusion()
			if c == "" {
				c = "others"
			}
			ss.Conclusions[ConclusionName(c)]++
			ss.Count++
			if c == "success" && s.GetStatus() == "completed" {
				ss.StepAverageDuration = (ss.StepAverageDuration + float64(math.Abs(s.GetCompletedAt().Time.Sub(s.GetStartedAt().Time).Seconds())))
			}
			if c == "failure" && s.GetStatus() == "completed" {
				ss.FailureHTMLURL = append(ss.FailureHTMLURL, j.GetHTMLURL())
			}
			sm.StepsSummary[JobsName(s.GetName())] = ss
		}
	}

	var keys []JobsName
	for k := range summary {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})

	jobResults := make([]OutputStepsSummary, 0, len(summary))
	for _, k := range keys {
		v := summary[k]
		v.JobAverageDuration = v.JobAverageDuration / max(int64(v.Conclusions["success"]), 1)
		tmpss := make([]StepsSummary, 0, len(v.StepsSummary))
		for kk, ss := range v.StepsSummary {
			tmp := ss.StepAverageDuration
			ss.StepAverageDuration = tmp / max(float64(ss.Conclusions["success"]), 1)
			ss.Rate = float64(ss.Conclusions["success"]) / max(float64(ss.Count-ss.Conclusions["skipped"]), 1)
			v.StepsSummary[kk] = ss
			tmpss = append(tmpss, ss)
		}

		jobResults = append(jobResults, OutputStepsSummary{
			Name:               k,
			TotalCount:         v.TotalCount,
			Conclusions:        v.Conclusions,
			Rate:               float64(v.Conclusions["success"]) / max(float64(v.TotalCount-v.Conclusions["skipped"]), 1),
			JobAverageDuration: v.JobAverageDuration,
			StepsSummary:       tmpss,
		})
	}

	res := &workflowStatsResult{
		WorkflowRunsSummary: *wrs,
		JobSummary:          jobResults,
	}

	bytes, err := json.MarshalIndent(res, "", "	")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))

	return nil
}
