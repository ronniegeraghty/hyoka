package eval

import (
	"log/slog"
	"os"
	"sync"
	"syscall"
	"time"
)

// DefaultTracker is the package-level ProcessTracker used to register all
// spawned Copilot processes and ensure cleanup on shutdown.
var DefaultTracker = &ProcessTracker{}

// ProcessTracker keeps a registry of spawned process PIDs and provides
// bulk termination with graceful-then-forced shutdown semantics.
type ProcessTracker struct {
	mu    sync.Mutex
	procs map[int]*os.Process
}

// Register adds a process to the tracker by PID.
func (pt *ProcessTracker) Register(pid int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.procs == nil {
		pt.procs = make(map[int]*os.Process)
	}
	if proc, err := os.FindProcess(pid); err == nil {
		pt.procs[pid] = proc
	}
}

// Deregister removes a process from the tracker.
func (pt *ProcessTracker) Deregister(pid int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.procs, pid)
}

// TerminateAll sends SIGTERM to every tracked process, waits up to timeout
// for them to exit, then sends SIGKILL to any stragglers. Returns any
// errors encountered during the kill phase.
func (pt *ProcessTracker) TerminateAll(timeout time.Duration) []error {
	pt.mu.Lock()
	// Snapshot current processes and clear the map.
	snapshot := make(map[int]*os.Process, len(pt.procs))
	for pid, proc := range pt.procs {
		snapshot[pid] = proc
	}
	pt.procs = nil
	pt.mu.Unlock()

	if len(snapshot) == 0 {
		return nil
	}

	// Phase 1: send SIGTERM to all tracked processes.
	for pid, proc := range snapshot {
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			// Process may already be gone — remove from snapshot.
			delete(snapshot, pid)
		}
	}

	if len(snapshot) == 0 {
		return nil
	}

	// Phase 2: wait up to timeout for processes to exit.
	deadline := time.After(timeout)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

waitLoop:
	for {
		select {
		case <-deadline:
			break waitLoop
		case <-tick.C:
			for pid, proc := range snapshot {
				// Signal(0) checks whether the process is still alive.
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					delete(snapshot, pid)
				}
			}
			if len(snapshot) == 0 {
				return nil
			}
		}
	}

	// Phase 3: SIGKILL any stragglers.
	var errs []error
	for pid, proc := range snapshot {
		slog.Warn("Process did not exit after SIGTERM, sending SIGKILL", "pid", pid)
		if err := proc.Kill(); err != nil {
			errs = append(errs, err)
		} else {
			// Reap zombie to avoid leaving defunct processes.
			proc.Release()
		}
		_ = pid // used in log above
	}
	return errs
}
