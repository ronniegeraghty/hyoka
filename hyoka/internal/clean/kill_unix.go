//go:build !windows

package clean

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"
	"time"
)

// KillHyokaProcesses sends SIGTERM to each process and SIGKILL after 5 s
// if it has not exited. Returns the number of processes successfully signaled.
func KillHyokaProcesses(procs []HyokaProcessInfo, out io.Writer) int {
	killed := 0
	for _, p := range procs {
		proc, findErr := os.FindProcess(p.PID)
		if findErr != nil {
			slog.Debug("process already gone", "pid", p.PID)
			continue
		}

		if sigErr := proc.Signal(syscall.SIGTERM); sigErr != nil {
			slog.Debug("SIGTERM failed (process may have exited)", "pid", p.PID, "error", sigErr)
			continue
		}

		killed++
		fmt.Fprintf(out, "  Terminated %s\n", formatProcessInfo(p))

		go func(pr *os.Process, id int) {
			deadline := time.After(5 * time.Second)
			tick := time.NewTicker(200 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-deadline:
					slog.Warn("Orphan did not exit after SIGTERM, sending SIGKILL", "pid", id)
					pr.Kill()
					pr.Release()
					return
				case <-tick.C:
					if err := pr.Signal(syscall.Signal(0)); err != nil {
						return
					}
				}
			}
		}(proc, p.PID)
	}
	return killed
}
