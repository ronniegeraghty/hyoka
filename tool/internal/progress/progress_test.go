package progress

import (
	"testing"
	"time"
)

func TestNewBar(t *testing.T) {
	// Debug mode — disabled
	b := New(10, 4, true)
	if !b.disabled {
		t.Error("expected bar to be disabled in debug mode")
	}
	b.Start("test")
	b.Complete("test", true, 5*time.Second, false)
	b.Done()
}

func TestBarProgress(t *testing.T) {
	// Non-debug mode
	b := New(3, 2, false)
	if b.disabled {
		t.Error("expected bar to be enabled in non-debug mode")
	}

	b.Start("eval-1")
	b.Complete("eval-1", true, 10*time.Second, false)

	b.Start("eval-2")
	b.Complete("eval-2", false, 20*time.Second, false)

	b.Start("eval-3")
	b.Complete("eval-3", false, 5*time.Second, true)

	b.Done()

	if b.completed != 3 {
		t.Errorf("expected 3 completed, got %d", b.completed)
	}
	if len(b.results) != 3 {
		t.Errorf("expected 3 results, got %d", len(b.results))
	}
}

func TestFmtDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1.5m"},
	}
	for _, tt := range tests {
		got := fmtDuration(tt.d)
		if got != tt.want {
			t.Errorf("fmtDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
