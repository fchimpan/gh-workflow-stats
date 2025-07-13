package cmd

import (
	"github.com/fchimpan/gh-workflow-stats/internal/types"
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
		resolveHost(cmd, &host)

		if err := validateFlags(org, repo, fileName, id); err != nil {
			return err
		}

		if numJobs < 1 {
			numJobs = 1
		}

		cfg := createConfig(host, org, repo, fileName, id)
		opts := createOptions(actor, branch, event, status, created, headSHA,
			excludePullRequests, all, js, checkSuiteID, numJobs)

		return workflowStats(cfg, opts, true)
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.Flags().IntVarP(&numJobs, "num-jobs", "n", types.DefaultJobCount, "Number of jobs to display")
}
