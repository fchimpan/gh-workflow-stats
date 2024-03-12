package github

import (
	"context"
	"sync"

	"github.com/google/go-github/v60/github"
)

const perPage = 100

type WorkflowRunsConfig struct {
	Org              string
	Repo             string
	WorkflowFileName string
	WorkflowID       int64
}

type WorkflowRunsOptions struct {
	Actor               string
	Branch              string
	Event               string
	Status              string
	Created             string
	HeadSHA             string
	ExcludePullRequests bool
	CheckSuiteID        int64
	All                 bool
}

func (c *WorkflowStatsClient) FetchWorkflowRuns(ctx context.Context, cfg *WorkflowRunsConfig, opt *WorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	o := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{
			PerPage: perPage,
		},
		Actor:               opt.Actor,
		Branch:              opt.Branch,
		Event:               opt.Event,
		Status:              opt.Status,
		Created:             opt.Created,
		HeadSHA:             opt.HeadSHA,
		ExcludePullRequests: opt.ExcludePullRequests,
		CheckSuiteID:        opt.CheckSuiteID,
	}

	initRuns, resp, err := c.listWorkflowRuns(ctx, cfg, o)
	if err != nil {
		return nil, err
	}
	if resp.FirstPage == resp.LastPage || !opt.All {
		return initRuns.WorkflowRuns, nil
	}

	var wg sync.WaitGroup
	wg.Add(resp.LastPage)
	runsCh := make(chan []*github.WorkflowRun, resp.LastPage)
	errCh := make(chan error, resp.LastPage)

	for i := resp.FirstPage + 1; i <= resp.LastPage; i++ {
		go func(i int) error {
			defer wg.Done()

			runs, _, err := c.listWorkflowRuns(ctx, cfg, &github.ListWorkflowRunsOptions{
				ListOptions: github.ListOptions{
					Page:    i,
					PerPage: perPage,
				},
				Actor:               opt.Actor,
				Branch:              opt.Branch,
				Event:               opt.Event,
				Status:              opt.Status,
				Created:             opt.Created,
				HeadSHA:             opt.HeadSHA,
				ExcludePullRequests: opt.ExcludePullRequests,
				CheckSuiteID:        opt.CheckSuiteID,
			})
			if err != nil {
				errCh <- err
				return err
			}
			runsCh <- runs.WorkflowRuns

			return nil
		}(i)
	}
	wg.Wait()
	close(runsCh)
	close(errCh)

	for e := range errCh {
		if e != nil {
			err = e
		}
	}
	if err != nil {
		return nil, err
	}
	allRuns := make([]*github.WorkflowRun, 0, *initRuns.TotalCount)
	for runs := range runsCh {
		allRuns = append(allRuns, runs...)
	}
	return allRuns, nil
}

func (c *WorkflowStatsClient) listWorkflowRuns(ctx context.Context, cfg *WorkflowRunsConfig, opt *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	if cfg.WorkflowFileName != "" {
		return c.client.Actions.ListWorkflowRunsByFileName(ctx, cfg.Org, cfg.Repo, cfg.WorkflowFileName, opt)
	} else {
		return c.client.Actions.ListWorkflowRunsByID(ctx, cfg.Org, cfg.Repo, cfg.WorkflowID, opt)
	}
}
