package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fchimpan/gh-workflow-stats/internal/github"
	"github.com/fchimpan/gh-workflow-stats/internal/parser"
	"github.com/fchimpan/gh-workflow-stats/internal/printer"
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
	status              string
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
		Status:              opt.status,
		Created:             opt.created,
		HeadSHA:             opt.headSHA,
		ExcludePullRequests: opt.excludePullRequests,
		CheckSuiteID:        opt.checkSuiteID,
	},
	)
	if err != nil {
		if _, ok := err.(github.RateLimitError); ok {
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
			if _, ok := err.(github.RateLimitError); ok {
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
		printer.Runs(wrs)
		if isJobs {
			printer.FailureJobs(jobs, opt.jobNum)
			printer.LongestDurationJobs(jobs, opt.jobNum)
		}
	}

	return nil
}
