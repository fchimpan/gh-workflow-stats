package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	host                string
	org                 string
	repo                string
	fileName            string
	id                  int64
	all                 bool
	js                  bool
	actor               string
	branch              string
	event               string
	status              []string
	created             string
	headSHA             string
	excludePullRequests bool
	checkSuiteID        int64
)

var rootCmd = &cobra.Command{
	Use:     "workflow-stats",
	Short:   "Fetch workflow runs stats. Retrieve the success rate and execution time of workflows.",
	Example: `$ gh workflow-stats --org $OWNER --repo $REPO -f ci.yaml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if envHost := os.Getenv("GH_HOST"); envHost != "" && !cmd.Flags().Changed("host") {
			host = envHost
		}
		if org == "" || repo == "" {
			return fmt.Errorf("--org and --repo flag must be specified. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host with --host flag")
		}
		if fileName == "" && id == -1 {
			return fmt.Errorf("--file or --id flag must be specified")
		}
		return workflowStats(config{
			host:             host,
			org:              org,
			repo:             repo,
			workflowFileName: fileName,
			workflowID:       id,
		}, options{
			actor:               actor,
			branch:              branch,
			event:               event,
			status:              status,
			created:             created,
			headSHA:             headSHA,
			excludePullRequests: excludePullRequests,
			checkSuiteID:        checkSuiteID,
			all:                 all,
			js:                  js,
		}, false)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "github.com", "GitHub host. If not specified, default is github.com. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host.")
	rootCmd.PersistentFlags().StringVarP(&org, "org", "o", "", "GitHub organization")
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "r", "", "GitHub repository")
	rootCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "The name of the workflow file. e.g. ci.yaml. You can also pass the workflow id as a integer.")
	rootCmd.PersistentFlags().Int64VarP(&id, "id", "i", -1, "The ID of the workflow. You can also pass the workflow file name as a string.")
	rootCmd.PersistentFlags().BoolVarP(&all, "all", "A", false, "Target all workflows in the repository. If specified, default fetches of 100 workflow runs is overridden to all workflow runs. Note the GitHub API rate limit.")
	rootCmd.PersistentFlags().BoolVar(&js, "json", false, "Output as JSON")

	// Workflow runs query parameters
	// See https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-workflow
	rootCmd.PersistentFlags().StringVarP(&actor, "actor", "a", "", "Workflow run actor")
	rootCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Workflow run branch. Returns workflow runs associated with a branch. Use the name of the branch of the push.")
	rootCmd.PersistentFlags().StringVarP(&event, "event", "e", "", "Workflow run event. e.g. push, pull_request, pull_request_target, etc.\n See https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows")
	rootCmd.PersistentFlags().StringSliceVarP(&status, "status", "s", []string{""}, "Workflow run status. e.g. completed, in_progress, queued, etc.\n Multiple values can be provided separated by a comma. For a full list of supported values see https://docs.github.com/en/rest/reference/actions#list-workflow-runs-for-a-repository")
	rootCmd.PersistentFlags().StringVarP(&created, "created", "c", "", "Workflow run createdAt. Returns workflow runs created within the given date-time range.\n For more information on the syntax, see https://docs.github.com/en/search-github/getting-started-with-searching-on-github/understanding-the-search-syntax#query-for-dates")
	rootCmd.PersistentFlags().StringVarP(&headSHA, "head-sha", "S", "", "Workflow run head SHA")
	rootCmd.PersistentFlags().BoolVarP(&excludePullRequests, "exclude-pull-requests", "x", false, "Workflow run exclude pull requests")
	rootCmd.PersistentFlags().Int64VarP(&checkSuiteID, "check-suite-id", "C", 0, "Workflow run check suite ID")
}
