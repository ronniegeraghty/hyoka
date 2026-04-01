//go:build !windows

package eval

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// findChildCopilotPIDs returns PIDs of copilot processes that are direct
// children of the current process. It scans /proc for processes whose PPID
// matches our PID and whose command line contains "copilot".
func findChildCopilotPIDs() []int {
	myPID := os.Getpid()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid == myPID {
			continue
		}

		// Read /proc/<pid>/stat to get PPID (4th field, after the comm in parens).
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
		if err != nil || ppid != myPID {
			continue
		}

		// Check cmdline for copilot.
		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil {
			continue
		}
		cmd := strings.ToLower(strings.ReplaceAll(string(cmdline), "\x00", " "))
		if strings.Contains(cmd, "copilot") {
			pids = append(pids, pid)
		}
	}
	return pids
}
