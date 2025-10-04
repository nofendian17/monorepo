package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// LoggerInterface defines the interface for logging operations
type LoggerInterface interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      slog.Level
	Output     io.Writer
	Format     string // "json" or "text"
	AddSource  bool
	WithTime   bool
	TimeFormat string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Level:      slog.LevelInfo,
		Output:     os.Stdout,
		Format:     "json", // default format
		AddSource:  false,
		WithTime:   true,
		TimeFormat: time.RFC3339,
	}
}

// New creates a new logger instance with the given configuration
func New(config Config) LoggerInterface {
	var handler slog.Handler

	// Set up options
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	// Choose handler based on format
	switch config.Format {
	case "text":
		handler = slog.NewTextHandler(config.Output, opts)
	default: // Default to JSON
		handler = slog.NewJSONHandler(config.Output, opts)
	}

	// If we need custom time formatting, we need to wrap the handler
	if config.WithTime && config.TimeFormat != time.RFC3339 {
		handler = &customTimeHandler{
			handler:    handler,
			timeFormat: config.TimeFormat,
		}
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewWithOptions creates a new logger with options
func NewWithOptions(opts ...Option) LoggerInterface {
	config := DefaultConfig()

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	return New(config)
}

// customTimeHandler implements slog.Handler to customize time formatting
type customTimeHandler struct {
	handler    slog.Handler
	timeFormat string
}

func (h *customTimeHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create a new record with custom time format as an attribute
	// Since we can't modify the original time field, we add a formatted time attribute
	r.AddAttrs(slog.String("time_formatted", r.Time.Format(h.timeFormat)))
	return h.handler.Handle(ctx, r)
}

func (h *customTimeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *customTimeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customTimeHandler{
		handler:    h.handler.WithAttrs(attrs),
		timeFormat: h.timeFormat,
	}
}

func (h *customTimeHandler) WithGroup(name string) slog.Handler {
	return &customTimeHandler{
		handler:    h.handler.WithGroup(name),
		timeFormat: h.timeFormat,
	}
}

// NewWithFormat creates a new logger with specified output format
func NewWithFormat(output io.Writer, level slog.Level, format string) LoggerInterface {
	config := DefaultConfig()
	config.Output = output
	config.Format = format
	config.Level = level
	return New(config)
}

// NewJSON creates a new JSON logger
func NewJSON(output io.Writer, level slog.Level) LoggerInterface {
	return NewWithFormat(output, level, "json")
}

// NewText creates a new text logger
func NewText(output io.Writer, level slog.Level) LoggerInterface {
	return NewWithFormat(output, level, "text")
}

// NewDefault creates a new logger with default configuration
func NewDefault() LoggerInterface {
	return New(DefaultConfig())
}

// NewJSONDefault creates a new JSON logger with default settings
func NewJSONDefault() LoggerInterface {
	config := DefaultConfig()
	config.Format = "json"
	return New(config)
}

// NewTextDefault creates a new text logger with default settings
func NewTextDefault() LoggerInterface {
	config := DefaultConfig()
	config.Format = "text"
	return New(config)
}

// WithContext returns a Logger that includes the provided context attributes
func WithContext(ctx context.Context, logger LoggerInterface) LoggerInterface {
	// This would typically extract request IDs or other context values
	// For now, we'll just return the logger as is
	// In a real implementation, you might extract values from context
	return logger
}

// InfoContext logs at the info level with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Log(ctx, slog.LevelInfo, msg, args...)
}

// ErrorContext logs at the error level with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Log(ctx, slog.LevelError, msg, args...)
}

// WarnContext logs at the warn level with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Log(ctx, slog.LevelWarn, msg, args...)
}

// DebugContext logs at the debug level with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Log(ctx, slog.LevelDebug, msg, args...)
}

// NoOpLogger returns a logger that does nothing - useful for testing
func NoOpLogger() LoggerInterface {
	return &Logger{
		Logger: slog.New(noOpHandler{}),
	}
}

// noOpHandler is a no-op implementation of slog.Handler
type noOpHandler struct{}

func (h noOpHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h noOpHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (h noOpHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h noOpHandler) WithGroup(_ string) slog.Handler {
	return h
}
