package logging

import (
	"log/slog"
	"testing"
)

func TestResolveLevel(t *testing.T) {
	tests := []struct {
		name  string
		opts  Options
		want  slog.Level
	}{
		{"default", Options{}, slog.LevelWarn},
		{"debug flag", Options{Debug: true}, slog.LevelDebug},
		{"level debug", Options{Level: "debug"}, slog.LevelDebug},
		{"level info", Options{Level: "info"}, slog.LevelInfo},
		{"level warn", Options{Level: "warn"}, slog.LevelWarn},
		{"level error", Options{Level: "error"}, slog.LevelError},
		{"level overrides debug flag", Options{Level: "info", Debug: true}, slog.LevelInfo},
		{"unknown level defaults to warn", Options{Level: "unknown"}, slog.LevelWarn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveLevel(tt.opts)
			if got != tt.want {
				t.Errorf("resolveLevel(%+v) = %v, want %v", tt.opts, got, tt.want)
			}
		})
	}
}

func TestSetup_Stderr(t *testing.T) {
	closer, err := Setup(Options{Level: "warn"})
	if err != nil {
		t.Fatalf("Setup() error: %v", err)
	}
	defer closer()

	// Verify that the default logger is set (no panic on use)
	slog.Info("test message — should be suppressed at warn level")
}

func TestEvalLogger(t *testing.T) {
	l := EvalLogger("my-prompt", "my-config", "generation", 1)
	if l == nil {
		t.Fatal("EvalLogger returned nil")
	}
	// Smoke test — should not panic
	l.Info("eval started")
}
