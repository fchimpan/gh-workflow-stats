package printer

import (
	"bytes"
	"testing"

	"github.com/fchimpan/gh-workflow-stats/internal/parser"
)

func TestFailureJobs(t *testing.T) {
	type args struct {
		jobs []*parser.WorkflowJobsStatsSummary
		n    int
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name: "Empty",
			args: args{
				jobs: []*parser.WorkflowJobsStatsSummary{},
				n:    3,
			},
			wantW: "\n📈 Top 0 jobs with the highest failure counts (failure jobs / total runs)\n",
		},
		{
			name: "Positive",
			args: args{
				jobs: []*parser.WorkflowJobsStatsSummary{
					{
						Name:           "job1",
						Conclusions:    map[string]int{"failure": 5},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job1-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 5},
							},
							{
								Name:        "job1-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 3},
							},
						},
					},
					{
						Name:           "job2",
						Conclusions:    map[string]int{"failure": 2},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job2-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 2},
							},
							{
								Name:        "job2-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 1},
							},
						},
					},
				},
				n: 5,
			},
			wantW: "\n📈 Top 2 jobs with the highest failure counts (failure jobs / total runs)\n  job1: 5/10\n    └──job1-step1: 5/10\n\n  job2: 2/10\n    └──job2-step1: 2/10\n\n",
		},
		{
			name: "Multiple Jobs with the Same Number of Failures",
			args: args{
				jobs: []*parser.WorkflowJobsStatsSummary{
					{
						Name:           "job2",
						Conclusions:    map[string]int{"failure": 5},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job2-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 5},
							},
							{
								Name:        "job2-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 5},
							},
						},
					},
					{
						Name:           "job1",
						Conclusions:    map[string]int{"failure": 5},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job1-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 5},
							},
							{
								Name:        "job1-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 5},
							},
						},
					},
					{
						Name:           "job3",
						Conclusions:    map[string]int{"failure": 7},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job3-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 3},
							},
							{
								Name:        "job3-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 4},
							},
						},
					},
				},
				n: 5,
			},
			wantW: "\n📈 Top 3 jobs with the highest failure counts (failure jobs / total runs)\n  job3: 7/10\n    └──job3-step2: 4/10\n\n  job2: 5/10\n    └──job2-step1: 5/10\n\n  job1: 5/10\n    └──job1-step1: 5/10\n\n",
		},
		{
			name: "Failure step not found",
			args: args{
				jobs: []*parser.WorkflowJobsStatsSummary{
					{
						Name:           "job1",
						Conclusions:    map[string]int{"failure": 0},
						TotalRunsCount: 10,
						StepSummary: []*parser.StepSummary{
							{
								Name:        "job1-step1",
								Number:      1,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 0},
							},
							{
								Name:        "job1-step2",
								Number:      2,
								RunsCount:   10,
								Conclusions: map[string]int{"failure": 0},
							},
						},
					},
				},
				n: 5,
			},
			wantW: "\n📈 Top 1 jobs with the highest failure counts (failure jobs / total runs)\n  job1: 0/10\n    └──Failed step not found: 0/10\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			FailureJobs(w, tt.args.jobs, tt.args.n)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("FailureJobs() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
