package cmd

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/cli/go-gh/pkg/auth"
	"github.com/google/go-github/v57/github"
)

type workflowStatsClient struct {
	client *github.Client
}

func authTokenForHost(host string) (string, error) {
	token, _ := auth.TokenForHost(host)
	if token == "" {
		return "", fmt.Errorf("gh auth token not found for host %q", host)
	}
	return token, nil
}

func newGithubClient(host string) (*workflowStatsClient, error) {
	token, err := authTokenForHost(host)
	if err != nil {
		return nil, err
	}
	client := github.NewClient(nil).WithAuthToken(token)
	if host != defaultHost {
		client.BaseURL.Host = host
		client.BaseURL.Path = "/api/v3/"
	}
	return &workflowStatsClient{client: client}, nil
}

// List all latest workflow runs.
func (c *workflowStatsClient) fetchWorkflowRuns(ctx context.Context, cfg *workflowStatsConfig) ([]*github.WorkflowRun, error) {

	initRuns, resp, err := c.listWorkflowRuns(ctx, cfg.RepoInfo, cfg.Params)
	if err != nil {
		return nil, err
	}
	if resp.FirstPage == resp.LastPage {
		return initRuns.WorkflowRuns, nil
	}

	var wg sync.WaitGroup
	wg.Add(resp.LastPage)
	runsCh := make(chan []*github.WorkflowRun, resp.LastPage)
	errCh := make(chan error, 1) // TODO: Handle some errors.

	for i := resp.FirstPage + 1; i <= resp.LastPage; i++ {
		go func(i int) error {
			defer wg.Done()
			runs, _, err := c.listWorkflowRuns(ctx, cfg.RepoInfo, &github.ListWorkflowRunsOptions{
				ListOptions: github.ListOptions{
					Page:    i,
					PerPage: cfg.Params.PerPage,
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

	close(errCh)
	close(runsCh)

	// TODO: Handle error.
	if err, ok := <-errCh; ok {
		return nil, err
	}
	allRuns := make([]*github.WorkflowRun, 0, *initRuns.TotalCount)
	for runs := range runsCh {
		allRuns = append(allRuns, runs...)
	}
	return allRuns, nil
}

func (c *workflowStatsClient) listWorkflowRuns(ctx context.Context, ri *RepositoryInfo, params *github.ListWorkflowRunsOptions) (*github.WorkflowRuns, *github.Response, error) {
	if ri.WorkflowFileName != "" {
		return c.client.Actions.ListWorkflowRunsByFileName(ctx, ri.Org, ri.Repo, ri.WorkflowFileName, params)
	} else {
		return c.client.Actions.ListWorkflowRunsByID(ctx, ri.Org, ri.Repo, ri.ID, params)
	}
}

func (c *workflowStatsClient) fetchWorkflowJobsAttempts(ctx context.Context, runs []*github.WorkflowRun, cfg *workflowStatsConfig) ([]*github.WorkflowJob, error) {
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
			jobs, resp, err := c.client.Actions.ListWorkflowJobs(ctx, cfg.RepoInfo.Org, cfg.RepoInfo.Repo, run.GetID(), &github.ListWorkflowJobsOptions{
				ListOptions: github.ListOptions{
					PerPage: cfg.Params.PerPage,
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

	var err error
	close(errCh)
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
