package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/fatih/color"
)

func workflowStats(cfg *workflowStatsConfig, isJobs, isJson bool) error {
	ctx := context.Background()
	client, err := newGithubClient(cfg.repoInfo.Host)
	if err != nil {
		return err
	}

	s := newSpinner(spinnerOptions{
		text:          "  fetching workflow runs...",
		charSetsIndex: 14,
		color:         "green",
	})

	s.start()

	runs, err := client.FetchWorkflowRuns(ctx, cfg)
	if err != nil {
		return err
	}

	var jobRes []*WorkflowJobsStatsSummary

	if isJobs {
		s.update(spinnerOptions{
			text: "  fetching workflow jobs...",
		})
		jobs, err := client.FetchWorkflowJobsAttempts(ctx, runs, cfg)
		if err != nil {
			return err
		}
		jobRes = workflowJobsParse(jobs)
	}

	s.stop()

	wrs := workflowRunsParse(runs)
	if isJson {
		res := &Result{
			WorkflowRunsStatsSummary: wrs,
		}
		if isJobs {
			res.WorkflowJobsStatsSummary = jobRes
		}
		bytes, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	} else {
		fmt.Printf("%s Total runs: %d\n", "\U0001F3C3", wrs.TotalRunsCount)

		if _, ok := wrs.Conclusions["success"]; ok {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("  %s %d (%.1f%%)\n", green("\u2714 Success:"), wrs.Conclusions[conclusionSuccess].RunsCount, wrs.Rate.SuccesRate*100)
		}
		if _, ok := wrs.Conclusions["failure"]; ok {
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("  %s %d (%.1f%%)\n", red("\u2716 Failure:"), wrs.Conclusions[conclusionFailure].RunsCount, wrs.Rate.FailureRate*100)
		}
		if _, ok := wrs.Conclusions["others"]; ok {
			fmt.Printf("  %s Others: %d (%.1f%%)\n", "\U0001F914", wrs.Conclusions[conclusionOthers].RunsCount, wrs.Rate.OthersRate*100)
		}

		fmt.Printf("\n%s Workflow run execution time stats\n", "\u23F0")
		fmt.Printf("  Min: %.1fs\n", wrs.ExecutionDurationStats.Min)
		fmt.Printf("  Max: %.1fs\n", wrs.ExecutionDurationStats.Max)
		fmt.Printf("  Avg: %.1fs\n", wrs.ExecutionDurationStats.Avg)
		fmt.Printf("  Med: %.1fs\n", wrs.ExecutionDurationStats.Med)
		fmt.Printf("  Std: %.1fs\n", wrs.ExecutionDurationStats.Std)

		if isJobs {
			sort.Slice(jobRes, func(i, j int) bool {
				return jobRes[i].Conclusions[conclusionFailure] > jobRes[j].Conclusions[conclusionFailure]
			})
			jobsNum := min(len(jobRes), 5)
			fmt.Printf("\n%s Top %d jobs with the highest failure count\n", "\U0001F4C8", jobsNum)

			cnt := 0
			red := color.New(color.FgRed).SprintFunc()
			purple := color.New(color.FgHiMagenta).SprintFunc()
			cyan := color.New(color.FgCyan).SprintFunc()
			for _, job := range jobRes {
				if cnt > jobsNum {
					break
				}
				if job.Rate.SuccesRate == 0 && job.Rate.FailureRate == 0 {
					continue
				}
				ss := &StepSummary{
					Rate: Rate{
						SuccesRate: 1.1,
					},
				}
				for _, step := range job.StepSummary {
					if step.Rate.SuccesRate == 0 && step.Rate.FailureRate == 0 {
						continue
					}
					if step.Rate.SuccesRate < ss.Rate.SuccesRate {
						ss = step
					}
				}

				fmt.Printf("  %s: %s\n", cyan(job.Name), red(fmt.Sprintf("%.1f%%", ss.Rate.SuccesRate*100)))
				fmt.Printf("    └──%s: %s\n\n", purple(ss.Name), red(fmt.Sprintf("%.1f%%", ss.Rate.SuccesRate*100)))

				cnt++
			}

		}
	}
	return nil
}
