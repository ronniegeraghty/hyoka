package eval

import (
	"os"
	"testing"
)

func TestProcessTrackerRegisterDeregister(t *testing.T) {
	pt := &ProcessTracker{}

	// Register our own PID (guaranteed to exist)
	pid := os.Getpid()
	pt.Register(pid, "test-process")

	tracked := pt.TrackedPIDs()
	if desc, ok := tracked[pid]; !ok {
		t.Fatalf("expected pid %d to be tracked", pid)
	} else if desc != "test-process" {
		t.Fatalf("expected description %q, got %q", "test-process", desc)
	}

	pt.Deregister(pid)
	tracked = pt.TrackedPIDs()
	if _, ok := tracked[pid]; ok {
		t.Fatalf("expected pid %d to be deregistered", pid)
	}
}

func TestProcessTrackerScanOrphansEmpty(t *testing.T) {
	pt := &ProcessTracker{}
	// With nothing tracked, ScanOrphans should not panic
	orphans := pt.ScanOrphans()
	// We can't assert exact count since system may have copilot processes,
	// but the call should not error.
	_ = orphans
}

func TestProcessTrackerTerminateOrphansNoOp(t *testing.T) {
	pt := &ProcessTracker{}
	// Only verify ScanOrphans works without panic. We do NOT call
	// TerminateOrphans because it sends SIGTERM to any copilot process
	// not tracked by this (empty) tracker — which would kill the
	// Copilot CLI process running interactive sessions.
	orphans := pt.ScanOrphans()
	_ = orphans
}

func TestFindCopilotProcesses(t *testing.T) {
	pids, err := FindCopilotProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// We can't assert specific PIDs, but verify it returns a valid slice
	_ = pids
}
