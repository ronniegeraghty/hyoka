package eval

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ResourceStats holds per-eval peak resource utilization.
type ResourceStats struct {
	PeakCPUPercent float64 `json:"peak_cpu_percent"`
	PeakMemoryMB   float64 `json:"peak_memory_mb"`
	SampleCount    int     `json:"sample_count"`
}

// RunResourceStats holds aggregate resource utilization across all evals.
type RunResourceStats struct {
	PeakCPUPercent float64 `json:"peak_cpu_percent"`
	PeakMemoryMB   float64 `json:"peak_memory_mb"`
	SessionCount   int     `json:"session_count"`
}

// ResourceMonitor periodically samples CPU and memory of tracked PIDs.
type ResourceMonitor struct {
	mu       sync.Mutex
	tracker  *ProcessTracker
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// Per-eval stats keyed by eval ID (promptID/configName).
	evalStats map[string]*ResourceStats

	// Run-wide peaks.
	peakCPU    float64
	peakMemMB  float64
	sessions   int

	// Warning threshold: single process exceeding this RAM triggers a log warning.
	memWarnMB float64
}

// NewResourceMonitor creates a monitor that samples the given ProcessTracker.
func NewResourceMonitor(tracker *ProcessTracker, interval time.Duration) *ResourceMonitor {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &ResourceMonitor{
		tracker:   tracker,
		interval:  interval,
		stopCh:    make(chan struct{}),
		evalStats: make(map[string]*ResourceStats),
		memWarnMB: 2048, // 2 GB per-process warning threshold
	}
}

// Start begins periodic sampling in a background goroutine.
func (rm *ResourceMonitor) Start() {
	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()
		ticker := time.NewTicker(rm.interval)
		defer ticker.Stop()
		for {
			select {
			case <-rm.stopCh:
				return
			case <-ticker.C:
				rm.sample()
			}
		}
	}()
	slog.Info("Resource monitor started", "interval", rm.interval.String())
}

// Stop halts sampling and waits for the goroutine to exit.
func (rm *ResourceMonitor) Stop() {
	close(rm.stopCh)
	rm.wg.Wait()
	slog.Info("Resource monitor stopped")
}

// RegisterEval notes that an eval is active. Call when an eval starts.
func (rm *ResourceMonitor) RegisterEval(evalID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, ok := rm.evalStats[evalID]; !ok {
		rm.evalStats[evalID] = &ResourceStats{}
		rm.sessions++
	}
}

// EvalStats returns the recorded stats for a specific eval.
func (rm *ResourceMonitor) EvalStats(evalID string) *ResourceStats {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if s, ok := rm.evalStats[evalID]; ok {
		cp := *s
		return &cp
	}
	return nil
}

// RunStats returns aggregate run-level resource stats.
func (rm *ResourceMonitor) RunStats() *RunResourceStats {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return &RunResourceStats{
		PeakCPUPercent: rm.peakCPU,
		PeakMemoryMB:   rm.peakMemMB,
		SessionCount:   rm.sessions,
	}
}

// SummaryLine returns a human-readable one-liner for post-run output.
func (rm *ResourceMonitor) SummaryLine() string {
	s := rm.RunStats()
	return fmt.Sprintf("Peak: %d sessions, %.1fGB RAM, %.0f%% CPU",
		s.SessionCount, s.PeakMemoryMB/1024.0, s.PeakCPUPercent)
}

// sample reads /proc/<pid>/stat and /proc/<pid>/statm for each tracked PID.
func (rm *ResourceMonitor) sample() {
	rm.tracker.mu.Lock()
	pids := make([]int, 0, len(rm.tracker.procs))
	for pid := range rm.tracker.procs {
		pids = append(pids, pid)
	}
	rm.tracker.mu.Unlock()

	if len(pids) == 0 {
		return
	}

	var totalCPU float64
	var totalMemMB float64

	for _, pid := range pids {
		cpu := readProcCPU(pid)
		memMB := readProcMemMB(pid)

		totalCPU += cpu
		totalMemMB += memMB

		if memMB > rm.memWarnMB {
			slog.Warn("Process exceeds memory threshold",
				"pid", pid,
				"memory_mb", fmt.Sprintf("%.0f", memMB),
				"threshold_mb", fmt.Sprintf("%.0f", rm.memWarnMB))
		}
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if totalCPU > rm.peakCPU {
		rm.peakCPU = totalCPU
	}
	if totalMemMB > rm.peakMemMB {
		rm.peakMemMB = totalMemMB
	}

	// Update all active eval stats (resource usage is attributed globally
	// since we can't map PIDs to specific evals without SDK support).
	for _, s := range rm.evalStats {
		s.SampleCount++
		if totalCPU > s.PeakCPUPercent {
			s.PeakCPUPercent = totalCPU
		}
		if totalMemMB > s.PeakMemoryMB {
			s.PeakMemoryMB = totalMemMB
		}
	}

	slog.Debug("Resource sample",
		"pids", len(pids),
		"total_cpu_pct", fmt.Sprintf("%.1f", totalCPU),
		"total_mem_mb", fmt.Sprintf("%.1f", totalMemMB))
}

// readProcCPU reads CPU usage from /proc/<pid>/stat.
// Returns aggregate CPU percentage (sum of utime+stime as a rough proxy).
// On non-Linux systems returns 0.
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
