package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/fatih/color"
	"github.com/fchimpan/gh-workflow-stats/internal/parser"
)

func FailureJobs(w io.Writer, jobs []*parser.WorkflowJobsStatsSummary, n int) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Conclusions[parser.ConclusionFailure] > jobs[j].Conclusions[parser.ConclusionFailure]
	})
	jobsNum := min(len(jobs), n)
	fmt.Fprintf(w, "\n%s Top %d jobs with the highest failure counts (failure jobs / total runs)\n", "\U0001F4C8", jobsNum)

	cnt := 0
	red := color.New(color.FgRed).SprintFunc()
	purple := color.New(color.FgHiMagenta).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for _, job := range jobs {
		maxFailuresStepCount := 0
		maxTotalStepCount := 0
		maxFailuresStepName := ""
		for _, step := range job.StepSummary {
			if step.Conclusions[parser.ConclusionFailure] > maxFailuresStepCount {
				maxFailuresStepCount = step.Conclusions[parser.ConclusionFailure]
				maxTotalStepCount = step.RunsCount
				maxFailuresStepName = step.Name
			}
		}
		if maxFailuresStepCount == 0 {
			maxTotalStepCount = job.TotalRunsCount
			maxFailuresStepName = "Failed step not found"
		}

		fmt.Fprintf(w, "  %s: %d/%d\n", cyan(job.Name), job.Conclusions[parser.ConclusionFailure], job.TotalRunsCount)
		fmt.Fprintf(w, "    └──%s: %s\n\n", purple(maxFailuresStepName), red(fmt.Sprintf("%d/%d", maxFailuresStepCount, maxTotalStepCount)))

		cnt++
		if cnt >= jobsNum {
			break
		}
	}
}

func LongestDurationJobs(w io.Writer, jobs []*parser.WorkflowJobsStatsSummary, n int) {
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].ExecutionDurationStats.Avg > jobs[j].ExecutionDurationStats.Avg
	})
	jobsNum := min(len(jobs), n)
	fmt.Fprintf(w, "\n%s Top %d jobs with the longest execution average duration\n", "\U0001F4CA", jobsNum)

	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for i := 0; i < jobsNum; i++ {
		job := jobs[i]
		fmt.Fprintf(w, "  %s: %s\n", cyan(job.Name), red(fmt.Sprintf("%.2fs", job.ExecutionDurationStats.Avg)))
	}
}
