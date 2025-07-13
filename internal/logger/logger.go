package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger interface for dependency injection and testing
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	With(keysAndValues ...interface{}) Logger
}

// StructuredLogger implements Logger using slog
type StructuredLogger struct {
	logger *slog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(level slog.Level, output io.Writer) Logger {
	if output == nil {
		output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(output, opts)
	logger := slog.New(handler)

	return &StructuredLogger{
		logger: logger,
	}
}

// NewJSONLogger creates a new JSON structured logger
func NewJSONLogger(level slog.Level, output io.Writer) Logger {
	if output == nil {
		output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	return &StructuredLogger{
		logger: logger,
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, keysAndValues...)
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}

// With creates a new logger with additional context
func (l *StructuredLogger) With(keysAndValues ...interface{}) Logger {
	return &StructuredLogger{
		logger: l.logger.With(keysAndValues...),
	}
}

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct{}

func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *NoOpLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *NoOpLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *NoOpLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *NoOpLogger) With(keysAndValues ...interface{}) Logger       { return l }
