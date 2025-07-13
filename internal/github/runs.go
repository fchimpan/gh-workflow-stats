package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/fchimpan/gh-workflow-stats/internal/concurrency"
	"github.com/fchimpan/gh-workflow-stats/internal/types"
	"github.com/google/go-github/v60/github"
)

const perPage = types.DefaultPerPage

// Legacy types for backward compatibility - use types package instead
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

// Adapter functions for new type system

// ToWorkflowConfig converts legacy WorkflowRunsConfig to types.WorkflowConfig
func (c *WorkflowRunsConfig) ToWorkflowConfig() *types.WorkflowConfig {
	return &types.WorkflowConfig{
		Host:             "", // Host is handled at the client level
		Org:              c.Org,
		Repo:             c.Repo,
		WorkflowFileName: c.WorkflowFileName,
		WorkflowID:       c.WorkflowID,
	}
}

// ToAPIRequestOptions converts legacy WorkflowRunsOptions to types.APIRequestOptions
func (o *WorkflowRunsOptions) ToAPIRequestOptions() *types.APIRequestOptions {
	return &types.APIRequestOptions{
		Actor:               o.Actor,
		Branch:              o.Branch,
		Event:               o.Event,
		Status:              o.Status,
		Created:             o.Created,
		HeadSHA:             o.HeadSHA,
		ExcludePullRequests: o.ExcludePullRequests,
		CheckSuiteID:        o.CheckSuiteID,
		All:                 o.All,
	}
}

// FromWorkflowConfig creates legacy WorkflowRunsConfig from types.WorkflowConfig
func FromWorkflowConfig(config *types.WorkflowConfig) *WorkflowRunsConfig {
	return &WorkflowRunsConfig{
		Org:              config.Org,
		Repo:             config.Repo,
		WorkflowFileName: config.WorkflowFileName,
		WorkflowID:       config.WorkflowID,
	}
}

// FromAPIRequestOptions creates legacy WorkflowRunsOptions from types.APIRequestOptions
func FromAPIRequestOptions(options *types.APIRequestOptions) *WorkflowRunsOptions {
	return &WorkflowRunsOptions{
		Actor:               options.Actor,
		Branch:              options.Branch,
		Event:               options.Event,
		Status:              options.Status,
		Created:             options.Created,
		HeadSHA:             options.HeadSHA,
		ExcludePullRequests: options.ExcludePullRequests,
		CheckSuiteID:        options.CheckSuiteID,
		All:                 options.All,
	}
}

// fetchRunAttempts fetches all attempts for a set of workflow runs
func (c *WorkflowStatsClient) fetchRunAttempts(ctx context.Context, cfg *WorkflowRunsConfig, runs []*github.WorkflowRun, excludePullRequests bool) ([]*github.WorkflowRun, error) {
	if len(runs) == 0 {
		c.logger.Debug("no workflow runs to fetch attempts for")
		return []*github.WorkflowRun{}, nil
	}

	// First pass: count runs that need attempt fetching
	runsNeedingAttempts := make([]*github.WorkflowRun, 0, len(runs))
	totalEstimatedAttempts := 0
	
	for _, run := range runs {
		if run == nil {
			c.logger.Warn("skipping nil workflow run")
			continue
		}

		runAttempt := run.GetRunAttempt()
		if runAttempt <= 1 {
			continue // No additional attempts to fetch
		}
		
		runsNeedingAttempts = append(runsNeedingAttempts, run)
		totalEstimatedAttempts += runAttempt - 1 // exclude the first attempt which we already have
	}
	
	if len(runsNeedingAttempts) == 0 {
		c.logger.Debug("no workflow runs need additional attempts")
		return []*github.WorkflowRun{}, nil
	}

	attempts := make([]*github.WorkflowRun, 0, totalEstimatedAttempts)
	c.logger.Debug("fetching workflow run attempts",
		"total_runs", len(runs),
		"runs_needing_attempts", len(runsNeedingAttempts),
		"estimated_attempts", totalEstimatedAttempts,
		"org", cfg.Org,
		"repo", cfg.Repo,
	)

	// Use controlled concurrency for attempt fetching
	sem := concurrency.NewAPIClientSemaphore()
	var mu sync.Mutex
	var wg sync.WaitGroup
	var fetchErrors []error

	for _, run := range runsNeedingAttempts {
		runAttempt := run.GetRunAttempt()
		c.logger.Debug("fetching attempts for workflow run",
			"run_id", run.GetID(),
			"total_attempts", runAttempt,
		)

		for a := 1; a < runAttempt; a++ {
			wg.Add(1)
			go func(runID int64, attemptNum int) {
				defer wg.Done()
				
				// Acquire semaphore to limit concurrent API calls
				if err := sem.Acquire(ctx); err != nil {
					mu.Lock()
					fetchErrors = append(fetchErrors, fmt.Errorf("failed to acquire semaphore for run %d attempt %d: %w", runID, attemptNum, err))
					mu.Unlock()
					return
				}
				defer sem.Release()
				
				r, resp, err := c.client.Actions.GetWorkflowRunAttempt(ctx, cfg.Org, cfg.Repo, runID, attemptNum, &github.WorkflowRunAttemptOptions{
					ExcludePullRequests: &excludePullRequests,
				})

				if err != nil {
					handledErr := c.handleHTTPError(resp, err, "fetch_workflow_run_attempt", fmt.Sprintf("runs/%d/attempts/%d", runID, attemptNum))

					// Check if it's a rate limit error
					if _, ok := handledErr.(*RateLimitError); ok {
						c.logger.Warn("rate limit hit while fetching run attempts",
							"run_id", runID,
							"attempt", attemptNum,
						)
						mu.Lock()
						fetchErrors = append(fetchErrors, handledErr)
						mu.Unlock()
						return
					}

					// For 404 errors, continue as the attempt might not exist
					if resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound {
						c.logger.Debug("workflow run attempt not found, skipping",
							"run_id", runID,
							"attempt", attemptNum,
						)
						return
					}

					// For other errors, collect error
					mu.Lock()
					fetchErrors = append(fetchErrors, handledErr)
					mu.Unlock()
					return
				}

				if r != nil {
					mu.Lock()
					attempts = append(attempts, r)
					mu.Unlock()
				}
			}(run.GetID(), a)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Handle collected errors
	if len(fetchErrors) > 0 {
		// Return rate limit errors first (highest priority)
		for _, err := range fetchErrors {
			if _, ok := err.(*RateLimitError); ok {
				c.logger.Warn("rate limit encountered while fetching attempts, returning partial results",
					"fetched_attempts", len(attempts),
					"total_errors", len(fetchErrors),
				)
				return attempts, err
			}
		}
		
		// Log other errors but continue
		c.logger.Warn("some attempt fetches failed, continuing with partial results",
			"fetched_attempts", len(attempts),
			"failed_attempts", len(fetchErrors),
		)
	}

	c.logger.Info("completed fetching workflow run attempts",
		"total_attempts_fetched", len(attempts),
		"total_runs_processed", len(runsNeedingAttempts),
		"total_errors", len(fetchErrors),
	)

	return attempts, nil
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createListOptions creates github.ListWorkflowRunsOptions from WorkflowRunsOptions
func createListOptions(opt *WorkflowRunsOptions, page int) *github.ListWorkflowRunsOptions {
	return &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{
			Page:    page,
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
}

func (c *WorkflowStatsClient) FetchWorkflowRuns(ctx context.Context, cfg *WorkflowRunsConfig, opt *WorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	c.logger.Info("starting workflow runs fetch",
		"org", cfg.Org,
		"repo", cfg.Repo,
		"workflow_file", cfg.WorkflowFileName,
		"workflow_id", cfg.WorkflowID,
		"fetch_all", opt.All,
	)

	o := createListOptions(opt, 0)

	initRuns, resp, err := c.listWorkflowRuns(ctx, cfg, o)
	if err != nil {
		handledErr := c.handleHTTPError(resp, err, "list_workflow_runs", "workflow_runs")
		return nil, handledErr
	}

	if initRuns == nil || len(initRuns.WorkflowRuns) == 0 {
		c.logger.Warn("no workflow runs found",
			"org", cfg.Org,
			"repo", cfg.Repo,
			"workflow_file", cfg.WorkflowFileName,
			"workflow_id", cfg.WorkflowID,
		)
		return []*github.WorkflowRun{}, nil
	}

	c.logger.Debug("fetched initial workflow runs",
		"count", len(initRuns.WorkflowRuns),
		"total_count", initRuns.GetTotalCount(),
		"first_page", resp.FirstPage,
		"last_page", resp.LastPage,
	)

	// More efficient initial capacity calculation
	initialCapacity := initRuns.GetTotalCount()
	if !opt.ExcludePullRequests {
		// Only add buffer for attempts if they will be fetched
		initialCapacity = int(float64(initialCapacity) * 1.2) // 20% buffer instead of 100%
	}
	w := make([]*github.WorkflowRun, 0, initialCapacity)
	w = append(w, initRuns.WorkflowRuns...)

	// Fetch attempts for initial runs only if necessary
	var attempts []*github.WorkflowRun
	if !opt.ExcludePullRequests && len(initRuns.WorkflowRuns) > 0 {
		var attemptsErr error
		attempts, attemptsErr = c.fetchRunAttempts(ctx, cfg, initRuns.WorkflowRuns, opt.ExcludePullRequests)
		err = attemptsErr
	}
	if err != nil {
		// Log but don't fail completely for rate limit errors
		if _, ok := err.(*RateLimitError); ok {
			c.logger.Warn("rate limit encountered while fetching attempts, continuing with partial data")
			w = append(w, attempts...)
			return w, err
		}
		return w, err
	}
	w = append(w, attempts...)

	if resp.FirstPage == resp.LastPage || !opt.All {
		c.logger.Info("completed workflow runs fetch (single page)",
			"total_runs", len(w),
			"initial_runs", len(initRuns.WorkflowRuns),
			"attempts", len(attempts),
		)
		return w, nil
	}

	c.logger.Debug("fetching additional pages",
		"total_pages", resp.LastPage,
		"remaining_pages", resp.LastPage-resp.FirstPage,
	)

	// Create a semaphore to limit concurrent API requests
	sem := concurrency.NewAPIClientSemaphore()
	
	var wg sync.WaitGroup
	remainingPages := resp.LastPage - resp.FirstPage
	wg.Add(remainingPages)
	
	// Optimize channel buffer size - use reasonable buffer based on page count
	bufferSize := min(remainingPages, 10)
	runsCh := make(chan []*github.WorkflowRun, bufferSize)
	errCh := make(chan error, remainingPages)

	for i := resp.FirstPage + 1; i <= resp.LastPage; i++ {
		go func(pageNum int) {
			defer wg.Done()
			
			// Acquire semaphore before making API request
			if err := sem.Acquire(ctx); err != nil {
				errCh <- fmt.Errorf("failed to acquire semaphore for page %d: %w", pageNum, err)
				return
			}
			defer sem.Release()

			c.logger.Debug("fetching workflow runs page", "page", pageNum)

			o := createListOptions(opt, pageNum)
			runs, pageResp, err := c.listWorkflowRuns(ctx, cfg, o)
			if err != nil {
				handledErr := c.handleHTTPError(pageResp, err, "list_workflow_runs_page", fmt.Sprintf("workflow_runs/page/%d", pageNum))

				if _, ok := handledErr.(*RateLimitError); ok {
					c.logger.Warn("rate limit hit on page fetch", "page", pageNum)
				} else {
					c.logger.Error("error fetching workflow runs page", "page", pageNum, "error", handledErr)
				}
				errCh <- handledErr
				return
			}

			if runs == nil || len(runs.WorkflowRuns) == 0 {
				c.logger.Debug("no runs found on page", "page", pageNum)
				runsCh <- []*github.WorkflowRun{}
				return
			}

			// Pre-calculate capacity more efficiently
			expectedCapacity := len(runs.WorkflowRuns)
			if !opt.ExcludePullRequests {
				// Only estimate additional capacity for attempts if needed
				expectedCapacity = int(float64(expectedCapacity) * 1.2) // 20% buffer instead of 100%
			}
			pageRuns := make([]*github.WorkflowRun, 0, expectedCapacity)
			pageRuns = append(pageRuns, runs.WorkflowRuns...)

			// Fetch attempts for this page's runs only if necessary
			if !opt.ExcludePullRequests && len(runs.WorkflowRuns) > 0 {
				attempts, err := c.fetchRunAttempts(ctx, cfg, runs.WorkflowRuns, opt.ExcludePullRequests)
				if err != nil {
					c.logger.Warn("error fetching attempts for page, continuing without attempts",
						"page", pageNum,
						"error", err,
						"runs_without_attempts", len(pageRuns),
					)
					// Continue with runs but without attempts for this page
				} else {
					pageRuns = append(pageRuns, attempts...)
				}
			}

			c.logger.Debug("completed page fetch",
				"page", pageNum,
				"runs_count", len(runs.WorkflowRuns),
				"attempts_count", len(attempts),
				"total_page_runs", len(pageRuns),
			)

			runsCh <- pageRuns
		}(i)
	}
	wg.Wait()
	close(runsCh)
	close(errCh)

	// Handle errors from parallel processing
	var firstError error
	errorCount := 0
	for e := range errCh {
		if e != nil {
			errorCount++
			if firstError == nil {
				firstError = e
			}
			c.logger.Error("error in parallel workflow runs fetch", "error", e)
		}
	}

	// Collect results from all successful pages
	totalAdditionalRuns := 0
	for runs := range runsCh {
		w = append(w, runs...)
		totalAdditionalRuns += len(runs)
	}

	c.logger.Info("completed workflow runs fetch",
		"total_runs", len(w),
		"initial_runs", len(initRuns.WorkflowRuns),
		"additional_runs", totalAdditionalRuns,
		"pages_processed", resp.LastPage,
		"error_count", errorCount,
	)

	// If we have some results but also errors, log warning but continue
	if firstError != nil && len(w) > 0 {
		c.logger.Warn("completed with partial results due to errors",
			"successful_runs", len(w),
			"first_error", firstError,
		)
		// Return partial results with the first error for potential retry handling
		return w, firstError
	} else if firstError != nil {
		// If we have no results and errors, return the error
		return nil, firstError
	}

	return w, nil
}

func (c *WorkflowStatsClient) listWorkflowRuns(ctx context.Context, cfg *WorkflowRunsConfig, opt *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	if cfg.WorkflowFileName != "" {
		return c.client.Actions.ListWorkflowRunsByFileName(ctx, cfg.Org, cfg.Repo, cfg.WorkflowFileName, opt)
	} else {
		return c.client.Actions.ListWorkflowRunsByID(ctx, cfg.Org, cfg.Repo, cfg.WorkflowID, opt)
	}
}
