package eval

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// descendantPIDs returns the set of all descendant PIDs of the given root PID
// by scanning /proc/*/stat for processes whose PPID is in the set. This builds
// the full process tree rooted at root.
func descendantPIDs(root int) map[int]bool {
	descendants := make(map[int]bool)
	descendants[root] = true

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return descendants
	}

	// Build parent→children map
	type procInfo struct {
		pid  int
		ppid int
	}
	var procs []procInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		stat, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue
		}
		closeParen := strings.LastIndex(string(stat), ")")
		if closeParen < 0 {
			continue
		}
		fields := strings.Fields(string(stat)[closeParen+1:])
		if len(fields) < 2 {
			continue
		}
		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		procs = append(procs, procInfo{pid: pid, ppid: ppid})
	}

	// BFS to find all descendants
	changed := true
	for changed {
		changed = false
		for _, p := range procs {
			if descendants[p.ppid] && !descendants[p.pid] {
				descendants[p.pid] = true
				changed = true
			}
		}
	}
	return descendants
}

// FindCopilotProcesses scans /proc for running processes with "copilot" or
// "github-copilot" in their command line that are descendants of the current
// process. This ensures hyoka only finds copilot processes it spawned, not
// unrelated Copilot CLI sessions (e.g., interactive sessions in other terminals).
func FindCopilotProcesses() ([]int, error) {
	myDescendants := descendantPIDs(os.Getpid())

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("reading /proc: %w", err)
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // not a PID directory
		}
		if pid == os.Getpid() {
			continue
		}
		if !myDescendants[pid] {
			continue // not a descendant — skip
		}
		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil {
			continue // process may have exited
		}
		// cmdline uses null bytes as separators
		cmd := strings.ToLower(strings.ReplaceAll(string(cmdline), "\x00", " "))
		if strings.Contains(cmd, "copilot") || strings.Contains(cmd, "github-copilot") {
			pids = append(pids, pid)
		}
	}
	return pids, nil
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

// TerminateOrphans finds and kills orphaned copilot processes.
// Returns the number of orphaned processes found.
func (pt *ProcessTracker) TerminateOrphans() int {
	orphans := pt.ScanOrphans()
	if len(orphans) == 0 {
		return 0
	}

	terminated := 0
	for _, pid := range orphans {
		slog.Warn("Terminating orphaned copilot process", "pid", pid)
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		// SIGTERM first
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			continue // already gone
		}
		terminated++

		// Wait up to 5s for graceful exit, then SIGKILL
		go func(p *os.Process, id int) {
			deadline := time.After(5 * time.Second)
			tick := time.NewTicker(200 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-deadline:
					slog.Warn("Orphaned process did not exit after SIGTERM, sending SIGKILL", "pid", id)
					p.Kill()
					p.Release()
					return
				case <-tick.C:
					if err := p.Signal(syscall.Signal(0)); err != nil {
						return // process exited
					}
				}
			}
		}(proc, pid)
	}

	slog.Info("Post-run cleanup", "orphans_found", len(orphans), "orphans_terminated", terminated)
	return len(orphans)
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
			_ = proc.Release()
		}
		_ = pid // used in log above
	}
	return errs
}
