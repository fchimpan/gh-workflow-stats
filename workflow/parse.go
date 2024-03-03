package workflow

import (
	"sort"
	"time"

	"github.com/google/go-github/v57/github"
)

const (
	conclusionSuccess = "success"
	conclusionFailure = "failure"
	conclusionOthers  = "others"
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
	RunStartedAt time.Time `json:"run_started_at"`
	Duration     float64   `json:"duration"`
}

func workflowRunsParse(wrs []*github.WorkflowRun) *WorkflowRunsStatsSummary {
	if len(wrs) == 0 {
		return &WorkflowRunsStatsSummary{
			TotalRunsCount: 0,
		}
	}
	wfrss := &WorkflowRunsStatsSummary{}
	durations := make([]float64, 0, len(wrs))
	for _, wr := range wrs {
		c := wr.GetConclusion()
		if c != conclusionSuccess && c != conclusionFailure {
			c = conclusionOthers
		}
		if wfrss.Conclusions == nil {
			wfrss.Conclusions = map[string]*WorkflowRunsConclusion{
				conclusionSuccess: {
					RunsCount:    0,
					WorkflowRuns: []*WorkflowRun{},
				},
				conclusionFailure: {
					RunsCount:    0,
					WorkflowRuns: []*WorkflowRun{},
				},
				conclusionOthers: {
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
			HTMLURL:      wr.GetHTMLURL(),
			RunStartedAt: wr.GetRunStartedAt().Time,
		}
		if c != conclusionOthers {
			d := wr.GetUpdatedAt().Sub(wr.GetRunStartedAt().Time).Seconds()
			w.Duration = d
			durations = append(durations, d)
		}
		wfrss.Conclusions[c].WorkflowRuns = append(wfrss.Conclusions[c].WorkflowRuns, &w)
	}
	wfrss.ExecutionDurationStats = calcStats(durations)

	wfrss.Rate.SuccesRate = float64(wfrss.Conclusions[conclusionSuccess].RunsCount) / max(float64(wfrss.TotalRunsCount), 1)
	wfrss.Rate.FailureRate = float64(wfrss.Conclusions[conclusionFailure].RunsCount) / max(float64(wfrss.TotalRunsCount), 1)
	wfrss.Rate.OthersRate = float64(1 - wfrss.Rate.SuccesRate - wfrss.Rate.FailureRate)

	return wfrss
}

type Result struct {
	WorkflowRunsStatsSummary *WorkflowRunsStatsSummary   `json:"workflow_runs_stats_summary"`
	WorkflowJobsStatsSummary []*WorkflowJobsStatsSummary `json:"workflow_jobs_stats_summary"`
}

type WorkflowJobsStatsSummary struct {
	Name                   string                 `json:"name"`
	TotalRunsCount         int                    `json:"total_runs_count"`
	Rate                   Rate                   `json:"rate"`
	Conclusions            map[string]int         `json:"conclusions"`
	ExecutionDurationStats executionDurationStats `json:"execution_duration_stats"`
	StepSummary            []*StepSummary         `json:"steps_summary"`
}

type StepSummary struct {
	Name                   string                 `json:"name"`
	Number                 int64                  `json:"number"`
	RunsCount              int                    `json:"runs_count"`
	Conclusions            map[string]int         `json:"conclusion"`
	Rate                   Rate                   `json:"rate"`
	ExecutionDurationStats executionDurationStats `json:"execution_duration_stats"`
	FailureHTMLURL         []string               `json:"failure_html_url"`
}

type WorkflowJobsStatsSummaryCalc struct {
	TotalRunsCount            int
	Name                      string
	Rate                      Rate
	Conclusions               map[string]int
	ExecutionWorkflowDuration []float64
	StepSummary               map[string]*StepSummaryCalc
}

type StepSummaryCalc struct {
	Name           string
	Number         int64
	RunsCount      int
	Conclusions    map[string]int
	StepDuration   []float64
	FailureHTMLURL []string
}

func workflowJobsParse(wjs []*github.WorkflowJob) []*WorkflowJobsStatsSummary {
	if len(wjs) == 0 {
		return []*WorkflowJobsStatsSummary{}
	}

	m := make(map[string]*WorkflowJobsStatsSummaryCalc)
	for _, wj := range wjs {
		if _, ok := m[wj.GetName()]; !ok {
			m[wj.GetName()] = &WorkflowJobsStatsSummaryCalc{
				Name:        wj.GetName(),
				StepSummary: make(map[string]*StepSummaryCalc),
				Conclusions: make(map[string]int),
			}
		}
		w := m[wj.GetName()]
		w.TotalRunsCount++
		c := wj.GetConclusion()
		if c != conclusionSuccess && c != conclusionFailure {
			c = conclusionOthers
		}
		w.Conclusions[c]++
		if wj.GetStatus() == "completed" && (c == conclusionSuccess || c == conclusionFailure) {
			d := wj.GetCompletedAt().Sub(wj.GetStartedAt().Time).Seconds()
			if d < 0 {
				d = 0
			}
			w.ExecutionWorkflowDuration = append(w.ExecutionWorkflowDuration, d)
		}

		for _, s := range wj.Steps {
			if _, ok := w.StepSummary[s.GetName()]; !ok {
				w.StepSummary[s.GetName()] = &StepSummaryCalc{
					Name:           s.GetName(),
					Number:         s.GetNumber(),
					Conclusions:    make(map[string]int),
					FailureHTMLURL: []string{},
				}
			}
			ss := w.StepSummary[s.GetName()]
			ss.RunsCount++
			c := s.GetConclusion()
			if c != conclusionSuccess && c != conclusionFailure {
				c = conclusionOthers
			}
			ss.Conclusions[c]++
			if s.GetStatus() == "completed" && s.GetConclusion() == conclusionFailure {
				ss.FailureHTMLURL = append(ss.FailureHTMLURL, wj.GetHTMLURL())
			}
			if s.GetStatus() == "completed" && (c == conclusionSuccess || c == conclusionFailure) {
				d := s.GetCompletedAt().Sub(s.GetStartedAt().Time).Seconds()
				if d < 0 {
					d = 0
				}
				ss.StepDuration = append(ss.StepDuration, d)
			}
			w.StepSummary[s.GetName()] = ss
		}
		m[wj.GetName()] = w
	}
	res := make([]*WorkflowJobsStatsSummary, 0, len(m))
	for _, w := range m {
		wjs := &WorkflowJobsStatsSummary{
			Name:                   w.Name,
			TotalRunsCount:         w.TotalRunsCount,
			Conclusions:            w.Conclusions,
			StepSummary:            make([]*StepSummary, 0, len(w.StepSummary)),
			ExecutionDurationStats: calcStats(w.ExecutionWorkflowDuration),
		}
		for _, ss := range w.StepSummary {
			wjs.StepSummary = append(wjs.StepSummary, &StepSummary{
				Name:                   ss.Name,
				Number:                 ss.Number,
				RunsCount:              ss.RunsCount,
				Conclusions:            ss.Conclusions,
				FailureHTMLURL:         ss.FailureHTMLURL,
				ExecutionDurationStats: calcStats(ss.StepDuration),
			})
		}

		wjs.Rate.SuccesRate = adjustRate(float64(w.Conclusions[conclusionSuccess]) / max(float64(w.TotalRunsCount), 1))
		wjs.Rate.FailureRate = adjustRate(float64(w.Conclusions[conclusionFailure]) / max(float64(w.TotalRunsCount), 1))
		wjs.Rate.OthersRate = adjustRate(max(float64(1-wjs.Rate.SuccesRate-wjs.Rate.FailureRate), 0))
		res = append(res, wjs)
	}
	for _, r := range res {
		for _, s := range r.StepSummary {
			s.Rate.SuccesRate = adjustRate(float64(s.Conclusions[conclusionSuccess]) / max(float64(s.RunsCount), 1))
			s.Rate.FailureRate = adjustRate(float64(s.Conclusions[conclusionFailure]) / max(float64(s.RunsCount), 1))
			s.Rate.OthersRate = adjustRate(max(float64(1-s.Rate.SuccesRate-s.Rate.FailureRate), 0))
		}

		sort.Slice(r.StepSummary, func(i, j int) bool {
			return r.StepSummary[i].Number < r.StepSummary[j].Number
		})
	}

	return res
}
