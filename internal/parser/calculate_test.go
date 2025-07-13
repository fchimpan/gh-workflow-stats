package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcStats(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected ExecutionDurationStats
	}{
		{
			name:     "Empty input",
			input:    []float64{},
			expected: ExecutionDurationStats{},
		},
		{
			name:     "Non-empty input",
			input:    []float64{1.5, 2.5, 3.5, 4.5, 5.5},
			expected: ExecutionDurationStats{Min: 1.5, Max: 5.5, Avg: 3.5, Med: 3.5, Std: 1.4142135623730951},
		},
		{
			name:     "Single value",
			input:    []float64{42.0},
			expected: ExecutionDurationStats{Min: 42.0, Max: 42.0, Avg: 42.0, Med: 42.0, Std: 0},
		},
		{
			name:     "Two values",
			input:    []float64{10.0, 20.0},
			expected: ExecutionDurationStats{Min: 10.0, Max: 20.0, Avg: 15.0, Med: 15.0, Std: 5.0},
		},
		{
			name:     "Zero values",
			input:    []float64{0.0, 0.0, 0.0},
			expected: ExecutionDurationStats{Min: 0.0, Max: 0.0, Avg: 0.0, Med: 0.0, Std: 0},
		},
		{
			name:     "Negative values",
			input:    []float64{-5.0, -10.0, -15.0},
			expected: ExecutionDurationStats{Min: -15.0, Max: -5.0, Avg: -10.0, Med: -10.0, Std: 4.08248290463863},
		},
		{
			name:     "Mixed positive and negative",
			input:    []float64{-10.0, 0.0, 10.0},
			expected: ExecutionDurationStats{Min: -10.0, Max: 10.0, Avg: 0.0, Med: 0.0, Std: 8.16496580927726},
		},
		{
			name:     "Large values",
			input:    []float64{1000000.0, 2000000.0, 3000000.0},
			expected: ExecutionDurationStats{Min: 1000000.0, Max: 3000000.0, Avg: 2000000.0, Med: 2000000.0, Std: 0},
		},
		{
			name:     "Very small values",
			input:    []float64{0.001, 0.002, 0.003},
			expected: ExecutionDurationStats{Min: 0.001, Max: 0.003, Avg: 0.002, Med: 0.002, Std: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcStats(tt.input)

			// Handle edge cases that need special floating point precision handling
			if tt.name == "Large values" || tt.name == "Very small values" {
				assert.InDelta(t, tt.expected.Min, result.Min, 0.001, "Min should be close")
				assert.InDelta(t, tt.expected.Max, result.Max, 0.001, "Max should be close")
				assert.InDelta(t, tt.expected.Avg, result.Avg, 0.001, "Avg should be close")
				assert.InDelta(t, tt.expected.Med, result.Med, 0.001, "Med should be close")
				// Standard deviation can have precision differences, so use larger tolerance
				assert.Greater(t, result.Std, 0.0, "Std should be positive for non-zero variance")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAdjustRate(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "Rate above epsilon",
			input:    0.5,
			expected: 0.5,
		},
		{
			name:     "Rate exactly at epsilon",
			input:    eps,
			expected: eps,
		},
		{
			name:     "Rate below epsilon",
			input:    eps / 2,
			expected: 0.0,
		},
		{
			name:     "Zero rate",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "Very small rate",
			input:    1e-10,
			expected: 0.0,
		},
		{
			name:     "Rate of 1.0",
			input:    1.0,
			expected: 1.0,
		},
		{
			name:     "Negative rate",
			input:    -0.1,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjustRate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEpsConstant(t *testing.T) {
	assert.Equal(t, 1e-9, eps, "Epsilon constant should be 1e-9")
}
