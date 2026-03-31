//go:build windows

package eval

import (
	"log/slog"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ronniegeraghty/hyoka/internal/pidfile"
)

// descendantPIDs returns the set of all descendant PIDs of the given root PID
// by enumerating processes via the Windows toolhelp API and walking the
// parent PID chain.
func descendantPIDs(root int) map[int]bool {
	descendants := make(map[int]bool)
	descendants[root] = true

	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return descendants
	}
	defer syscall.CloseHandle(snap)

	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := syscall.Process32First(snap, &pe); err != nil {
		return descendants
	}

	type procInfo struct {
		pid  int
		ppid int
	}
	var procs []procInfo
	for {
		procs = append(procs, procInfo{pid: int(pe.ProcessID), ppid: int(pe.ParentProcessID)})
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}

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

// FindCopilotProcesses enumerates running processes via the toolhelp API
// and returns PIDs of descendant processes whose executable name contains
// "copilot".
func FindCopilotProcesses() ([]int, error) {
	myDescendants := descendantPIDs(os.Getpid())

	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(snap)

	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := syscall.Process32First(snap, &pe); err != nil {
		return nil, err
	}

	myPID := os.Getpid()
	var pids []int
	for {
		pid := int(pe.ProcessID)
		if pid != myPID && myDescendants[pid] {
			name := strings.ToLower(syscall.UTF16ToString(pe.ExeFile[:]))
			if strings.Contains(name, "copilot") {
				pids = append(pids, pid)
			}
		}
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return pids, nil
}

// FindHyokaProcesses reads PID files written by the eval engine to find
// tracked Copilot processes. On Windows there is no /proc equivalent, so
// we rely on the PID file mechanism from the pidfile package.
func FindHyokaProcesses() ([]HyokaProcessInfo, error) {
	entries, err := pidfile.ReadAlive()
	if err != nil {
		return nil, err
	}
	procs := make([]HyokaProcessInfo, len(entries))
	for i, e := range entries {
		procs[i] = HyokaProcessInfo{
			PID:      e.PID,
			PromptID: e.PromptID,
			Config:   e.Config,
		}
	}
	return procs, nil
}

// TerminateOrphans finds and kills orphaned copilot processes.
// On Windows, processes are killed immediately (no SIGTERM equivalent).
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
		if err := proc.Kill(); err != nil {
			slog.Debug("Kill failed (process may have exited)", "pid", pid, "error", err)
			continue
		}
		proc.Release()
		terminated++
	}

	slog.Info("Post-run cleanup", "orphans_found", len(orphans), "orphans_terminated", terminated)
	return len(orphans)
}

// TerminateAll kills every tracked process immediately and waits up to
// timeout for them to exit. On Windows there is no SIGTERM, so we use Kill
// (TerminateProcess) directly.
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

	var errs []error
	for pid, proc := range snapshot {
		slog.Info("Terminating tracked process", "pid", pid)
		if err := proc.Kill(); err != nil {
			errs = append(errs, err)
		} else {
			_ = proc.Release()
		}
	}
	return errs
}
