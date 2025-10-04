package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := Config{
		Level:     slog.LevelInfo,
		Output:    &bytes.Buffer{},
		Format:    "json",
		AddSource: false,
		WithTime:  true,
	}

	logger := New(config)
	require.NotNil(t, logger, "New() should not return nil")
}

func TestNewWithOptions(t *testing.T) {
	logger := NewWithOptions(
		WithLevel(slog.LevelDebug),
		WithOutput(&bytes.Buffer{}),
		WithFormat("text"),
	)

	require.NotNil(t, logger, "NewWithOptions() should not return nil")
}

func TestNewWithFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithFormat(buf, slog.LevelInfo, "json")

	require.NotNil(t, logger, "NewWithFormat() should not return nil")
}

func TestNewJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelInfo)

	require.NotNil(t, logger, "NewJSON() should not return nil")
}

func TestNewText(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewText(buf, slog.LevelInfo)

	require.NotNil(t, logger, "NewText() should not return nil")
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	require.NotNil(t, logger, "NewDefault() should not return nil")
}

func TestNewJSONDefault(t *testing.T) {
	logger := NewJSONDefault()
	require.NotNil(t, logger, "NewJSONDefault() should not return nil")
}

func TestNewTextDefault(t *testing.T) {
	logger := NewTextDefault()
	require.NotNil(t, logger, "NewTextDefault() should not return nil")
}

func TestLogger_Log(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelInfo)

	ctx := context.Background()
	logger.Log(ctx, slog.LevelInfo, "test message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "test message", "Log output should contain 'test message'")
	assert.Contains(t, output, "key", "Log output should contain 'key'")
	assert.Contains(t, output, "value", "Log output should contain 'value'")
}

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelInfo)

	logger.Info("info message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "info message", "Log output should contain 'info message'")
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelError)

	logger.Error("error message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "error message", "Log output should contain 'error message'")
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelWarn)

	logger.Warn("warn message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "warn message", "Log output should contain 'warn message'")
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelDebug)

	logger.Debug("debug message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "debug message", "Log output should contain 'debug message'")
}

func TestLogger_ContextMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelDebug)
	ctx := context.Background()

	logger.InfoContext(ctx, "info context message")
	logger.ErrorContext(ctx, "error context message")
	logger.WarnContext(ctx, "warn context message")
	logger.DebugContext(ctx, "debug context message")

	output := buf.String()
	assert.Contains(t, output, "info context message", "Should contain info context message")
	assert.Contains(t, output, "error context message", "Should contain error context message")
	assert.Contains(t, output, "warn context message", "Should contain warn context message")
	assert.Contains(t, output, "debug context message", "Should contain debug context message")
}

func TestNoOpLogger(t *testing.T) {
	logger := NoOpLogger()
	require.NotNil(t, logger, "NoOpLogger() should not return nil")

	// No-op logger should not panic
	logger.Info("test")
	logger.Error("test")
	logger.Warn("test")
	logger.Debug("test")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, slog.LevelInfo, config.Level, "Default level should be Info")
	assert.Equal(t, "json", config.Format, "Default format should be 'json'")
	assert.True(t, config.WithTime, "WithTime should be true by default")
	assert.Equal(t, time.RFC3339, config.TimeFormat, "Default time format should be RFC3339")
}

func TestWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelInfo)
	ctx := context.Background()

	result := WithContext(ctx, logger)
	require.NotNil(t, result, "WithContext() should not return nil")

	// For now, WithContext just returns the logger as is
	assert.Equal(t, logger, result, "WithContext() should return the same logger instance")
}

func TestWithSource(t *testing.T) {
	config := &Config{}
	opt := WithSource(true)
	opt(config)
	assert.True(t, config.AddSource, "WithSource should set AddSource to true")
}

func TestWithTime(t *testing.T) {
	config := &Config{}
	opt := WithTime(false)
	opt(config)
	assert.False(t, config.WithTime, "WithTime should set WithTime to false")
}

func TestWithTimeFormat(t *testing.T) {
	config := &Config{}
	format := "2006-01-02"
	opt := WithTimeFormat(format)
	opt(config)
	assert.Equal(t, format, config.TimeFormat, "WithTimeFormat should set TimeFormat")
}

func TestWithJSONFormat(t *testing.T) {
	config := &Config{}
	opt := WithJSONFormat()
	opt(config)
	assert.Equal(t, "json", config.Format, "WithJSONFormat should set format to json")
}

func TestWithTextFormat(t *testing.T) {
	config := &Config{}
	opt := WithTextFormat()
	opt(config)
	assert.Equal(t, "text", config.Format, "WithTextFormat should set format to text")
}

func TestWithStdout(t *testing.T) {
	config := &Config{}
	opt := WithStdout()
	opt(config)
	assert.Equal(t, os.Stdout, config.Output, "WithStdout should set output to stdout")
}

func TestWithStderr(t *testing.T) {
	config := &Config{}
	opt := WithStderr()
	opt(config)
	assert.Equal(t, os.Stderr, config.Output, "WithStderr should set output to stderr")
}

func TestLogger_HandlerInterface(t *testing.T) {
	buf := &bytes.Buffer{}
	concreteLogger := &Logger{slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))}

	// Test Enabled method - this is covered by the embedded slog.Logger
	ctx := context.Background()
	enabled := concreteLogger.Logger.Enabled(ctx, slog.LevelInfo)
	assert.True(t, enabled, "Logger should be enabled for Info level")

	// Test that logging works (this indirectly tests Handle)
	concreteLogger.Info("test message")
	assert.True(t, buf.Len() > 0, "Logger should write to buffer")
}

func TestLogger_BasicFunctionality(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSON(buf, slog.LevelInfo)

	// Test that logging works
	logger.Info("test message")
	assert.True(t, buf.Len() > 0, "Logger should write to buffer")
}

func TestNewWithOptions_Source(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithOptions(WithSource(true), WithOutput(buf))
	require.NotNil(t, logger, "WithSource option should work")
}

func TestNewWithOptions_Time(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithOptions(WithTime(true), WithOutput(buf))
	require.NotNil(t, logger, "WithTime option should work")
}

func TestNewWithOptions_TimeFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithOptions(WithTimeFormat(time.RFC3339), WithOutput(buf))
	require.NotNil(t, logger, "WithTimeFormat option should work")
}

func TestNewWithOptions_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithOptions(WithJSONFormat(), WithOutput(buf))
	require.NotNil(t, logger, "WithJSONFormat option should work")
}

func TestNewWithOptions_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithOptions(WithTextFormat(), WithOutput(buf))
	require.NotNil(t, logger, "WithTextFormat option should work")
}

func TestNewWithOptions_Stdout(t *testing.T) {
	logger := NewWithOptions(WithStdout())
	require.NotNil(t, logger, "WithStdout option should work")
}

func TestNewWithOptions_Stderr(t *testing.T) {
	logger := NewWithOptions(WithStderr())
	require.NotNil(t, logger, "WithStderr option should work")
}
