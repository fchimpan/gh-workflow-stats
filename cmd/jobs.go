package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	numJobs int
)

var jobsCmd = &cobra.Command{
	Use:     "jobs",
	Short:   "Fetch workflow jobs stats. Retrieve the steps and jobs success rate.",
	Example: `$ gh workflow-stats jobs --org=OWNER --repo=REPO --id=WORKFLOW_ID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if envHost := os.Getenv("GH_HOST"); envHost != "" && !cmd.Flags().Changed("host") {
			host = envHost
		}
		if org == "" || repo == "" {
			return fmt.Errorf("--org and --repo flag must be specified. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host with --host flag")
		}
		if fileName == "" && id == -1 {
			return fmt.Errorf("--file or --id flag must be specified")
		}
		if numJobs < 1 {
			numJobs = 1
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
			jobNum:              numJobs,
		}, true)
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.Flags().IntVarP(&numJobs, "num-jobs", "n", 3, "Number of jobs to display")
}
