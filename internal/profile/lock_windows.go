//go:build windows

package profile

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
	procOpenProcess  = modkernel32.NewProc("OpenProcess")
)

const (
	lockfileExclusiveLock   = 0x00000002
	lockfileFailImmediately = 0x00000001
	processQueryLimitedInfo = 0x1000

	// The OS lock covers one byte far past EOF (offset 0x7FFFFFFF_FFFFFFFF)
	// instead of byte 0. Win32 byte-range locks are enforced per handle, and
	// AcquireLock writes the JSON metadata at offset 0 through a second handle
	// (os.WriteFile). Locking byte 0 made that write fail with
	// ERROR_LOCK_VIOLATION and also stopped other processes from reading the
	// metadata. The metadata never reaches this offset, so only competing
	// LockFileEx calls conflict with it.
	lockRangeOffsetLow  = 0xFFFFFFFF
	lockRangeOffsetHigh = 0x7FFFFFFF
)

// lockOverlapped returns an OVERLAPPED that addresses the lock byte.
func lockOverlapped() *syscall.Overlapped {
	return &syscall.Overlapped{Offset: lockRangeOffsetLow, OffsetHigh: lockRangeOffsetHigh}
}

// isProcessAlive checks if a process with the given PID is still running on Windows.
func isProcessAlive(pid int) bool {
	handle, _, _ := procOpenProcess.Call(
		uintptr(processQueryLimitedInfo),
		0,
		uintptr(pid),
	)
	if handle == 0 {
		return false
	}
	syscall.CloseHandle(syscall.Handle(handle))
	return true
}

// acquireFileLock acquires an exclusive lock on the given path using LockFileEx.
func acquireFileLock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %w", err)
	}

	handle := syscall.Handle(f.Fd())
	ol := lockOverlapped()

	r1, _, err := procLockFileEx.Call(
		uintptr(handle),
		uintptr(lockfileExclusiveLock|lockfileFailImmediately),
		0,
		1, 0,
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		f.Close()
		return nil, fmt.Errorf("LockFileEx failed: %w", err)
	}

	return func() {
		ol2 := lockOverlapped()
		procUnlockFileEx.Call(
			uintptr(handle),
			0,
			1, 0,
			uintptr(unsafe.Pointer(ol2)),
		)
		f.Close()
	}, nil
}

// holdGuardDuringMove: Windows cannot remove or rename a directory while a
// handle to the guard file inside it is open, so the guard is dropped right
// before the move.
const holdGuardDuringMove = false
