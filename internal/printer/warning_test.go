package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitWarning(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *bytes.Buffer
		expected string
	}{
		{
			name: "Rate limit warning output",
			setup: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
			expected: "\U000026A0  You have reached the rate limit for the GitHub API. These results may not be accurate.\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := tt.setup()
			RateLimitWarning(buf)

			output := buf.String()
			assert.Equal(t, tt.expected, output)

			// Additional assertions
			assert.Contains(t, output, "rate limit")
			assert.Contains(t, output, "GitHub API")
			assert.Contains(t, output, "not be accurate")
			assert.True(t, strings.HasPrefix(output, "\U000026A0"))
			assert.True(t, strings.HasSuffix(output, "\n\n"))
		})
	}
}

func TestRateLimitWarning_MultipleWrites(t *testing.T) {
	buf := &bytes.Buffer{}

	// Write warning multiple times
	RateLimitWarning(buf)
	RateLimitWarning(buf)

	output := buf.String()

	// Should contain the warning twice
	count := strings.Count(output, "rate limit")
	assert.Equal(t, 2, count, "Warning should appear twice")

	// Should end with double newlines
	assert.True(t, strings.HasSuffix(output, "\n\n"))
}
