package workflow

import (
	"context"
	"sync"

	"github.com/google/go-github/v57/github"
)

func (c *workflowStatsClient) FetchWorkflowRuns(ctx context.Context, cfg *workflowStatsConfig) ([]*github.WorkflowRun, error) {
	initRuns, resp, err := c.listWorkflowRuns(ctx, cfg.repoInfo, cfg.params)
	if err != nil {
		return nil, err
	}
	if resp.FirstPage == resp.LastPage {
		return initRuns.WorkflowRuns, nil
	}

	var wg sync.WaitGroup
	wg.Add(resp.LastPage)
	runsCh := make(chan []*github.WorkflowRun, resp.LastPage)
	errCh := make(chan error, resp.LastPage)

	for i := resp.FirstPage + 1; i <= resp.LastPage; i++ {
		go func(i int) error {
			defer wg.Done()
			runs, _, err := c.listWorkflowRuns(ctx, cfg.repoInfo, &github.ListWorkflowRunsOptions{
				ListOptions: github.ListOptions{
					Page:    i,
					PerPage: cfg.params.PerPage,
				},
			})
			if err != nil {
				errCh <- err
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

func (c *workflowStatsClient) listWorkflowRuns(ctx context.Context, ri *repositoryInfo, params *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	if ri.WorkflowFileName != "" {
		return c.client.Actions.ListWorkflowRunsByFileName(ctx, ri.Org, ri.Repo, ri.WorkflowFileName, params)
	} else {
		return c.client.Actions.ListWorkflowRunsByID(ctx, ri.Org, ri.Repo, ri.WorkflowID, params)
	}
}
