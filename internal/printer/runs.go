package printer

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/fchimpan/gh-workflow-stats/internal/parser"
)

const (
	totalRunsFormat     = "%s Total runs: %d\n"
	conclusionFormat    = "  %s: %d (%.1f%%)\n"
	executionTimeFormat = "\n%s Workflow run execution time stats\n"
	executionFormat     = "  %s: %.1fs\n"
)

func Runs(w io.Writer, wrs *parser.WorkflowRunsStatsSummary) {
	var sc, fc, oc int
	var sr, fr, or float64
	if _, ok := wrs.Conclusions[parser.ConclusionSuccess]; ok {
		sc = wrs.Conclusions[parser.ConclusionSuccess].RunsCount
		sr = wrs.Rate.SuccesRate * 100
	}
	if _, ok := wrs.Conclusions[parser.ConclusionFailure]; ok {
		fc = wrs.Conclusions[parser.ConclusionFailure].RunsCount
		fr = wrs.Rate.FailureRate * 100
	}
	if _, ok := wrs.Conclusions[parser.ConclusionOthers]; ok {
		oc = wrs.Conclusions[parser.ConclusionOthers].RunsCount
		or = wrs.Rate.OthersRate * 100
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	_, _ = fmt.Fprintf(w, totalRunsFormat, "\U0001F3C3", wrs.TotalRunsCount)

	_, _ = fmt.Fprintf(w, conclusionFormat, green("\u2714 Success"), sc, sr)
	_, _ = fmt.Fprintf(w, conclusionFormat, red("\u2716 Failure"), fc, fr)
	_, _ = fmt.Fprintf(w, conclusionFormat, yellow("\U0001F914 Others"), oc, or)

	_, _ = fmt.Fprintf(w, executionTimeFormat, "\u23F0")
	_, _ = fmt.Fprintf(w, executionFormat, "Min", wrs.ExecutionDurationStats.Min)
	_, _ = fmt.Fprintf(w, executionFormat, "Max", wrs.ExecutionDurationStats.Max)
	_, _ = fmt.Fprintf(w, executionFormat, "Avg", wrs.ExecutionDurationStats.Avg)
	_, _ = fmt.Fprintf(w, executionFormat, "Med", wrs.ExecutionDurationStats.Med)
	_, _ = fmt.Fprintf(w, executionFormat, "Std", wrs.ExecutionDurationStats.Std)
}
