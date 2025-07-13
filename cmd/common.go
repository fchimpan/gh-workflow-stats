package cmd

import (
	"fmt"
	"os"

	"github.com/fchimpan/gh-workflow-stats/internal/errors"
	"github.com/fchimpan/gh-workflow-stats/internal/types"
	"github.com/spf13/cobra"
)

// Error messages as constants to avoid duplication
const (
	ErrMissingOrgRepo  = "--org and --repo flag must be specified. If you want to use GitHub Enterprise Server, specify your GitHub Enterprise Server host with --host flag"
	ErrMissingWorkflow = "--file or --id flag must be specified"
)

// validateFlags validates common flags across commands
func validateFlags(org, repo, fileName string, id int64) error {
	if org == "" || repo == "" {
		return errors.NewConfigurationError(ErrMissingOrgRepo, nil).
			WithContext("org", org).
			WithContext("repo", repo)
	}
	if fileName == "" && id == -1 {
		return errors.NewConfigurationError(ErrMissingWorkflow, nil).
			WithContext("workflow_file", fileName).
			WithContext("workflow_id", fmt.Sprintf("%d", id))
	}
	return nil
}

// resolveHost resolves the host from environment variable if not set via flag
func resolveHost(cmd *cobra.Command, host *string) {
	if envHost := os.Getenv("GH_HOST"); envHost != "" && !cmd.Flags().Changed("host") {
		*host = envHost
	}
}

// createWorkflowConfig creates a WorkflowConfig from command flags
func createWorkflowConfig(host, org, repo, fileName string, id int64) *types.WorkflowConfig {
	return &types.WorkflowConfig{
		Host:             host,
		Org:              org,
		Repo:             repo,
		WorkflowFileName: fileName,
		WorkflowID:       id,
	}
}

// createWorkflowFetchOptions creates WorkflowFetchOptions from command flags
func createWorkflowFetchOptions(actor, branch, event string, status []string, created, headSHA string,
	excludePullRequests, all, js bool, checkSuiteID int64, jobNum int) *types.WorkflowFetchOptions {
	
	if jobNum <= 0 {
		jobNum = types.DefaultJobCount
	}
	
	return &types.WorkflowFetchOptions{
		Actor:               actor,
		Branch:              branch,
		Event:               event,
		Status:              status,
		Created:             created,
		HeadSHA:             headSHA,
		ExcludePullRequests: excludePullRequests,
		CheckSuiteID:        checkSuiteID,
		All:                 all,
		OutputJSON:          js,
		JobCount:            jobNum,
	}
}

// Legacy functions for backward compatibility
// These will be removed in a future version

// createConfig creates a legacy config struct from command flags
func createConfig(host, org, repo, fileName string, id int64) config {
	return config{
		host:             host,
		org:              org,
		repo:             repo,
		workflowFileName: fileName,
		workflowID:       id,
	}
}

// createOptions creates a legacy options struct from command flags
func createOptions(actor, branch, event string, status []string, created, headSHA string,
	excludePullRequests, all, js bool, checkSuiteID int64, jobNum int) options {
	
	if jobNum <= 0 {
		jobNum = types.DefaultJobCount
	}
	
	return options{
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
		jobNum:              jobNum,
	}
}
