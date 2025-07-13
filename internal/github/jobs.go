package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/fchimpan/gh-workflow-stats/internal/concurrency"
	"github.com/google/go-github/v60/github"
)

func (c *WorkflowStatsClient) FetchWorkflowJobsAttempts(ctx context.Context, runs []*github.WorkflowRun, cfg *WorkflowRunsConfig) ([]*github.WorkflowJob, error) {
	if len(runs) == 0 {
		c.logger.Debug("no workflow runs to fetch jobs for")
		return []*github.WorkflowJob{}, nil
	}

	c.logger.Info("starting workflow jobs fetch",
		"org", cfg.Org,
		"repo", cfg.Repo,
		"runs_count", len(runs),
	)

	// Optimize channel buffer size
	bufferSize := min(len(runs), 100)
	jobsCh := make(chan []*github.WorkflowJob, bufferSize)
	errCh := make(chan error, len(runs))

	// Use controlled concurrency with semaphore
	sem := concurrency.NewAPIClientSemaphore()
	var wg sync.WaitGroup
	wg.Add(len(runs))
	
	for _, run := range runs {
		go func(run *github.WorkflowRun) {
			defer wg.Done()
			
			// Acquire semaphore before making API request
			if err := sem.Acquire(ctx); err != nil {
				errCh <- fmt.Errorf("failed to acquire semaphore for run %d: %w", run.GetID(), err)
				return
			}
			defer sem.Release()

			c.logger.Debug("fetching jobs for workflow run",
				"run_id", run.GetID(),
				"run_attempt", run.GetRunAttempt(),
			)

			jobs, resp, err := c.client.Actions.ListWorkflowJobsAttempt(ctx, cfg.Org, cfg.Repo, run.GetID(), int64(run.GetRunAttempt()), &github.ListOptions{
				PerPage: perPage,
			})
			
			if err != nil {
				handledErr := c.handleHTTPError(resp, err, "list_workflow_jobs_attempt", fmt.Sprintf("jobs/run/%d/attempt/%d", run.GetID(), run.GetRunAttempt()))
				
				// Check if it's a rate limit error
				if _, ok := handledErr.(*RateLimitError); ok {
					c.logger.Warn("rate limit hit while fetching jobs",
						"run_id", run.GetID(),
						"run_attempt", run.GetRunAttempt(),
					)
					errCh <- handledErr
					return
				}
				
				// For 404 errors, skip silently (job might not exist)
				if resp != nil && resp.Response != nil && resp.Response.StatusCode == http.StatusNotFound {
					c.logger.Debug("workflow jobs not found, skipping",
						"run_id", run.GetID(),
						"run_attempt", run.GetRunAttempt(),
					)
					return
				}
				
				// For other errors, report them
				errCh <- handledErr
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

	// Handle collected errors
	var firstRateLimitErr error
	errorCount := 0
	
	for e := range errCh {
		if e != nil {
			errorCount++
			// Prioritize rate limit errors
			if _, ok := e.(*RateLimitError); ok && firstRateLimitErr == nil {
				firstRateLimitErr = e
			}
		}
	}

	// Collect all jobs
	allJobs := make([]*github.WorkflowJob, 0, len(runs)*10) // Estimate ~10 jobs per run
	totalJobs := 0
	
	for jobs := range jobsCh {
		allJobs = append(allJobs, jobs...)
		totalJobs += len(jobs)
	}

	c.logger.Info("completed workflow jobs fetch",
		"total_runs", len(runs),
		"total_jobs", totalJobs,
		"error_count", errorCount,
	)

	// Return rate limit error if encountered
	if firstRateLimitErr != nil {
		c.logger.Warn("returning partial results due to rate limit",
			"jobs_fetched", totalJobs,
		)
		return allJobs, firstRateLimitErr
	}

	return allJobs, nil
}
