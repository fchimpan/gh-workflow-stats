package workflow

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/google/go-github/v57/github"
)

func (c *workflowStatsClient) FetchWorkflowJobsAttempts(ctx context.Context, runs []*github.WorkflowRun, cfg *workflowStatsConfig) ([]*github.WorkflowJob, error) {
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
			jobs, resp, err := c.client.Actions.ListWorkflowJobs(ctx, cfg.repoInfo.Org, cfg.repoInfo.Repo, run.GetID(), &github.ListWorkflowJobsOptions{
				ListOptions: github.ListOptions{
					PerPage: cfg.params.PerPage,
				},
				Filter: "all",
			})
			if err != nil || resp.StatusCode != http.StatusOK {
				errCh <- err
			} else if jobs == nil || jobs.Jobs == nil {
				errCh <- fmt.Errorf("no jobs found for run %d", run.GetID())
			} else {
				jobsCh <- jobs.Jobs
			}
		}(run)
	}
	wg.Wait()
	close(jobsCh)
	close(errCh)

	var err error
	for e := range errCh {
		if e != nil {
			err = e
		}
	}
	if err != nil {
		return nil, err
	}
	allJobs := make([]*github.WorkflowJob, 0, len(runs))
	for jobs := range jobsCh {
		allJobs = append(allJobs, jobs...)
	}
	return allJobs, nil
}
