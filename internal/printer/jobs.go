package printer

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/fchimpan/workflow-stats/internal/parser"
)

func FailureJobs(jobRes []*parser.WorkflowJobsStatsSummary, n int) {
	sort.Slice(jobRes, func(i, j int) bool {
		return jobRes[i].Conclusions[parser.ConclusionFailure] > jobRes[j].Conclusions[parser.ConclusionFailure]
	})
	jobsNum := min(len(jobRes), n)
	fmt.Printf("\n%s Top %d jobs with the highest failure counts (failure runs / total runs)\n", "\U0001F4C8", jobsNum)

	cnt := 0
	red := color.New(color.FgRed).SprintFunc()
	purple := color.New(color.FgHiMagenta).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for _, job := range jobRes {
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

		fmt.Printf("  %s\n", cyan(job.Name))
		fmt.Printf("    └──%s: %s\n\n", purple(maxFailuresStepName), red(fmt.Sprintf("%d/%d", maxFailuresStepCount, maxTotalStepCount)))

		cnt++
		if cnt >= jobsNum {
			break
		}
	}
}

func LongestDurationJobs(jobRes []*parser.WorkflowJobsStatsSummary, n int) {
	sort.Slice(jobRes, func(i, j int) bool {
		return jobRes[i].ExecutionDurationStats.Avg > jobRes[j].ExecutionDurationStats.Avg
	})
	jobsNum := min(len(jobRes), n)
	fmt.Printf("\n%s Top %d jobs with the longest execution average duration\n", "\U0001F4CA", jobsNum)

	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for i := 0; i < jobsNum; i++ {
		job := jobRes[i]
		fmt.Printf("  %s: %s\n", cyan(job.Name), red(fmt.Sprintf("%.2fs", job.ExecutionDurationStats.Avg)))
	}
}
