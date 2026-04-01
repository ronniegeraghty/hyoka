package eval

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// resMonWarnOnce gates the "no process data" warning to fire at most once.
var resMonWarnOnce sync.Once

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
	peakCPU   float64
	peakMemMB float64
	sessions  int

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

	if totalCPU == 0 && totalMemMB == 0 && len(pids) > 0 {
		resMonWarnOnce.Do(func() {
			slog.Warn("Resource monitor: no process data available")
		})
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
