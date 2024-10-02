package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/fchimpan/gh-workflow-stats/internal/github"
	"github.com/fchimpan/gh-workflow-stats/internal/parser"
	"github.com/fchimpan/gh-workflow-stats/internal/printer"

	go_github "github.com/google/go-github/v60/github"
)

const (
	workflowRunsText = "  fetching workflow runs..."
	workflowJobsText = "  fetching workflow jobs..."
	charSize         = 14
)

type config struct {
	host             string
	org              string
	repo             string
	workflowFileName string
	workflowID       int64
}

type options struct {
	actor               string
	branch              string
	event               string
	status              []string
	created             string
	headSHA             string
	excludePullRequests bool
	checkSuiteID        int64
	all                 bool
	js                  bool
	jobNum              int
}

func workflowStats(cfg config, opt options, isJobs bool) error {
	ctx := context.Background()
	w := io.Writer(os.Stdout)
	a := &github.GitHubAuthenticator{}
	client, err := github.NewClient(cfg.host, a)
	if err != nil {
		return err
	}

	s, err := printer.NewSpinner(printer.SpinnerOptions{
		Text:          workflowRunsText,
		CharSetsIndex: charSize,
		Color:         "green",
	})
	if err != nil {
		return err
	}
	s.Start()
	defer s.Stop()

	isRateLimit := false
	runs, err := fetchWorkflowRuns(ctx, client, cfg, opt)
	if err != nil {
		if errors.As(err, &github.RateLimitError{}) {
			isRateLimit = true
		} else {
			return err
		}
	}

	var jobs []*parser.WorkflowJobsStatsSummary
	if isJobs {
		s.Update(printer.SpinnerOptions{
			Text:          workflowJobsText,
			CharSetsIndex: charSize,
			Color:         "pink",
		})
		j, err := client.FetchWorkflowJobsAttempts(ctx, runs, &github.WorkflowRunsConfig{
			Org:  cfg.org,
			Repo: cfg.repo,
		})
		if err != nil {
			if errors.As(err, &github.RateLimitError{}) {
				isRateLimit = true
			} else {
				return err
			}
		}
		jobs = parser.WorkflowJobsParse(j)
	}

	s.Stop()

	wrs := parser.WorkflowRunsParse(runs)

	if opt.js {
		res := &parser.Result{
			WorkflowRunsStatsSummary: wrs,
			WorkflowJobsStatsSummary: []*parser.WorkflowJobsStatsSummary{},
		}
		if isJobs {
			res.WorkflowJobsStatsSummary = jobs
		}
		bytes, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	} else {
		if isRateLimit {
			printer.RateLimitWarning(os.Stdout)
		}
		printer.Runs(w, wrs)
		if isJobs {
			printer.FailureJobs(w, jobs, opt.jobNum)
			printer.LongestDurationJobs(w, jobs, opt.jobNum)
		}
	}

	return nil
}

func fetchWorkflowRuns(ctx context.Context, client *github.WorkflowStatsClient, cfg config, opt options) ([]*go_github.WorkflowRun, error) {
	// Intentionally not using Github API status filter as it applies only to the last run attempt.
	// Instead retrieving all qualifying workflow runs and their run attempts and filtering by status manually (if needed)
	runs, err := client.FetchWorkflowRuns(ctx, &github.WorkflowRunsConfig{
		Org:              cfg.org,
		Repo:             cfg.repo,
		WorkflowFileName: cfg.workflowFileName,
		WorkflowID:       cfg.workflowID,
	}, &github.WorkflowRunsOptions{
		All:                 opt.all,
		Actor:               opt.actor,
		Branch:              opt.branch,
		Event:               opt.event,
		Status:              "",
		Created:             opt.created,
		HeadSHA:             opt.headSHA,
		ExcludePullRequests: opt.excludePullRequests,
		CheckSuiteID:        opt.checkSuiteID,
	},
	)
	if err == nil {
		return filterRunAttemptsByStatus(runs, opt.status), nil
	} else {
		return nil, err
	}
}

func filterRunAttemptsByStatus(runs []*go_github.WorkflowRun, status []string) []*go_github.WorkflowRun {
	if len(status) == 0 || (len(status) == 1 && status[0] == "") {
		return runs
	}
	filteredRuns := []*go_github.WorkflowRun{}
	for _, r := range runs {
		if (r.Status != nil && slices.Contains(status, *r.Status)) || (r.Conclusion != nil && slices.Contains(status, *r.Conclusion)) {
			filteredRuns = append(filteredRuns, r)
		}
	}
	return filteredRuns
}
