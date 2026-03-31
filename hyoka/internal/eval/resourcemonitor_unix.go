//go:build !windows

package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// readProcCPU reads CPU usage from /proc/<pid>/stat.
// Returns aggregate CPU percentage (sum of utime+stime as a rough proxy).
func readProcCPU(pid int) float64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0
	}
	// /proc/<pid>/stat format: pid (comm) state ... field14=utime field15=stime
	// We need to find the closing ')' of comm first because comm can contain spaces.
	s := string(data)
	closeIdx := strings.LastIndex(s, ")")
	if closeIdx < 0 || closeIdx+2 >= len(s) {
		return 0
	}
	fields := strings.Fields(s[closeIdx+2:])
	// fields[0]=state, fields[11]=utime, fields[12]=stime (0-indexed after state)
	if len(fields) < 13 {
		return 0
	}
	utime, _ := strconv.ParseFloat(fields[11], 64)
	stime, _ := strconv.ParseFloat(fields[12], 64)
	// Convert clock ticks to a percentage approximation.
	// This is a cumulative value; for a snapshot we report the raw tick sum
	// which gives a rough magnitude indicator across samples.
	clkTck := 100.0 // sysconf(_SC_CLK_TCK) is typically 100 on Linux
	return (utime + stime) / clkTck
}

// readProcMemMB reads RSS from /proc/<pid>/statm and converts to MB.
func readProcMemMB(pid int) float64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	// fields[1] = RSS in pages
	rssPages, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0
	}
	pageSize := float64(os.Getpagesize())
	return (rssPages * pageSize) / (1024 * 1024)
}

// readSelfMemMB reads the current process's own RSS as a fallback when
// no tracked PIDs are available. Uses /proc/self/statm.
func readSelfMemMB() float64 {
	return readProcMemMB(os.Getpid())
}

// discoverChildPIDs finds all child PIDs of the current process by scanning
// /proc/*/status. This is used as a fallback when the ProcessTracker has no
// registered PIDs (since the Copilot SDK doesn't expose child PIDs).
func discoverChildPIDs(parentPID int) []int {
	entries, err := filepath.Glob("/proc/[0-9]*/status")
	if err != nil {
		return nil
	}
	ppidPrefix := fmt.Sprintf("PPid:\t%d", parentPID)
	var children []int
	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), ppidPrefix) {
			// Extract PID from path: /proc/<pid>/status
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				if pid, err := strconv.Atoi(parts[2]); err == nil {
					children = append(children, pid)
				}
			}
		}
	}
	return children
}
