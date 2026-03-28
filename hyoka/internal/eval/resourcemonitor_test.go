package eval

import (
	"testing"
	"time"
)

func TestResourceMonitorStartStop(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 50*time.Millisecond)

	rm.Start()
	time.Sleep(100 * time.Millisecond)
	rm.Stop()

	stats := rm.RunStats()
	if stats == nil {
		t.Fatal("expected non-nil RunStats")
	}
	if stats.SessionCount != 0 {
		t.Errorf("expected 0 sessions, got %d", stats.SessionCount)
	}
}

func TestResourceMonitorRegisterEval(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 100*time.Millisecond)

	rm.RegisterEval("prompt-1/config-a")
	rm.RegisterEval("prompt-2/config-b")
	rm.RegisterEval("prompt-1/config-a") // duplicate, should not increment

	stats := rm.RunStats()
	if stats.SessionCount != 2 {
		t.Errorf("expected 2 sessions, got %d", stats.SessionCount)
	}
}

func TestResourceMonitorEvalStats(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 100*time.Millisecond)

	rm.RegisterEval("test-eval")

	// No tracked PIDs, so stats should have zero peaks
	es := rm.EvalStats("test-eval")
	if es == nil {
		t.Fatal("expected non-nil EvalStats for registered eval")
	}
	if es.PeakCPUPercent != 0 {
		t.Errorf("expected 0 peak CPU, got %f", es.PeakCPUPercent)
	}
	if es.PeakMemoryMB != 0 {
		t.Errorf("expected 0 peak memory, got %f", es.PeakMemoryMB)
	}
}

func TestResourceMonitorEvalStatsUnregistered(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 100*time.Millisecond)

	es := rm.EvalStats("nonexistent")
	if es != nil {
		t.Error("expected nil EvalStats for unregistered eval")
	}
}

func TestResourceMonitorSummaryLine(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 100*time.Millisecond)

	rm.RegisterEval("eval-1")
	rm.RegisterEval("eval-2")

	line := rm.SummaryLine()
	if line == "" {
		t.Error("expected non-empty summary line")
	}
	// Should contain session count
	if !contains(line, "2 sessions") {
		t.Errorf("expected summary to mention '2 sessions', got %q", line)
	}
}

func TestResourceMonitorSampleNoTrackedPIDs(t *testing.T) {
	tracker := &ProcessTracker{}
	rm := NewResourceMonitor(tracker, 50*time.Millisecond)

	rm.RegisterEval("empty-eval")
	rm.Start()
	time.Sleep(120 * time.Millisecond)
	rm.Stop()

	es := rm.EvalStats("empty-eval")
	if es == nil {
		t.Fatal("expected non-nil stats")
	}
	// With no tracked PIDs, sample should be a no-op — zero peaks
	if es.PeakCPUPercent != 0 || es.PeakMemoryMB != 0 {
		t.Errorf("expected zero peaks with no tracked PIDs, got cpu=%f mem=%f",
			es.PeakCPUPercent, es.PeakMemoryMB)
	}
}

func TestResourceMonitorDefaultInterval(t *testing.T) {
	rm := NewResourceMonitor(&ProcessTracker{}, 0)
	if rm.interval != 5*time.Second {
		t.Errorf("expected default interval 5s, got %v", rm.interval)
	}
}

func TestReadProcMemMBSelf(t *testing.T) {
	// Reading our own process's memory should return a positive value on Linux.
	mem := readSelfMemMB()
	// On non-Linux this will be 0 — that's fine, just ensure no panic.
	if mem < 0 {
		t.Errorf("expected non-negative memory, got %f", mem)
	}
}

func TestDiscoverChildPIDs(t *testing.T) {
	// Should not panic. On Linux will scan /proc, on other OS returns nil.
	children := discoverChildPIDs(1)
	_ = children // just ensure no crash
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
