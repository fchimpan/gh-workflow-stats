package parser

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowJobsParse(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want []*WorkflowJobsStatsSummary
	}{
		{
			name: "Empty",
			args: args{
				file: "empty.json",
			},
			want: []*WorkflowJobsStatsSummary{},
		},
		{
			name: "Multiple success",
			args: args{
				file: "multiple-success.json",
			},
			want: []*WorkflowJobsStatsSummary{
				{
					Name:           "test",
					TotalRunsCount: 2,
					Rate: Rate{
						SuccesRate:  1,
						FailureRate: 0,
						OthersRate:  0,
					},
					Conclusions: map[string]int{
						"success": 2,
						"failure": 0,
						"others":  0,
					},
					ExecutionDurationStats: ExecutionDurationStats{
						Min: 60,
						Max: 120,
						Avg: 90,
						Std: 30,
						Med: 90,
					},
					StepSummary: []*StepSummary{
						{
							Name:      "Set up job",
							Number:    1,
							RunsCount: 2,
							Conclusions: map[string]int{
								"success": 2,
								"failure": 0,
								"others":  0,
							},
							Rate: Rate{
								SuccesRate:  1,
								FailureRate: 0,
								OthersRate:  0,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 10,
								Max: 20,
								Avg: 15,
								Std: 5,
								Med: 15,
							},
							FailureHTMLURL: []string{},
						},
						{
							Name:      "Run test",
							Number:    2,
							RunsCount: 2,
							Conclusions: map[string]int{
								"success": 2,
								"failure": 0,
								"others":  0,
							},
							Rate: Rate{
								SuccesRate:  1,
								FailureRate: 0,
								OthersRate:  0,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 50,
								Max: 100,
								Avg: 75,
								Std: 25,
								Med: 75,
							},
							FailureHTMLURL: []string{},
						},
						{
							Name:      "Complete job",
							Number:    3,
							RunsCount: 2,
							Conclusions: map[string]int{
								"success": 2,
								"failure": 0,
								"others":  0,
							},
							Rate: Rate{
								SuccesRate:  1,
								FailureRate: 0,
								OthersRate:  0,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 0,
								Max: 0,
								Avg: 0,
								Std: 0,
								Med: 0,
							},

							FailureHTMLURL: []string{},
						},
					},
				},
			},
		},
		{
			name: "Multiple jobs",
			args: args{
				file: "multiple-jobs.json",
			},
			want: []*WorkflowJobsStatsSummary{
				{
					Name:           "test",
					TotalRunsCount: 3,
					Rate: Rate{
						SuccesRate:  0.3333333333333333,
						FailureRate: 0.3333333333333333,
						OthersRate:  0.3333333333333334,
					},
					Conclusions: map[string]int{
						"success": 1,
						"failure": 1,
						"others":  1,
					},
					ExecutionDurationStats: ExecutionDurationStats{
						Min: 60,
						Max: 60,
						Avg: 60,
						Std: 0,
						Med: 60,
					},
					StepSummary: []*StepSummary{
						{
							Name:      "Set up job",
							Number:    1,
							RunsCount: 3,
							Conclusions: map[string]int{
								"success": 3,
								"failure": 0,
								"others":  0,
							},
							Rate: Rate{
								SuccesRate:  1,
								FailureRate: 0,
								OthersRate:  0,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 10,
								Max: 20,
								Avg: 16.666666666666668,
								Std: 4.714045207910316,
								Med: 20,
							},
							FailureHTMLURL: []string{},
						},
						{
							Name:      "Run test",
							Number:    2,
							RunsCount: 3,
							Conclusions: map[string]int{
								"success": 1,
								"failure": 1,
								"others":  1,
							},
							Rate: Rate{
								SuccesRate:  0.3333333333333333,
								FailureRate: 0.3333333333333333,
								OthersRate:  0.3333333333333334,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 50,
								Max: 100,
								Avg: 75,
								Std: 25,
								Med: 75,
							},
							FailureHTMLURL: []string{
								"https://github.com/owner/repo/actions/runs/10002/job/2",
							},
						},
						{
							Name:      "Complete job",
							Number:    3,
							RunsCount: 3,
							Conclusions: map[string]int{
								"success": 3,
								"failure": 0,
								"others":  0,
							},
							Rate: Rate{
								SuccesRate:  1,
								FailureRate: 0,
								OthersRate:  0,
							},
							ExecutionDurationStats: ExecutionDurationStats{
								Min: 0,
								Max: 0,
								Avg: 0,
								Std: 0,
								Med: 0,
							},

							FailureHTMLURL: []string{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.ReadFile("testdata/jobs/" + tt.args.file)
			if err != nil {
				t.Fatal(err)
			}
			var wjs []*github.WorkflowJob
			if err := json.Unmarshal(d, &wjs); err != nil {
				t.Fatal(err)
			}
			got := WorkflowJobsParse(wjs)
			assert.Equal(t, tt.want, got)

		})
	}
}
