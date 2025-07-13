package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  slog.Level
		output *bytes.Buffer
	}{
		{
			name:   "Debug level logger",
			level:  LevelDebug,
			output: &bytes.Buffer{},
		},
		{
			name:   "Info level logger",
			level:  LevelInfo,
			output: &bytes.Buffer{},
		},
		{
			name:   "Error level logger",
			level:  LevelError,
			output: &bytes.Buffer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level, tt.output)
			assert.NotNil(t, logger)
			assert.IsType(t, &StructuredLogger{}, logger)
		})
	}
}

func TestNewJSONLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(LevelInfo, buf)

	assert.NotNil(t, logger)
	assert.IsType(t, &StructuredLogger{}, logger)
}

func TestStructuredLogger_Logging(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(LevelDebug, buf)

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Debug message",
			logFunc: func() {
				logger.Debug("debug message", "key", "value")
			},
			expected: "debug message",
		},
		{
			name: "Info message",
			logFunc: func() {
				logger.Info("info message", "count", 42)
			},
			expected: "info message",
		},
		{
			name: "Warn message",
			logFunc: func() {
				logger.Warn("warning message", "status", "degraded")
			},
			expected: "warning message",
		},
		{
			name: "Error message",
			logFunc: func() {
				logger.Error("error message", "error_code", "E001")
			},
			expected: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestStructuredLogger_With(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(LevelInfo, buf)

	contextLogger := logger.With("service", "test", "version", "1.0")
	assert.NotNil(t, contextLogger)
	assert.IsType(t, &StructuredLogger{}, contextLogger)

	contextLogger.Info("test message")
	output := buf.String()

	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "service")
	assert.Contains(t, output, "test")
}

func TestJSONLogger_Output(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(LevelInfo, buf)

	logger.Info("json test", "field", "value", "number", 123)
	output := buf.String()

	// Should be valid JSON format
	assert.Contains(t, output, `"msg":"json test"`)
	assert.Contains(t, output, `"field":"value"`)
	assert.Contains(t, output, `"number":123`)
}

func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()
	assert.NotNil(t, logger)
	assert.IsType(t, &NoOpLogger{}, logger)

	// These should not panic and do nothing
	assert.NotPanics(t, func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error")
	})

	contextLogger := logger.With("key", "value")
	assert.Equal(t, logger, contextLogger)
}

func TestLogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(LevelWarn, buf)

	// Debug and Info should be filtered out
	logger.Debug("debug message")
	logger.Info("info message")
	assert.Empty(t, buf.String())

	// Warn and Error should appear
	logger.Warn("warn message")
	output := buf.String()
	assert.Contains(t, output, "warn message")

	buf.Reset()
	logger.Error("error message")
	output = buf.String()
	assert.Contains(t, output, "error message")
}

func TestLogger_KeyValuePairs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(LevelInfo, buf)

	logger.Info("test with multiple fields",
		"string_field", "value",
		"int_field", 42,
		"bool_field", true,
		"float_field", 3.14,
	)

	output := buf.String()
	assert.Contains(t, output, "string_field")
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "int_field")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "bool_field")
	assert.Contains(t, output, "true")
}

func TestLogger_DefaultOutput(t *testing.T) {
	// Test with nil output (should default to stderr)
	logger := NewLogger(LevelInfo, nil)
	assert.NotNil(t, logger)

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Info("test message")
	})
}
