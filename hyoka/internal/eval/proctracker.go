package eval

import (
	"log/slog"
	"os"
	"sync"
)

// DefaultTracker is the package-level ProcessTracker used to register all
// spawned Copilot processes and ensure cleanup on shutdown.
var DefaultTracker = &ProcessTracker{}

// procWarnOnce gates the /proc unavailability warning to fire at most once.
var procWarnOnce sync.Once

// ProcessTracker keeps a registry of spawned process PIDs and provides
// bulk termination with graceful-then-forced shutdown semantics.
type ProcessTracker struct {
	mu    sync.Mutex
	procs map[int]*os.Process
	pids  map[int]string // pid -> description
}

// Register adds a process to the tracker by PID with a description.
func (pt *ProcessTracker) Register(pid int, description string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.procs == nil {
		pt.procs = make(map[int]*os.Process)
	}
	if pt.pids == nil {
		pt.pids = make(map[int]string)
	}
	if proc, err := os.FindProcess(pid); err == nil {
		pt.procs[pid] = proc
		pt.pids[pid] = description
		slog.Info("Registering copilot process", "pid", pid, "description", description)
	}
}

// Deregister removes a process from the tracker.
func (pt *ProcessTracker) Deregister(pid int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	slog.Info("Deregistering copilot process", "pid", pid)
	delete(pt.procs, pid)
	delete(pt.pids, pid)
}

// TrackedPIDs returns a snapshot of all currently tracked PIDs.
func (pt *ProcessTracker) TrackedPIDs() map[int]string {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	out := make(map[int]string, len(pt.pids))
	for pid, desc := range pt.pids {
		out[pid] = desc
	}
	return out
}

// HyokaProcessInfo describes a running process tagged with HYOKA_SESSION.
type HyokaProcessInfo struct {
	PID      int
	PromptID string
	Config   string
}

// ScanOrphans finds copilot processes not tracked by this ProcessTracker.
func (pt *ProcessTracker) ScanOrphans() []int {
	allCopilot, err := FindCopilotProcesses()
	if err != nil {
		slog.Warn("Failed to scan for copilot processes", "error", err)
		return nil
	}

	pt.mu.Lock()
	tracked := make(map[int]bool, len(pt.procs))
	for pid := range pt.procs {
		tracked[pid] = true
	}
	pt.mu.Unlock()

	var orphans []int
	for _, pid := range allCopilot {
		if !tracked[pid] {
			orphans = append(orphans, pid)
		}
	}
	return orphans
}
