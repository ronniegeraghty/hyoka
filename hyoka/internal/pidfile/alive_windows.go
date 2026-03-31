//go:build windows

package pidfile

import "syscall"

const processQueryLimitedInformation = 0x1000

// isProcessAlive checks whether a process with the given PID is still
// running by attempting to open a handle with minimal access rights.
func isProcessAlive(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(h)
	return true
}
