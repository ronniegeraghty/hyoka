//go:build !windows

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

// descendantPIDs returns the set of all descendant PIDs of the given root PID
// by scanning /proc/*/stat for processes whose PPID is in the set. This builds
// the full process tree rooted at root.
func descendantPIDs(root int) map[int]bool {
	descendants := make(map[int]bool)
	descendants[root] = true

	entries, err := os.ReadDir("/proc")
	if err != nil {
		procWarnOnce.Do(func() {
			slog.Warn("Cannot scan /proc, process tracking degraded")
		})
		return descendants
	}

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

	children := make(map[int][]int)
	for _, p := range procs {
		children[p.ppid] = append(children[p.ppid], p.pid)
	}
	queue := []int{root}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		descendants[pid] = true
		queue = append(queue, children[pid]...)
	}
	return descendants
}

// FindCopilotProcesses scans /proc for running processes with "copilot" or
// "github-copilot" in their command line that are descendants of the current
// process.
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
			continue
		}
		if pid == os.Getpid() {
			continue
		}
		if !myDescendants[pid] {
			continue
		}
		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil {
			continue
		}
		cmd := strings.ToLower(strings.ReplaceAll(string(cmdline), "\x00", " "))
		if strings.Contains(cmd, "copilot") || strings.Contains(cmd, "github-copilot") {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

// FindHyokaProcesses scans /proc for any running process whose environment
// contains HYOKA_SESSION=true, regardless of ancestry.
func FindHyokaProcesses() ([]HyokaProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("reading /proc: %w", err)
	}

	var procs []HyokaProcessInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if pid == os.Getpid() {
			continue
		}
		info, ok := readHyokaEnv(pid)
		if !ok {
			continue
		}
		info.PID = pid
		procs = append(procs, info)
	}
	return procs, nil
}

// readHyokaEnv reads /proc/<pid>/environ and checks for HYOKA_SESSION=true.
func readHyokaEnv(pid int) (HyokaProcessInfo, bool) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return HyokaProcessInfo{}, false
	}

	var info HyokaProcessInfo
	found := false
	for _, envVar := range strings.Split(string(data), "\x00") {
		switch {
		case envVar == EnvHyokaSession+"=true":
			found = true
		case strings.HasPrefix(envVar, EnvHyokaPromptID+"="):
			info.PromptID = envVar[len(EnvHyokaPromptID)+1:]
		case strings.HasPrefix(envVar, EnvHyokaConfig+"="):
			info.Config = envVar[len(EnvHyokaConfig)+1:]
		}
	}
	return info, found
}

// TerminateOrphans finds and kills orphaned copilot processes.
// Returns the number of orphaned processes found.
func (pt *ProcessTracker) TerminateOrphans() int {
	orphans := pt.ScanOrphans()
	if len(orphans) == 0 {
		return 0
	}

	terminated := 0
	var wg sync.WaitGroup
	for _, pid := range orphans {
		slog.Warn("Terminating orphaned copilot process", "pid", pid)
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			continue
		}
		terminated++

		wg.Add(1)
		go func(p *os.Process, id int) {
			defer wg.Done()
			deadline := time.After(5 * time.Second)
			tick := time.NewTicker(1 * time.Second)
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
						return
					}
				}
			}
		}(proc, pid)
	}

	wg.Wait()
	slog.Info("Post-run cleanup", "orphans_found", len(orphans), "orphans_terminated", terminated)
	return len(orphans)
}

// TerminateAll sends SIGTERM to every tracked process, waits up to timeout
// for them to exit, then sends SIGKILL to any stragglers.
func (pt *ProcessTracker) TerminateAll(timeout time.Duration) []error {
	pt.mu.Lock()
	snapshot := make(map[int]*os.Process, len(pt.procs))
	for pid, proc := range pt.procs {
		snapshot[pid] = proc
	}
	pt.procs = make(map[int]*os.Process)
	pt.pids = make(map[int]string)
	pt.mu.Unlock()

	if len(snapshot) == 0 {
		return nil
	}

	for pid, proc := range snapshot {
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			delete(snapshot, pid)
		}
	}

	if len(snapshot) == 0 {
		return nil
	}

	deadline := time.After(timeout)
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

waitLoop:
	for {
		select {
		case <-deadline:
			break waitLoop
		case <-tick.C:
			for pid, proc := range snapshot {
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					delete(snapshot, pid)
				}
			}
			if len(snapshot) == 0 {
				return nil
			}
		}
	}

	var errs []error
	for pid, proc := range snapshot {
		slog.Warn("Process did not exit after SIGTERM, sending SIGKILL", "pid", pid)
		if err := proc.Kill(); err != nil {
			errs = append(errs, err)
		} else {
			_ = proc.Release()
		}
		_ = pid
	}
	return errs
}
