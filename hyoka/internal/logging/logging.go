// Package logging provides structured logging for hyoka using log/slog.
//
// It configures a global slog.Logger based on CLI flags (--log-level,
// --log-file, --debug) and provides helpers to create child loggers
// with structured context fields for eval sessions.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options controls how the global logger is configured.
type Options struct {
	// Level is the minimum log level: "debug", "info", "warn", "error".
	// Default: "warn".
	Level string
	// FilePath redirects log output to a file, keeping stderr clean.
	FilePath string
	// Debug is a deprecated shorthand for Level="debug".
	Debug bool
}

// Setup initialises the global slog.Logger and returns a closer function
// that must be called to flush/close any open log file. If no log file
// is configured the closer is a no-op.
func Setup(opts Options) (closer func(), err error) {
	level := resolveLevel(opts)

	var w io.Writer = os.Stderr
	closer = func() {}

	if opts.FilePath != "" {
		f, err := os.OpenFile(opts.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("opening log file %s: %w", opts.FilePath, err)
		}
		w = f
		closer = func() { f.Close() }
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
	return closer, nil
}

// EvalLogger returns a child logger pre-loaded with eval-session context
// fields. Every message emitted through this logger carries the structured
// fields required by issue #42.
func EvalLogger(prompt, config, phase string, worker int) *slog.Logger {
	return slog.With(
		"prompt", prompt,
		"config", config,
		"phase", phase,
		"worker", worker,
	)
}

// WithPhase returns a copy of the logger with an updated phase field.
func WithPhase(l *slog.Logger, phase string) *slog.Logger {
	return l.With("phase", phase)
}

// WithTurn returns a copy of the logger with a turn field.
func WithTurn(l *slog.Logger, turn int) *slog.Logger {
	return l.With("turn", turn)
}

// resolveLevel converts flag values into a slog.Level, honouring the
// --debug backward-compat alias.
func resolveLevel(opts Options) slog.Level {
	if opts.Debug && opts.Level == "" {
		return slog.LevelDebug
	}
	switch strings.ToLower(opts.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
