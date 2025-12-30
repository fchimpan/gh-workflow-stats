package github

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/google/go-github/v60/github"
)

func (c *WorkflowStatsClient) FetchWorkflowJobsAttempts(ctx context.Context, runs []*github.WorkflowRun, cfg *WorkflowRunsConfig) ([]*github.WorkflowJob, error) {
	// Validate inputs
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	if len(runs) == 0 {
		c.logger.Debug("no workflow runs to fetch jobs for")
		return []*github.WorkflowJob{}, nil
	}

	c.logger.Info("starting workflow jobs fetch",
		"org", cfg.Org,
		"repo", cfg.Repo,
		"runs_count", len(runs),
	)

	jobsCh := make(chan []*github.WorkflowJob, len(runs))
	errCh := make(chan error, len(runs))

	// Use original high-concurrency semaphore for performance
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

			// Skip nil runs
			if run == nil {
				c.logger.Debug("skipping nil workflow run")
				return
			}

			c.logger.Debug("fetching jobs for workflow run",
				"run_id", run.GetID(),
				"run_attempt", run.GetRunAttempt(),
			)

			jobs, resp, err := c.client.Actions.ListWorkflowJobsAttempt(ctx, cfg.Org, cfg.Repo, run.GetID(), int64(run.GetRunAttempt()), &github.ListOptions{
				PerPage: perPage,
			})

			if err != nil {
				// Handle rate limit errors specifically
				if _, ok := err.(*github.RateLimitError); ok {
					errCh <- RateLimitError{Err: err}
					return
				}

				// For 404 errors, skip silently (job might not exist)
				if resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound {
					c.logger.Debug("workflow jobs not found, skipping",
						"run_id", run.GetID(),
						"run_attempt", run.GetRunAttempt(),
					)
					return
				}

				// For other HTTP errors, skip
				if resp != nil && resp.Response != nil {
					c.logger.Debug("HTTP error fetching jobs, skipping",
						"run_id", run.GetID(),
						"status_code", resp.StatusCode,
					)
					return
				}

				// For other errors, report them
				errCh <- err
				return
			}

			if jobs == nil || jobs.Jobs == nil || len(jobs.Jobs) == 0 {
				c.logger.Debug("no jobs found for run",
					"run_id", run.GetID(),
					"run_attempt", run.GetRunAttempt(),
				)
				return
			}

			c.logger.Debug("fetched jobs for run",
				"run_id", run.GetID(),
				"jobs_count", len(jobs.Jobs),
			)

			jobsCh <- jobs.Jobs
		}(run)
	}

	wg.Wait()
	close(jobsCh)
	close(errCh)

	// Handle errors - prioritize rate limit errors
	var err error
	for e := range errCh {
		if e != nil {
			if _, ok := e.(*RateLimitError); ok {
				err = e
			} else if err == nil {
				err = e
			}
		}
	}

	// Collect all jobs
	allJobs := make([]*github.WorkflowJob, 0, len(runs))
	for jobs := range jobsCh {
		allJobs = append(allJobs, jobs...)
	}

	c.logger.Info("completed workflow jobs fetch",
		"total_runs", len(runs),
		"total_jobs", len(allJobs),
	)

	return allJobs, err
}
