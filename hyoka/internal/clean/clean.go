// Package clean provides utilities for cleaning up stale Copilot CLI
// session state accumulated by hyoka eval runs.
package clean

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ronniegeraghty/hyoka/internal/pidfile"
)

// Options configures the clean operation.
type Options struct {
	// DryRun previews what would be cleaned without deleting.
	DryRun bool
	// KeepLogs is the number of recent log files to keep (default 50).
	KeepLogs int
	// All cleans all session-state, not just Hyoka-created sessions.
	All bool
	// KillOrphans terminates running processes tagged with HYOKA_SESSION.
	// When false, processes are listed by ScanHyokaProcesses but not killed.
	KillOrphans bool
	// Out is the writer for user-facing output.
	Out io.Writer
}

// Result holds the outcome of a clean operation.
type Result struct {
	SessionsFound   int
	SessionsRemoved int
	LogsRemoved     int
	BytesFreed      int64
	// Process cleanup stats.
	ProcessesFound  int
	ProcessesKilled int
}

// copilotStateDirFn returns the Copilot CLI state directory.
// It is a package-level variable so tests can override it.
var copilotStateDirFn = copilotStateDir

// copilotStateDir returns the Copilot CLI state directory.
func copilotStateDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "copilot-cli")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".copilot")
}

// Run performs the clean operation.
func Run(opts Options) (*Result, error) {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.KeepLogs <= 0 {
		opts.KeepLogs = 50
	}

	stateDir := copilotStateDirFn()
	if stateDir == "" {
		return nil, fmt.Errorf("cannot determine Copilot state directory")
	}

	result := &Result{}

	// Phase 1: Kill orphaned hyoka-tagged processes (#70)
	if opts.KillOrphans {
		if err := cleanProcesses(opts, result); err != nil {
			fmt.Fprintf(opts.Out, "Warning: process cleanup: %v\n", err)
		}
	}

	// Phase 2: Clean session-state directories
	sessDir := filepath.Join(stateDir, "session-state")
	if err := cleanSessions(sessDir, opts, result); err != nil {
		fmt.Fprintf(opts.Out, "Warning: session cleanup: %v\n", err)
	}

	// Phase 3: Clean old log files (only with --all since we can't
	// distinguish hyoka-spawned logs from Copilot CLI logs).
	if opts.All {
		logsDir := filepath.Join(stateDir, "logs")
		if err := cleanLogs(logsDir, opts, result); err != nil {
			fmt.Fprintf(opts.Out, "Warning: log cleanup: %v\n", err)
		}
	}

	return result, nil
}

// findHyokaProcessesFn reads PID files to discover tracked Copilot processes.
// Package-level variable so tests can override it.
var findHyokaProcessesFn = findHyokaProcesses

// HyokaProcessInfo describes a running process tagged with HYOKA_SESSION.
type HyokaProcessInfo struct {
	PID      int
	PromptID string
	Config   string
}

// findHyokaProcesses reads PID files written by the eval engine and returns
// entries whose processes are still alive. This is cross-platform — it does
// not depend on /proc or any OS-specific process introspection.
func findHyokaProcesses() ([]HyokaProcessInfo, error) {
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

// cleanProcesses finds and terminates orphaned hyoka-tagged copilot processes.
func cleanProcesses(opts Options, result *Result) error {
	procs, err := ScanHyokaProcesses(opts.Out)
	if err != nil {
		return err
	}

	result.ProcessesFound = len(procs)
	if len(procs) == 0 {
		return nil
	}

	if opts.DryRun {
		for _, p := range procs {
			fmt.Fprintf(opts.Out, "  [dry-run] would kill %s\n", formatProcessInfo(p))
		}
		return nil
	}

	killed := KillHyokaProcesses(procs, opts.Out)
	result.ProcessesKilled = killed
	return nil
}

// ScanHyokaProcesses finds all running processes tagged with HYOKA_SESSION.
// It prints a summary to out and returns the list of found processes.
func ScanHyokaProcesses(out io.Writer) ([]HyokaProcessInfo, error) {
	procs, err := findHyokaProcessesFn()
	if err != nil {
		return nil, err
	}
	if len(procs) == 0 {
		fmt.Fprintln(out, "No orphaned hyoka processes found.")
		return nil, nil
	}
	fmt.Fprintf(out, "Found %d hyoka-tagged process(es):\n", len(procs))
	for _, p := range procs {
		fmt.Fprintf(out, "  • %s\n", formatProcessInfo(p))
	}
	return procs, nil
}

// KillHyokaProcesses terminates each process in procs.
// On Unix it sends SIGTERM then SIGKILL after 5 s; on Windows it calls
// TerminateProcess immediately. Returns the number of processes successfully
// signaled. The implementation is in kill_unix.go / kill_windows.go.

// formatProcessInfo returns a human-readable description of a process.
func formatProcessInfo(p HyokaProcessInfo) string {
	desc := fmt.Sprintf("PID %d", p.PID)
	if p.PromptID != "" {
		desc += fmt.Sprintf("  prompt=%s", p.PromptID)
	}
	if p.Config != "" {
		desc += fmt.Sprintf("  config=%s", p.Config)
	}
	return desc
}

// cleanSessions removes stale session-state directories.
func cleanSessions(sessDir string, opts Options, result *Result) error {
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionPath := filepath.Join(sessDir, entry.Name())

		// If not --all, only clean sessions that look like hyoka-created ones.
		if !opts.All {
			if !isHyokaSession(sessionPath) {
				continue
			}
		}

		result.SessionsFound++
		size := dirSize(sessionPath)

		if opts.DryRun {
			fmt.Fprintf(opts.Out, "  [dry-run] would remove %s (%s)\n",
				entry.Name(), humanBytes(size))
		} else {
			if err := os.RemoveAll(sessionPath); err != nil {
				fmt.Fprintf(opts.Out, "  warning: failed to remove %s: %v\n", entry.Name(), err)
				continue
			}
			result.SessionsRemoved++
			result.BytesFreed += size
		}
	}
	return nil
}

// cleanLogs removes old log files, keeping the most recent N.
func cleanLogs(logsDir string, opts Options, result *Result) error {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Collect log files (not directories)
	type logEntry struct {
		name    string
		modTime int64
		size    int64
	}
	var logs []logEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		logs = append(logs, logEntry{
			name:    e.Name(),
			modTime: info.ModTime().UnixNano(),
			size:    info.Size(),
		})
	}

	// Sort newest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].modTime > logs[j].modTime
	})

	// Remove everything after the keep threshold
	for i := opts.KeepLogs; i < len(logs); i++ {
		logPath := filepath.Join(logsDir, logs[i].name)
		if opts.DryRun {
			fmt.Fprintf(opts.Out, "  [dry-run] would remove log %s (%s)\n",
				logs[i].name, humanBytes(logs[i].size))
		} else {
			if err := os.Remove(logPath); err != nil {
				continue
			}
			result.LogsRemoved++
			result.BytesFreed += logs[i].size
		}
	}
	return nil
}

// isHyokaSession checks if a session directory was created by hyoka.
// It looks for workspace.yaml files referencing hyoka report directories
// or temp directories with hyoka prefixes.
func isHyokaSession(sessionPath string) bool {
	// Check workspace.yaml for hyoka working directories
	wsYaml := filepath.Join(sessionPath, "workspace.yaml")
	data, err := os.ReadFile(wsYaml)
	if err == nil {
		content := string(data)
		if strings.Contains(content, "hyoka") ||
			strings.Contains(content, "reports/") ||
			strings.Contains(content, "hyoka-gen-") ||
			strings.Contains(content, "hyoka-config-") {
			return true
		}
	}

	// Check workspace.json as fallback
	wsJSON := filepath.Join(sessionPath, "workspace.json")
	data, err = os.ReadFile(wsJSON)
	if err == nil {
		content := string(data)
		if strings.Contains(content, "hyoka") ||
			strings.Contains(content, "reports/") ||
			strings.Contains(content, "hyoka-gen-") ||
			strings.Contains(content, "hyoka-config-") {
			return true
		}
	}

	return false
}

// dirSize returns the total size of all files in a directory tree.
func dirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files but continue walk
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// humanBytes formats a byte count as a human-readable string.
func humanBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1fGB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1fKB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
