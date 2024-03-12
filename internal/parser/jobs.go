package parser

import (
	"sort"

	"github.com/google/go-github/v60/github"
)

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

func WorkflowJobsParse(wjs []*github.WorkflowJob) []*WorkflowJobsStatsSummary {
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
		if c != ConclusionSuccess && c != ConclusionFailure {
			c = ConclusionOthers
		}
		w.Conclusions[c]++
		if wj.GetStatus() == "completed" && c == ConclusionSuccess {
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
			if c != ConclusionSuccess && c != ConclusionFailure {
				c = ConclusionOthers
			}
			ss.Conclusions[c]++
			if s.GetStatus() == "completed" && s.GetConclusion() == ConclusionFailure {
				ss.FailureHTMLURL = append(ss.FailureHTMLURL, wj.GetHTMLURL())
			}
			if s.GetStatus() == "completed" && (c == ConclusionSuccess || c == ConclusionFailure) {
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

		wjs.Rate.SuccesRate = adjustRate(float64(w.Conclusions[ConclusionSuccess]) / max(float64(w.TotalRunsCount), 1))
		wjs.Rate.FailureRate = adjustRate(float64(w.Conclusions[ConclusionFailure]) / max(float64(w.TotalRunsCount), 1))
		wjs.Rate.OthersRate = adjustRate(max(float64(1-wjs.Rate.SuccesRate-wjs.Rate.FailureRate), 0))
		res = append(res, wjs)
	}
	for _, r := range res {
		for _, s := range r.StepSummary {
			s.Rate.SuccesRate = adjustRate(float64(s.Conclusions[ConclusionSuccess]) / max(float64(s.RunsCount), 1))
			s.Rate.FailureRate = adjustRate(float64(s.Conclusions[ConclusionFailure]) / max(float64(s.RunsCount), 1))
			s.Rate.OthersRate = adjustRate(max(float64(1-s.Rate.SuccesRate-s.Rate.FailureRate), 0))
		}

		sort.Slice(r.StepSummary, func(i, j int) bool {
			return r.StepSummary[i].Number < r.StepSummary[j].Number
		})
	}

	return res
}
