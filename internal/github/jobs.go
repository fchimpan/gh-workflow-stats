package github

import (
	"context"
	"net/http"
	"runtime"
	"sync"

	"github.com/google/go-github/v60/github"
)

func (c *WorkflowStatsClient) FetchWorkflowJobsAttempts(ctx context.Context, runs []*github.WorkflowRun, cfg *WorkflowRunsConfig) ([]*github.WorkflowJob, error) {
	if len(runs) == 0 {
		return []*github.WorkflowJob{}, nil
	}
	jobsCh := make(chan []*github.WorkflowJob, len(runs))
	errCh := make(chan error, len(runs))

	// TODO: Semaphore to limit the number of concurrent requests. It's a assumption, not accurate.
	sem := make(chan struct{}, runtime.NumCPU()*8)
	var wg sync.WaitGroup
	wg.Add(len(runs))
	for _, run := range runs {
		sem <- struct{}{}

		go func(run *github.WorkflowRun) {
			defer func() {
				<-sem
				wg.Done()
			}()
			jobs, resp, err := c.client.Actions.ListWorkflowJobsAttempt(ctx, cfg.Org, cfg.Repo, run.GetID(), int64(run.GetRunAttempt()), &github.ListOptions{
				PerPage: perPage,
			},
			)
			if err != nil && resp.StatusCode != http.StatusNotFound {
				errCh <- err
			}
			jobsCh <- jobs.Jobs

		}(run)
	}
	wg.Wait()
	close(jobsCh)
	close(errCh)

	for e := range errCh {
		if e != nil {
			return nil, e
		}
	}

	allJobs := make([]*github.WorkflowJob, 0, len(runs))
	for jobs := range jobsCh {
		allJobs = append(allJobs, jobs...)
	}
	return allJobs, nil
}
