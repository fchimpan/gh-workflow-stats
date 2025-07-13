package types

import "errors"

var (
	// Configuration errors
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrMissingOrgRepo      = errors.New("organization and repository must be specified")
	ErrMissingWorkflow     = errors.New("workflow file or ID must be specified")
	ErrInvalidWorkflowID   = errors.New("invalid workflow ID")
	ErrInvalidTimeRange    = errors.New("invalid time range: start time must be before end time")
	
	// Validation errors
	ErrInvalidStatus       = errors.New("invalid workflow status")
	ErrInvalidConclusion   = errors.New("invalid workflow conclusion")
	ErrInvalidOutputFormat = errors.New("invalid output format")
	ErrInvalidJobCount     = errors.New("invalid job count: must be positive")
	
	// Data processing errors
	ErrEmptyWorkflowRuns   = errors.New("no workflow runs found")
	ErrEmptyWorkflowJobs   = errors.New("no workflow jobs found")
	ErrInvalidDuration     = errors.New("invalid duration value")
)