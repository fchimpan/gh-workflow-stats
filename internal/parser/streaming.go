package parser

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/fchimpan/gh-workflow-stats/internal/types"
	"github.com/google/go-github/v60/github"
)

// StreamingProcessor processes workflow data in streaming fashion to reduce memory usage
type StreamingProcessor struct {
	mu                sync.RWMutex
	runStats          *types.WorkflowRunStats
	jobAggregators    map[string]*jobStreamAggregator
	successDurations  []float64
	calculator        *StatisticsCalculator
	maxMemoryItems    int
}

// jobStreamAggregator accumulates job statistics without storing all raw data
type jobStreamAggregator struct {
	name           string
	totalCount     int
	conclusions    map[types.WorkflowConclusion]int
	durations      []float64
	stepAggregators map[string]*stepStreamAggregator
}

// stepStreamAggregator accumulates step statistics
type stepStreamAggregator struct {
	name        string
	number      int64
	runCount    int
	conclusions map[types.WorkflowConclusion]int
	durations   []float64
	failureURLs []string
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(maxMemoryItems int) *StreamingProcessor {
	if maxMemoryItems <= 0 {
		maxMemoryItems = 10000 // Default memory limit
	}

	return &StreamingProcessor{
		runStats: &types.WorkflowRunStats{
			TotalCount:  0,
			Conclusions: make(map[types.WorkflowConclusion]*types.ConclusionSummary),
		},
		jobAggregators:   make(map[string]*jobStreamAggregator),
		successDurations: make([]float64, 0, maxMemoryItems),
		calculator:       NewStatisticsCalculator(),
		maxMemoryItems:   maxMemoryItems,
	}
}

// ProcessWorkflowRun processes a single workflow run
func (sp *StreamingProcessor) ProcessWorkflowRun(run *github.WorkflowRun) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if run == nil {
		return
	}

	sp.runStats.TotalCount++
	if sp.runStats.Name == "" {
		sp.runStats.Name = run.GetName()
	}

	conclusion := types.NormalizeConclusion(run.GetConclusion())
	
	// Initialize conclusion summary if not exists
	if _, exists := sp.runStats.Conclusions[conclusion]; !exists {
		sp.runStats.Conclusions[conclusion] = &types.ConclusionSummary{
			Count: 0,
			Runs:  []*types.WorkflowRunSummary{},
		}
	}
	
	sp.runStats.Conclusions[conclusion].Count++

	// Only store run details if under memory limit
	if sp.getTotalStoredRuns() < sp.maxMemoryItems {
		summary := sp.convertRunToSummary(run)
		sp.runStats.Conclusions[conclusion].Runs = append(
			sp.runStats.Conclusions[conclusion].Runs, 
			summary,
		)
	}

	// Collect durations for statistics (with memory limit)
	if conclusion == types.ConclusionSuccess && 
	   run.GetStatus() == "completed" && 
	   len(sp.successDurations) < sp.maxMemoryItems {
		duration := sp.calculateRunDuration(run)
		if duration > 0 {
			sp.successDurations = append(sp.successDurations, duration)
		}
	}
}

// ProcessWorkflowJob processes a single workflow job
func (sp *StreamingProcessor) ProcessWorkflowJob(job *github.WorkflowJob) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if job == nil {
		return
	}

	jobName := job.GetName()
	
	// Get or create job aggregator
	if _, exists := sp.jobAggregators[jobName]; !exists {
		sp.jobAggregators[jobName] = &jobStreamAggregator{
			name:            jobName,
			totalCount:      0,
			conclusions:     make(map[types.WorkflowConclusion]int),
			durations:       make([]float64, 0, sp.maxMemoryItems/10), // Smaller buffer for jobs
			stepAggregators: make(map[string]*stepStreamAggregator),
		}
	}

	jobAgg := sp.jobAggregators[jobName]
	jobAgg.totalCount++
	
	conclusion := types.NormalizeConclusion(job.GetConclusion())
	jobAgg.conclusions[conclusion]++

	// Collect job duration (with memory limit)
	if conclusion == types.ConclusionSuccess && 
	   job.GetStatus() == "completed" && 
	   len(jobAgg.durations) < sp.maxMemoryItems/10 {
		duration := sp.calculateJobDuration(job)
		if duration > 0 {
			jobAgg.durations = append(jobAgg.durations, duration)
		}
	}

	// Process steps
	sp.processJobSteps(job, jobAgg)
}

// processJobSteps processes all steps in a job
func (sp *StreamingProcessor) processJobSteps(job *github.WorkflowJob, jobAgg *jobStreamAggregator) {
	for _, step := range job.Steps {
		stepName := step.GetName()
		
		// Get or create step aggregator
		if _, exists := jobAgg.stepAggregators[stepName]; !exists {
			jobAgg.stepAggregators[stepName] = &stepStreamAggregator{
				name:        stepName,
				number:      step.GetNumber(),
				runCount:    0,
				conclusions: make(map[types.WorkflowConclusion]int),
				durations:   make([]float64, 0, sp.maxMemoryItems/50), // Even smaller buffer for steps
				failureURLs: make([]string, 0, 100), // Limit failure URLs
			}
		}

		stepAgg := jobAgg.stepAggregators[stepName]
		stepAgg.runCount++

		conclusion := types.NormalizeConclusion(step.GetConclusion())
		stepAgg.conclusions[conclusion]++

		// Record failure URL (with limit)
		if conclusion == types.ConclusionFailure && 
		   step.GetStatus() == "completed" && 
		   len(stepAgg.failureURLs) < 100 {
			stepAgg.failureURLs = append(stepAgg.failureURLs, job.GetHTMLURL())
		}

		// Collect step duration (with memory limit)
		if step.GetStatus() == "completed" && 
		   (conclusion == types.ConclusionSuccess || conclusion == types.ConclusionFailure) &&
		   len(stepAgg.durations) < sp.maxMemoryItems/50 {
			duration := sp.calculateStepDuration(step)
			if duration >= 0 {
				stepAgg.durations = append(stepAgg.durations, duration)
			}
		}
	}
}

// ProcessBatch processes multiple workflow runs in batch
func (sp *StreamingProcessor) ProcessBatch(ctx context.Context, runs []*github.WorkflowRun, jobs []*github.WorkflowJob) error {
	// Process runs
	for _, run := range runs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			sp.ProcessWorkflowRun(run)
		}
	}

	// Process jobs
	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			sp.ProcessWorkflowJob(job)
		}
	}

	return nil
}

// GetResults returns the final processed results
func (sp *StreamingProcessor) GetResults() (*types.WorkflowRunStats, []*types.WorkflowJobStats) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Finalize run stats
	sp.finalizeRunStats()

	// Convert job aggregators to job stats
	jobStats := sp.convertJobAggregators()

	return sp.runStats, jobStats
}

// finalizeRunStats calculates final statistics for runs
func (sp *StreamingProcessor) finalizeRunStats() {
	// Calculate success rates
	totalCount := float64(sp.runStats.TotalCount)
	if totalCount > 0 {
		sp.runStats.SuccessRate = types.SuccessRate{
			Success: float64(sp.runStats.Conclusions[types.ConclusionSuccess].Count) / totalCount,
			Failure: float64(sp.runStats.Conclusions[types.ConclusionFailure].Count) / totalCount,
			Others:  float64(sp.runStats.Conclusions[types.ConclusionOthers].Count) / totalCount,
		}
	}

	// Calculate execution statistics
	sp.runStats.ExecutionStats = sp.calculator.CalculateExecutionStats(sp.successDurations)
}

// convertJobAggregators converts job aggregators to final job stats
func (sp *StreamingProcessor) convertJobAggregators() []*types.WorkflowJobStats {
	var jobStats []*types.WorkflowJobStats

	for _, jobAgg := range sp.jobAggregators {
		stats := &types.WorkflowJobStats{
			Name:           jobAgg.name,
			TotalCount:     jobAgg.totalCount,
			Conclusions:    jobAgg.conclusions,
			ExecutionStats: sp.calculator.CalculateExecutionStats(jobAgg.durations),
			Steps:          sp.convertStepAggregators(jobAgg.stepAggregators),
		}

		// Calculate success rate
		stats.SuccessRate = sp.calculator.CalculateSuccessRate(jobAgg.conclusions)

		jobStats = append(jobStats, stats)
	}

	return jobStats
}

// convertStepAggregators converts step aggregators to step stats
func (sp *StreamingProcessor) convertStepAggregators(stepAggregators map[string]*stepStreamAggregator) []*types.StepStats {
	var steps []*types.StepStats

	for _, stepAgg := range stepAggregators {
		stepStats := &types.StepStats{
			Name:           stepAgg.name,
			Number:         stepAgg.number,
			RunCount:       stepAgg.runCount,
			Conclusions:    stepAgg.conclusions,
			ExecutionStats: sp.calculator.CalculateExecutionStats(stepAgg.durations),
			FailureURLs:    stepAgg.failureURLs,
		}

		// Calculate success rate
		stepStats.SuccessRate = sp.calculator.CalculateSuccessRate(stepAgg.conclusions)

		steps = append(steps, stepStats)
	}

	// Sort steps by number
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Number < steps[j].Number
	})

	return steps
}

// getTotalStoredRuns returns total number of stored run summaries
func (sp *StreamingProcessor) getTotalStoredRuns() int {
	total := 0
	for _, conclusion := range sp.runStats.Conclusions {
		total += len(conclusion.Runs)
	}
	return total
}

// Helper methods for duration calculations
func (sp *StreamingProcessor) convertRunToSummary(run *github.WorkflowRun) *types.WorkflowRunSummary {
	return &types.WorkflowRunSummary{
		ID:           run.GetID(),
		Status:       types.WorkflowStatus(run.GetStatus()),
		Conclusion:   types.WorkflowConclusion(run.GetConclusion()),
		Actor:        run.GetActor().GetLogin(),
		RunAttempt:   run.GetRunAttempt(),
		HTMLURL:      sp.buildRunURL(run),
		JobsURL:      run.GetJobsURL(),
		LogsURL:      run.GetLogsURL(),
		RunStartedAt: run.GetRunStartedAt().Time.UTC(),
		UpdatedAt:    run.GetUpdatedAt().Time.UTC(),
		CreatedAt:    run.GetCreatedAt().Time.UTC(),
		Duration:     sp.calculateRunDuration(run),
	}
}

func (sp *StreamingProcessor) buildRunURL(run *github.WorkflowRun) string {
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

func (sp *StreamingProcessor) calculateRunDuration(run *github.WorkflowRun) float64 {
	startTime := run.GetRunStartedAt().Time
	endTime := run.GetUpdatedAt().Time
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime).Seconds()
	return sp.calculator.ValidateDuration(duration)
}

func (sp *StreamingProcessor) calculateJobDuration(job *github.WorkflowJob) float64 {
	startTime := job.GetStartedAt().Time
	endTime := job.GetCompletedAt().Time
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime).Seconds()
	return max(duration, 0)
}

func (sp *StreamingProcessor) calculateStepDuration(step *github.TaskStep) float64 {
	startTime := step.GetStartedAt().Time
	endTime := step.GetCompletedAt().Time
	
	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}
	
	duration := endTime.Sub(startTime).Seconds()
	return max(duration, 0)
}

// GetMemoryUsage returns current memory usage statistics
func (sp *StreamingProcessor) GetMemoryUsage() MemoryUsage {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	usage := MemoryUsage{
		StoredRuns:        sp.getTotalStoredRuns(),
		SuccessDurations:  len(sp.successDurations),
		JobAggregators:    len(sp.jobAggregators),
		MaxMemoryItems:    sp.maxMemoryItems,
	}

	for _, jobAgg := range sp.jobAggregators {
		usage.JobDurations += len(jobAgg.durations)
		usage.StepAggregators += len(jobAgg.stepAggregators)
		
		for _, stepAgg := range jobAgg.stepAggregators {
			usage.StepDurations += len(stepAgg.durations)
			usage.FailureURLs += len(stepAgg.failureURLs)
		}
	}

	return usage
}

// MemoryUsage provides information about current memory usage
type MemoryUsage struct {
	StoredRuns       int
	SuccessDurations int
	JobAggregators   int
	StepAggregators  int
	JobDurations     int
	StepDurations    int
	FailureURLs      int
	MaxMemoryItems   int
}