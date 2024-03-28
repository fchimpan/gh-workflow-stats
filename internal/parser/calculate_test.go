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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcStats(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
