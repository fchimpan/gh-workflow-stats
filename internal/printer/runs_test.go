package printer

import (
	"bytes"
	"testing"

	"github.com/fchimpan/gh-workflow-stats/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestRuns(t *testing.T) {
	type args struct {
		wrs *parser.WorkflowRunsStatsSummary
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name: "Empty",
			args: args{
				wrs: &parser.WorkflowRunsStatsSummary{
					TotalRunsCount: 0,
					Rate:           parser.Rate{},
					Conclusions: map[string]*parser.WorkflowRunsConclusion{
						parser.ConclusionSuccess: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
						parser.ConclusionFailure: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
						parser.ConclusionOthers: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
					},
				},
			},
			wantW: "🏃 Total runs: 0\n  ✔ Success: 0 (0.0%)\n  ✖ Failure: 0 (0.0%)\n  🤔 Others: 0 (0.0%)\n\n⏰ Workflow run execution time stats\n  Min: 0.0s\n  Max: 0.0s\n  Avg: 0.0s\n  Med: 0.0s\n  Std: 0.0s\n",
		},
		{
			name: "Success",
			args: args{
				wrs: &parser.WorkflowRunsStatsSummary{
					TotalRunsCount: 2,
					Name:           "CI",
					Rate: parser.Rate{
						SuccesRate:  1,
						FailureRate: 0.00001,
						OthersRate:  0,
					},
					ExecutionDurationStats: parser.ExecutionDurationStats{
						Min: 20.0,
						Max: 40.0,
						Avg: 30.0,
						Std: 10,
						Med: 30.0001,
					},
					Conclusions: map[string]*parser.WorkflowRunsConclusion{
						parser.ConclusionSuccess: {
							RunsCount: 2,
							WorkflowRuns: []*parser.WorkflowRun{
								{
									ID:     1,
									Status: "completed",
								},
								{
									ID:     2,
									Status: "completed",
								},
							},
						},
						parser.ConclusionFailure: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
						parser.ConclusionOthers: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
					},
				},
			},
			wantW: "🏃 Total runs: 2\n  ✔ Success: 2 (100.0%)\n  ✖ Failure: 0 (0.0%)\n  🤔 Others: 0 (0.0%)\n\n⏰ Workflow run execution time stats\n  Min: 20.0s\n  Max: 40.0s\n  Avg: 30.0s\n  Med: 30.0s\n  Std: 10.0s\n",
		},
		{
			name: "Mixed",
			args: args{
				wrs: &parser.WorkflowRunsStatsSummary{
					TotalRunsCount: 3,
					Name:           "CI",
					Rate: parser.Rate{
						SuccesRate:  0.6666666666666666,
						FailureRate: 0.3333333333333333,
						OthersRate:  0,
					},
					ExecutionDurationStats: parser.ExecutionDurationStats{
						Min: 20.0,
						Max: 40.0,
						Avg: 30.0,
						Std: 10,
						Med: 30.0001,
					},
					Conclusions: map[string]*parser.WorkflowRunsConclusion{
						parser.ConclusionSuccess: {
							RunsCount: 2,
							WorkflowRuns: []*parser.WorkflowRun{
								{
									ID:     1,
									Status: "completed",
								},
								{
									ID:     2,
									Status: "completed",
								},
							},
						},
						parser.ConclusionFailure: {
							RunsCount: 1,
							WorkflowRuns: []*parser.WorkflowRun{
								{
									ID:     3,
									Status: "completed",
								},
							},
						},
						parser.ConclusionOthers: {
							RunsCount:    0,
							WorkflowRuns: []*parser.WorkflowRun{},
						},
					},
				},
			},
			wantW: "🏃 Total runs: 3\n  ✔ Success: 2 (66.7%)\n  ✖ Failure: 1 (33.3%)\n  🤔 Others: 0 (0.0%)\n\n⏰ Workflow run execution time stats\n  Min: 20.0s\n  Max: 40.0s\n  Avg: 30.0s\n  Med: 30.0s\n  Std: 10.0s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			Runs(w, tt.args.wrs)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}
