package logging

import (
	"bytes"
	"log/slog"
	"strings"
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

func TestEvalLoggerStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	l := EvalLogger("cosmos-crud", "baseline-opus", "generation", 3)
	l.Info("phase started")

	out := buf.String()
	for _, want := range []string{"prompt=cosmos-crud", "config=baseline-opus", "phase=generation", "worker=3"} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %q\ngot: %s", want, out)
		}
	}
}

func TestWithPhaseReplacesPhase(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	l := EvalLogger("p1", "c1", "generation", 0)
	l2 := WithPhase(l, "review")
	l2.Info("reviewing")

	out := buf.String()
	if !strings.Contains(out, "phase=review") {
		t.Errorf("expected phase=review in output, got: %s", out)
	}
}

func TestWithTurnAddsTurnField(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	l := EvalLogger("p1", "c1", "generation", 0)
	l2 := WithTurn(l, 5)
	l2.Info("turn started")

	out := buf.String()
	if !strings.Contains(out, "turn=5") {
		t.Errorf("expected turn=5 in output, got: %s", out)
	}
}

func TestSetupLogFile(t *testing.T) {
	tmp := t.TempDir()
	logFile := tmp + "/test.log"

	closer, err := Setup(Options{Level: "info", FilePath: logFile})
	if err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	slog.Info("file log test message")
	closer()

	// Restore stderr logger
	slog.SetDefault(slog.New(slog.NewTextHandler(nil, nil)))
}
