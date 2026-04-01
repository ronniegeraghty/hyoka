//go:build windows

package eval

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modKernel32              = syscall.NewLazyDLL("kernel32.dll")
	modPsapi                 = syscall.NewLazyDLL("psapi.dll")
	procGetProcessTimes      = modKernel32.NewProc("GetProcessTimes")
	procGetProcessMemoryInfo = modPsapi.NewProc("GetProcessMemoryInfo")
)

// processMemoryCounters corresponds to the Windows PROCESS_MEMORY_COUNTERS struct.
type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

const processQueryLimitedInfo = 0x1000

// readProcCPU reads CPU usage via GetProcessTimes.
// Returns cumulative user+kernel time in seconds (analogous to the Unix /proc version).
func readProcCPU(pid int) float64 {
	h, err := syscall.OpenProcess(processQueryLimitedInfo, false, uint32(pid))
	if err != nil {
		return 0
	}
	defer syscall.CloseHandle(h)

	var creation, exit, kernel, user syscall.Filetime
	r, _, _ := procGetProcessTimes.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&creation)),
		uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r == 0 {
		return 0
	}

	// FILETIME is in 100-nanosecond intervals; convert to seconds.
	kNs := int64(kernel.HighDateTime)<<32 | int64(kernel.LowDateTime)
	uNs := int64(user.HighDateTime)<<32 | int64(user.LowDateTime)
	return float64(kNs+uNs) / 1e7
}

// readProcMemMB reads working-set size via GetProcessMemoryInfo and returns MB.
func readProcMemMB(pid int) float64 {
	h, err := syscall.OpenProcess(processQueryLimitedInfo|0x0400, false, uint32(pid)) // 0x0400 = PROCESS_QUERY_INFORMATION
	if err != nil {
		return 0
	}
	defer syscall.CloseHandle(h)

	var pmc processMemoryCounters
	pmc.CB = uint32(unsafe.Sizeof(pmc))
	r, _, _ := procGetProcessMemoryInfo.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&pmc)),
		uintptr(pmc.CB),
	)
	if r == 0 {
		return 0
	}
	return float64(pmc.WorkingSetSize) / (1024 * 1024)
}

// readSelfMemMB reads the current process's own working-set size.
func readSelfMemMB() float64 {
	return readProcMemMB(os.Getpid())
}

// discoverChildPIDs finds all child PIDs of the given parent by walking the
// process tree via the Windows toolhelp API.
func discoverChildPIDs(parentPID int) []int {
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

	var children []int
	for {
		if int(pe.ParentProcessID) == parentPID {
			children = append(children, int(pe.ProcessID))
		}
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return children
}
