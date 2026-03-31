//go:build !windows

package pidfile

import "syscall"

// isProcessAlive checks whether a process with the given PID is still
// running by sending signal 0 (which errors if the process is gone).
func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
