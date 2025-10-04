package logger

import (
	"io"
	"log/slog"
	"os"
)

// Option is a function that configures a logger
type Option func(*Config)

// WithLevel sets the logging level
func WithLevel(level slog.Level) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithOutput sets the output writer
func WithOutput(output io.Writer) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithFormat sets the log format ("json" or "text")
func WithFormat(format string) Option {
	return func(c *Config) {
		c.Format = format
	}
}

// WithSource enables/disables source code location in logs
func WithSource(enabled bool) Option {
	return func(c *Config) {
		c.AddSource = enabled
	}
}

// WithTime enables/disables time in logs
func WithTime(enabled bool) Option {
	return func(c *Config) {
		c.WithTime = enabled
	}
}

// WithTimeFormat sets the time format
func WithTimeFormat(format string) Option {
	return func(c *Config) {
		c.TimeFormat = format
	}
}

// WithJSONFormat sets the format to JSON
func WithJSONFormat() Option {
	return WithFormat("json")
}

// WithTextFormat sets the format to text
func WithTextFormat() Option {
	return WithFormat("text")
}

// WithStdout sets the output to stdout
func WithStdout() Option {
	return WithOutput(os.Stdout)
}

// WithStderr sets the output to stderr
func WithStderr() Option {
	return WithOutput(os.Stderr)
}
