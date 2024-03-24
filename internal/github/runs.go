package github

import (
	"context"
	"net/http"
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
		if _, ok := err.(*github.RateLimitError); ok {
			return nil, RateLimitError{Err: err}
		}
		return nil, err
	}

	w := make([]*github.WorkflowRun, 0, initRuns.GetTotalCount()*2)
	w = append(w, initRuns.WorkflowRuns...)
	if resp.FirstPage == resp.LastPage || !opt.All {
		for _, run := range initRuns.WorkflowRuns {
			for a := 1; a < run.GetRunAttempt(); a++ {
				r, resp, err := c.client.Actions.GetWorkflowRunAttempt(ctx, cfg.Org, cfg.Repo, run.GetID(), a, &github.WorkflowRunAttemptOptions{
					ExcludePullRequests: &opt.ExcludePullRequests,
				})
				if _, ok := err.(*github.RateLimitError); ok {
					return w, RateLimitError{Err: err}
				} else if err != nil && resp.StatusCode != http.StatusNotFound {
					return nil, err
				}
				w = append(w, r)
			}
		}
		return w, nil
	}

	var wg sync.WaitGroup
	wg.Add(resp.LastPage)
	runsCh := make(chan []*github.WorkflowRun, *initRuns.TotalCount*10)
	errCh := make(chan error, resp.LastPage)

	for i := resp.FirstPage + 1; i <= resp.LastPage; i++ {
		go func(i int) {
			defer wg.Done()

			runs, resp, err := c.listWorkflowRuns(ctx, cfg, &github.ListWorkflowRunsOptions{
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
			if _, ok := err.(*github.RateLimitError); ok {
				errCh <- RateLimitError{Err: err}
				return
			} else if err != nil && resp.StatusCode != http.StatusNotFound {
				errCh <- err
			}
			w := make([]*github.WorkflowRun, 0, runs.GetTotalCount()*2)
			w = append(w, runs.WorkflowRuns...)
			for _, run := range runs.WorkflowRuns {
				for a := 1; a < run.GetRunAttempt(); a++ {
					r, resp, err := c.client.Actions.GetWorkflowRunAttempt(ctx, cfg.Org, cfg.Repo, run.GetID(), a, nil)
					if _, ok := err.(*github.RateLimitError); ok {
						errCh <- RateLimitError{Err: err}
						return
					} else if err != nil && resp.StatusCode != http.StatusNotFound {
						errCh <- err
					}
					w = append(w, r)
				}
			}
			runsCh <- w
		}(i)
	}
	wg.Wait()
	close(runsCh)
	close(errCh)

	for e := range errCh {
		if e != nil {
			return nil, e
		}
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
