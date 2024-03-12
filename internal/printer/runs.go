package printer

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/fchimpan/gh-workflow-stats/internal/parser"
)

func Runs(wrs *parser.WorkflowRunsStatsSummary) {
	fmt.Printf("%s Total runs: %d\n", "\U0001F3C3", wrs.TotalRunsCount)

	if _, ok := wrs.Conclusions["success"]; ok {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("  %s: %d (%.1f%%)\n", green("\u2714 Success"), wrs.Conclusions[parser.ConclusionSuccess].RunsCount, wrs.Rate.SuccesRate*100)
	}
	if _, ok := wrs.Conclusions["failure"]; ok {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("  %s: %d (%.1f%%)\n", red("\u2716 Failure"), wrs.Conclusions[parser.ConclusionFailure].RunsCount, wrs.Rate.FailureRate*100)
	}
	if _, ok := wrs.Conclusions["others"]; ok {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("  %s%s : %d (%.1f%%)\n", "\U0001F914", yellow("Others"), wrs.Conclusions[parser.ConclusionOthers].RunsCount, wrs.Rate.OthersRate*100)
	}

	fmt.Printf("\n%s Workflow run execution time stats\n", "\u23F0")
	fmt.Printf("  Min: %.1fs\n", wrs.ExecutionDurationStats.Min)
	fmt.Printf("  Max: %.1fs\n", wrs.ExecutionDurationStats.Max)
	fmt.Printf("  Avg: %.1fs\n", wrs.ExecutionDurationStats.Avg)
	fmt.Printf("  Med: %.1fs\n", wrs.ExecutionDurationStats.Med)
	fmt.Printf("  Std: %.1fs\n", wrs.ExecutionDurationStats.Std)
}
