package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSpinner(t *testing.T) {
	tests := []struct {
		name      string
		options   SpinnerOptions
		shouldErr bool
	}{
		{
			name: "Valid spinner options",
			options: SpinnerOptions{
				Text:          "Loading...",
				CharSetsIndex: 14,
				Color:         "cyan",
			},
			shouldErr: false,
		},
		{
			name: "Another valid configuration",
			options: SpinnerOptions{
				Text:          "Processing...",
				CharSetsIndex: 11,
				Color:         "green",
			},
			shouldErr: false,
		},
		{
			name: "Empty text",
			options: SpinnerOptions{
				Text:          "",
				CharSetsIndex: 14,
				Color:         "cyan",
			},
			shouldErr: false,
		},
		{
			name: "Different charset index",
			options: SpinnerOptions{
				Text:          "Working...",
				CharSetsIndex: 1,
				Color:         "yellow",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner, err := NewSpinner(tt.options)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, spinner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, spinner)
			}
		})
	}
}

func TestSpinnerOptions(t *testing.T) {
	opt := SpinnerOptions{
		Text:          "Test text",
		CharSetsIndex: 14,
		Color:         "blue",
	}

	assert.Equal(t, "Test text", opt.Text)
	assert.Equal(t, 14, opt.CharSetsIndex)
	assert.Equal(t, "blue", opt.Color)
}

func TestSpinnerOperations(t *testing.T) {
	// Test that spinner operations don't panic
	opt := SpinnerOptions{
		Text:          "Test...",
		CharSetsIndex: 14,
		Color:         "cyan",
	}

	spinner, err := NewSpinner(opt)
	assert.NoError(t, err)
	assert.NotNil(t, spinner)

	// Test Start - should not panic
	assert.NotPanics(t, func() {
		spinner.Start()
	})

	// Test Update - should not panic
	updateOpt := SpinnerOptions{
		Text:          "Updated text...",
		CharSetsIndex: 14,
		Color:         "cyan",
	}
	assert.NotPanics(t, func() {
		spinner.Update(updateOpt)
	})

	// Test Stop - should not panic
	assert.NotPanics(t, func() {
		spinner.Stop()
	})
}
