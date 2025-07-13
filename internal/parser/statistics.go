package parser

import (
	"math"
	"sort"

	"github.com/fchimpan/gh-workflow-stats/internal/types"
)

// calculateExecutionStats calculates statistical data from a slice of durations
func calculateExecutionStats(durations []float64) types.ExecutionStats {
	if len(durations) == 0 {
		return types.ExecutionStats{Count: 0}
	}

	// Make a copy and sort for percentile calculations
	sorted := make([]float64, len(durations))
	copy(sorted, durations)
	sort.Float64s(sorted)

	stats := types.ExecutionStats{
		Count:  len(durations),
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Mean:   calculateMean(durations),
		Median: calculatePercentile(sorted, 50),
		P95:    calculatePercentile(sorted, 95),
		P99:    calculatePercentile(sorted, 99),
	}

	stats.StdDev = calculateStandardDeviation(durations, stats.Mean)

	return stats
}

// calculateMean calculates the arithmetic mean
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculatePercentile calculates the specified percentile from sorted data
func calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	if percentile <= 0 {
		return sortedValues[0]
	}
	if percentile >= 100 {
		return sortedValues[len(sortedValues)-1]
	}

	// Linear interpolation method
	index := (percentile / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	// Interpolate between lower and upper values
	weight := index - math.Floor(index)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// calculateStandardDeviation calculates the standard deviation
func calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values)-1) // Sample standard deviation
	return math.Sqrt(variance)
}

// adjustRatePrecise ensures rates are properly rounded and within valid bounds
func adjustRatePrecise(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	
	// Round to 6 decimal places to avoid floating point precision issues
	return math.Round(rate*1000000) / 1000000
}

// StatisticsCalculator provides methods for calculating various statistics
type StatisticsCalculator struct{}

// NewStatisticsCalculator creates a new statistics calculator
func NewStatisticsCalculator() *StatisticsCalculator {
	return &StatisticsCalculator{}
}

// CalculateSuccessRate calculates success rate from conclusion counts
func (c *StatisticsCalculator) CalculateSuccessRate(conclusions map[types.WorkflowConclusion]int) types.SuccessRate {
	total := 0
	for _, count := range conclusions {
		total += count
	}

	if total == 0 {
		return types.SuccessRate{}
	}

	totalFloat := float64(total)
	return types.SuccessRate{
		Success: adjustRatePrecise(float64(conclusions[types.ConclusionSuccess]) / totalFloat),
		Failure: adjustRatePrecise(float64(conclusions[types.ConclusionFailure]) / totalFloat),
		Others:  adjustRatePrecise(float64(conclusions[types.ConclusionOthers]) / totalFloat),
	}
}

// CalculateExecutionStats is a wrapper around calculateExecutionStats for external use
func (c *StatisticsCalculator) CalculateExecutionStats(durations []float64) types.ExecutionStats {
	return calculateExecutionStats(durations)
}

// ValidateDuration validates and normalizes duration values
func (c *StatisticsCalculator) ValidateDuration(duration float64) float64 {
	if duration < 0 {
		return 0
	}
	if duration > types.MaxWorkflowDurationSeconds {
		return types.MaxWorkflowDurationCapped
	}
	return duration
}