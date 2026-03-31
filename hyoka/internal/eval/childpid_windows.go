//go:build windows

package eval

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// findChildCopilotPIDs returns PIDs of copilot processes that are direct
// children of the current process. It uses the documented
// CreateToolhelp32Snapshot / Process32First / Process32Next API to enumerate
// processes and match by parent PID and executable name.
func findChildCopilotPIDs() []int {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer syscall.CloseHandle(snap)

	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	if err := syscall.Process32First(snap, &pe); err != nil {
		return nil
	}

	myPID := uint32(os.Getpid())
	var pids []int

	for {
		if pe.ParentProcessID == myPID && pe.ProcessID != 0 {
			name := syscall.UTF16ToString(pe.ExeFile[:])
			if strings.Contains(strings.ToLower(name), "copilot") {
				pids = append(pids, int(pe.ProcessID))
			}
		}
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return pids
}
