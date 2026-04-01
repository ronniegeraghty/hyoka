//go:build windows

package clean

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// KillHyokaProcesses terminates each process via TerminateProcess (there is
// no SIGTERM equivalent on Windows). Returns the number of processes killed.
func KillHyokaProcesses(procs []HyokaProcessInfo, out io.Writer) int {
	killed := 0
	for _, p := range procs {
		proc, findErr := os.FindProcess(p.PID)
		if findErr != nil {
			slog.Debug("process already gone", "pid", p.PID)
			continue
		}

		if err := proc.Kill(); err != nil {
			slog.Debug("Kill failed (process may have exited)", "pid", p.PID, "error", err)
			continue
		}
		proc.Release()

		killed++
		fmt.Fprintf(out, "  Terminated %s\n", formatProcessInfo(p))
	}
	return killed
}
