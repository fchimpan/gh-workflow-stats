package parser

import (
	"sort"
	"strconv"

	"github.com/fchimpan/gh-workflow-stats/internal/types"
	"github.com/google/go-github/v60/github"
)

// WorkflowRunConverter converts GitHub API workflow runs to internal types
type WorkflowRunConverter struct{}

// NewWorkflowRunConverter creates a new converter instance
func NewWorkflowRunConverter() *WorkflowRunConverter {
	return &WorkflowRunConverter{}
}

// ConvertRuns converts GitHub workflow runs to WorkflowRunStats
func (c *WorkflowRunConverter) ConvertRuns(githubRuns []*github.WorkflowRun) *types.WorkflowRunStats {
	if len(githubRuns) == 0 {
		return &types.WorkflowRunStats{
			TotalCount:  0,
			Conclusions: make(map[types.WorkflowConclusion]*types.ConclusionSummary),
		}
	}

	stats := &types.WorkflowRunStats{
		Name:        githubRuns[0].GetName(),
		TotalCount:  len(githubRuns),
		Conclusions: map[types.WorkflowConclusion]*types.ConclusionSummary{
			types.ConclusionSuccess: {Count: 0, Runs: []*types.WorkflowRunSummary{}},
			types.ConclusionFailure: {Count: 0, Runs: []*types.WorkflowRunSummary{}},
			types.ConclusionOthers:  {Count: 0, Runs: []*types.WorkflowRunSummary{}},
		},
	}

	var successDurations []float64
	
	for _, run := range githubRuns {
		summary := c.convertSingleRun(run)
		conclusion := types.NormalizeConclusion(summary.Conclusion.String())
		
		stats.Conclusions[conclusion].Count++
		stats.Conclusions[conclusion].Runs = append(stats.Conclusions[conclusion].Runs, summary)
		
		// Collect durations for successful, completed runs for statistics
		if conclusion == types.ConclusionSuccess && 
		   summary.Status == types.StatusCompleted && 
		   summary.Duration > 0 {
			successDurations = append(successDurations, summary.Duration)
		}
	}

	// Calculate rates
	totalCount := float64(stats.TotalCount)
	if totalCount > 0 {
		stats.SuccessRate = types.SuccessRate{
			Success: float64(stats.Conclusions[types.ConclusionSuccess].Count) / totalCount,
			Failure: float64(stats.Conclusions[types.ConclusionFailure].Count) / totalCount,
			Others:  float64(stats.Conclusions[types.ConclusionOthers].Count) / totalCount,
		}
	}

	// Calculate execution statistics
	stats.ExecutionStats = calculateExecutionStats(successDurations)

	return stats
}

// convertSingleRun converts a single GitHub workflow run to internal type
func (c *WorkflowRunConverter) convertSingleRun(run *github.WorkflowRun) *types.WorkflowRunSummary {
	summary := &types.WorkflowRunSummary{
		ID:           run.GetID(),
		Status:       types.WorkflowStatus(run.GetStatus()),
		Conclusion:   types.WorkflowConclusion(run.GetConclusion()),
		Actor:        run.GetActor().GetLogin(),
		RunAttempt:   run.GetRunAttempt(),
		HTMLURL:      c.buildRunURL(run),
		JobsURL:      run.GetJobsURL(),
		LogsURL:      run.GetLogsURL(),
		RunStartedAt: run.GetRunStartedAt().UTC(),
		UpdatedAt:    run.GetUpdatedAt().UTC(),
		CreatedAt:    run.GetCreatedAt().UTC(),
	}

	// Calculate duration
	summary.Duration = c.calculateDuration(run)

	return summary
}

// buildRunURL builds the HTML URL with attempt number
func (c *WorkflowRunConverter) buildRunURL(run *github.WorkflowRun) string {
	baseURL := run.GetHTMLURL()
	if baseURL == "" {
		return ""
	}
	
	attempt := run.GetRunAttempt()
	if attempt > 1 {
		return baseURL + "/attempts/" + strconv.Itoa(attempt)
	}
	
	return baseURL
}

// calculateDuration calculates run duration with bounds checking
func (c *WorkflowRunConverter) calculateDuration(run *github.WorkflowRun) float64 {
	startTime := run.GetRunStartedAt()
	endTime := run.GetUpdatedAt()
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime.Time).Seconds()
	
	// Apply bounds checking
	if duration < 0 {
		return 0
	}
	
	if duration > types.MaxWorkflowDurationSeconds {
		return types.MaxWorkflowDurationCapped
	}
	
	return duration
}

// WorkflowJobConverter converts GitHub API workflow jobs to internal types
type WorkflowJobConverter struct{}

// NewWorkflowJobConverter creates a new converter instance
func NewWorkflowJobConverter() *WorkflowJobConverter {
	return &WorkflowJobConverter{}
}

// ConvertJobs converts GitHub workflow jobs to WorkflowJobStats
func (c *WorkflowJobConverter) ConvertJobs(githubJobs []*github.WorkflowJob) []*types.WorkflowJobStats {
	if len(githubJobs) == 0 {
		return []*types.WorkflowJobStats{}
	}

	// Group jobs by name
	jobGroups := c.groupJobsByName(githubJobs)
	
	var result []*types.WorkflowJobStats
	for jobName, jobs := range jobGroups {
		jobStats := c.convertJobGroup(jobName, jobs)
		result = append(result, jobStats)
	}

	return result
}

// groupJobsByName groups jobs by their name
func (c *WorkflowJobConverter) groupJobsByName(jobs []*github.WorkflowJob) map[string][]*github.WorkflowJob {
	groups := make(map[string][]*github.WorkflowJob)
	
	for _, job := range jobs {
		name := job.GetName()
		groups[name] = append(groups[name], job)
	}
	
	return groups
}

// convertJobGroup converts a group of jobs with the same name to JobStats
func (c *WorkflowJobConverter) convertJobGroup(name string, jobs []*github.WorkflowJob) *types.WorkflowJobStats {
	stats := &types.WorkflowJobStats{
		Name:        name,
		TotalCount:  len(jobs),
		Conclusions: make(map[types.WorkflowConclusion]int),
		Steps:       []*types.StepStats{},
	}

	// Initialize conclusions count
	stats.Conclusions[types.ConclusionSuccess] = 0
	stats.Conclusions[types.ConclusionFailure] = 0
	stats.Conclusions[types.ConclusionOthers] = 0

	var successDurations []float64
	stepGroups := make(map[string]*stepAggregator)

	// Process each job
	for _, job := range jobs {
		conclusion := types.NormalizeConclusion(job.GetConclusion())
		stats.Conclusions[conclusion]++

		// Collect successful job durations
		if conclusion == types.ConclusionSuccess && job.GetStatus() == "completed" {
			duration := c.calculateJobDuration(job)
			if duration > 0 {
				successDurations = append(successDurations, duration)
			}
		}

		// Process steps
		c.processJobSteps(job, stepGroups)
	}

	// Calculate rates
	totalCount := float64(stats.TotalCount)
	if totalCount > 0 {
		stats.SuccessRate = types.SuccessRate{
			Success: float64(stats.Conclusions[types.ConclusionSuccess]) / totalCount,
			Failure: float64(stats.Conclusions[types.ConclusionFailure]) / totalCount,
			Others:  float64(stats.Conclusions[types.ConclusionOthers]) / totalCount,
		}
	}

	// Calculate execution statistics
	stats.ExecutionStats = calculateExecutionStats(successDurations)

	// Convert step aggregators to step stats
	stats.Steps = c.convertStepAggregators(stepGroups)

	return stats
}

// stepAggregator accumulates data for steps with the same name
type stepAggregator struct {
	name        string
	number      int64
	runCount    int
	conclusions map[types.WorkflowConclusion]int
	durations   []float64
	failureURLs []string
}

// processJobSteps processes all steps in a job
func (c *WorkflowJobConverter) processJobSteps(job *github.WorkflowJob, stepGroups map[string]*stepAggregator) {
	for _, step := range job.Steps {
		stepName := step.GetName()
		
		if _, exists := stepGroups[stepName]; !exists {
			stepGroups[stepName] = &stepAggregator{
				name:        stepName,
				number:      step.GetNumber(),
				runCount:    0,
				conclusions: make(map[types.WorkflowConclusion]int),
				durations:   []float64{},
				failureURLs: []string{},
			}
		}

		aggregator := stepGroups[stepName]
		aggregator.runCount++

		conclusion := types.NormalizeConclusion(step.GetConclusion())
		aggregator.conclusions[conclusion]++

		// Record failure URL
		if conclusion == types.ConclusionFailure && step.GetStatus() == "completed" {
			aggregator.failureURLs = append(aggregator.failureURLs, job.GetHTMLURL())
		}

		// Calculate step duration
		if step.GetStatus() == "completed" && (conclusion == types.ConclusionSuccess || conclusion == types.ConclusionFailure) {
			duration := c.calculateStepDuration(step)
			if duration >= 0 {
				aggregator.durations = append(aggregator.durations, duration)
			}
		}
	}
}

// convertStepAggregators converts step aggregators to StepStats
func (c *WorkflowJobConverter) convertStepAggregators(stepGroups map[string]*stepAggregator) []*types.StepStats {
	var steps []*types.StepStats
	
	for _, aggregator := range stepGroups {
		stepStats := &types.StepStats{
			Name:        aggregator.name,
			Number:      aggregator.number,
			RunCount:    aggregator.runCount,
			Conclusions: aggregator.conclusions,
			FailureURLs: aggregator.failureURLs,
		}

		// Calculate success rate
		runCount := float64(aggregator.runCount)
		if runCount > 0 {
			stepStats.SuccessRate = types.SuccessRate{
				Success: float64(aggregator.conclusions[types.ConclusionSuccess]) / runCount,
				Failure: float64(aggregator.conclusions[types.ConclusionFailure]) / runCount,
				Others:  float64(aggregator.conclusions[types.ConclusionOthers]) / runCount,
			}
		}

		// Calculate execution statistics
		stepStats.ExecutionStats = calculateExecutionStats(aggregator.durations)

		steps = append(steps, stepStats)
	}

	// Sort steps by number efficiently using sort.Slice (O(n log n))
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Number < steps[j].Number
	})

	return steps
}

// calculateJobDuration calculates the duration of a job
func (c *WorkflowJobConverter) calculateJobDuration(job *github.WorkflowJob) float64 {
	startTime := job.GetStartedAt()
	endTime := job.GetCompletedAt()
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime.Time).Seconds()
	return max(duration, 0)
}

// calculateStepDuration calculates the duration of a step
func (c *WorkflowJobConverter) calculateStepDuration(step *github.TaskStep) float64 {
	startTime := step.GetStartedAt()
	endTime := step.GetCompletedAt()
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime.Time).Seconds()
	return max(duration, 0)
}