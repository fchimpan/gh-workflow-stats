package types

import "time"

const (
	// Duration constants
	MaxWorkflowDurationSeconds = 35 * 24 * 60 * 60 // 35 days in seconds (GitHub Actions limit)
	MaxWorkflowDurationCapped  = 3024000           // Capped duration value for invalid durations

	// GitHub API constants
	DefaultPerPage  = 100
	MaxPerPage      = 100
	DefaultJobCount = 3

	// Time format constants
	GitHubTimeFormat = time.RFC3339

	// Rate limiting
	DefaultRetryAttempts = 3
	DefaultRetryDelay    = time.Second * 5

	// File extensions
	WorkflowFileExtension = ".yml"
	YAMLFileExtension     = ".yaml"
)
