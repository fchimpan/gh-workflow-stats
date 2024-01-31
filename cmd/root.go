package cmd

import (
	"fmt"

	"github.com/google/go-github/v57/github"
	"github.com/spf13/cobra"
)

const defaultHost = "github.com"

type RepositoryInfo struct {
	Host             string
	Org              string
	Repo             string
	WorkflowFileName string
	ID               int64
}

type workflowStatsConfig struct {
	RepoInfo *RepositoryInfo
	Params   *github.ListWorkflowRunsOptions
}

func Execute() error {
	err := newCmdWorkflowStats().Execute()
	if err != nil {
		return err
	}
	return nil
}

func newCmdWorkflowStats() *cobra.Command {
	var (
		host                string
		org                 string
		repo                string
		fileName            string
		id                  int64
		actor               string
		branch              string
		event               string
		status              string
		created             string
		headSHA             string
		excludePullRequests bool
		checkSuiteID        int64
		perPage             int
	)

	cmd := &cobra.Command{
		Use:   "workflow-stats",
		Short: "API to get workflow stats. See https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-workflow",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if host == "" {
				host = defaultHost
			}
			if org == "" || repo == "" {
				return fmt.Errorf("--org and --repo flag must be specified")
			}
			if fileName == "" && id == -1 {
				return fmt.Errorf("--file or --id flag must be specified")
			}
			if perPage > 100 || perPage < 1 {
				perPage = 100
			}
			s := newSpinner(spinnerOptions{
				text:          spinnerText,
				charSetsIndex: 14,
				color:         "green",
			})

			return workflowStats(&workflowStatsConfig{
				RepoInfo: &RepositoryInfo{
					Host:             host,
					Org:              org,
					Repo:             repo,
					WorkflowFileName: fileName,
					ID:               id,
				},
				Params: &github.ListWorkflowRunsOptions{
					Actor:               actor,
					Branch:              branch,
					Event:               event,
					Status:              status,
					Created:             created,
					HeadSHA:             headSHA,
					ExcludePullRequests: excludePullRequests,
					CheckSuiteID:        checkSuiteID,
					ListOptions:         github.ListOptions{PerPage: perPage},
				},
			}, s)
		},
	}
	cmd.Flags().StringVarP(&host, "host", "H", "", "GitHub host. If not specified, default is github.com. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host.")
	cmd.Flags().StringVarP(&org, "org", "o", "", "GitHub organization")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repository")
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "The name of the workflow file. e.g. ci.yaml. You can also pass the workflow id as a integer.")
	cmd.Flags().Int64VarP(&id, "id", "i", -1, "The ID of the workflow. You can also pass the workflow file name as a string.")
	// Workflow runs query parameters
	// See https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-workflow
	cmd.Flags().StringVarP(&actor, "actor", "a", "", "Workflow run actor")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Workflow run branch. Returns workflow runs associated with a branch. Use the name of the branch of the push.")
	cmd.Flags().StringVarP(&event, "event", "e", "", "Workflow run event. e.g. push, pull_request, pull_request_target, etc.\n See https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows")
	cmd.Flags().StringVarP(&status, "status", "s", "", "Workflow run status. e.g. completed, in_progress, queued, etc.\n See https://docs.github.com/en/rest/reference/actions#list-workflow-runs-for-a-repository")
	cmd.Flags().StringVarP(&created, "created", "c", "", "Workflow run createdAt. Returns workflow runs created within the given date-time range.\n For more information on the syntax, see https://docs.github.com/en/search-github/getting-started-with-searching-on-github/understanding-the-search-syntax#query-for-dates")
	cmd.Flags().StringVarP(&headSHA, "head-sha", "S", "", "Workflow run head SHA")
	cmd.Flags().BoolVarP(&excludePullRequests, "exclude-pull-requests", "x", false, "Workflow run exclude pull requests")
	cmd.Flags().Int64VarP(&checkSuiteID, "check-suite-id", "C", 0, "Workflow run check suite ID")
	cmd.Flags().IntVarP(&perPage, "per-page", "p", 100, "GitHub API request per page")

	return cmd
}
