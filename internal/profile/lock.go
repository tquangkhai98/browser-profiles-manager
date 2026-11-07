package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const lockFileName = ".bpm.lock"

// LockInfo contains metadata about who holds a lock on a profile.
type LockInfo struct {
	PID      int       `json:"pid"`
	Caller   string    `json:"caller"`
	LockedAt time.Time `json:"locked_at"`
}

// AcquireLock holds the acquisition guard until release.
func AcquireLock(profileDir, caller string) (func(), error) {
	unlock, err := acquireFileLock(filepath.Join(profileDir, ".bpm.lock.guard"))
	if err != nil {
		if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); info != nil && isProcessAlive(info.PID) {
			return nil, fmt.Errorf("profile is already in use by pid %d (%s) since %s", info.PID, info.Caller, info.LockedAt.Format(time.RFC3339))
		}
		return nil, fmt.Errorf("cannot acquire file lock: %w", err)
	}
	path := filepath.Join(profileDir, lockFileName)
	if existing, _ := readLockInfo(path); existing != nil && isProcessAlive(existing.PID) {
		unlock()
		return nil, fmt.Errorf("profile is already in use by pid %d (%s) since %s", existing.PID, existing.Caller, existing.LockedAt.Format(time.RFC3339))
	}
	if err := WriteLock(profileDir, os.Getpid(), caller); err != nil {
		unlock()
		return nil, err
	}
	return func() { os.Remove(path); unlock() }, nil
}

// WriteLock updates lock metadata while the acquisition guard is held.
func WriteLock(profileDir string, pid int, caller string) error {
	data, err := json.MarshalIndent(LockInfo{PID: pid, Caller: caller, LockedAt: time.Now().UTC()}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profileDir, lockFileName), data, 0600)
}

// AcquireForHandoff acquires a lock and exposes a handoff that preserves metadata.
func AcquireForHandoff(profileDir, caller string) (release func(), handoff func(int) error, err error) {
	unlock, err := acquireFileLock(filepath.Join(profileDir, ".bpm.lock.guard"))
	if err != nil {
		if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); info != nil && isProcessAlive(info.PID) {
			return nil, nil, fmt.Errorf("profile is already in use by pid %d (%s) since %s", info.PID, info.Caller, info.LockedAt.Format(time.RFC3339))
		}
		return nil, nil, err
	}
	path := filepath.Join(profileDir, lockFileName)
	if info, _ := readLockInfo(path); info != nil && isProcessAlive(info.PID) {
		unlock()
		return nil, nil, fmt.Errorf("profile is already in use by pid %d (%s) since %s", info.PID, info.Caller, info.LockedAt.Format(time.RFC3339))
	}
	if err = WriteLock(profileDir, os.Getpid(), caller); err != nil {
		unlock()
		return nil, nil, err
	}
	release = func() { os.Remove(path); unlock() }
	handoff = func(pid int) error {
		err := WriteLock(profileDir, pid, caller)
		if err == nil {
			unlock()
		}
		return err
	}
	return release, handoff, nil
}

// ProcessAlive reports whether a PID still exists.
func ProcessAlive(pid int) bool { return pid > 0 && isProcessAlive(pid) }

// ReleaseLock removes the lock file from a profile directory.
func ReleaseLock(profileDir string) error {
	lockPath := filepath.Join(profileDir, lockFileName)
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove lock file: %w", err)
	}
	return nil
}

// IsLocked checks if a profile directory is currently locked.
// Returns the lock info if locked, nil otherwise.
// Automatically cleans up stale locks.
func IsLocked(profileDir string) (bool, *LockInfo) {
	lockPath := filepath.Join(profileDir, lockFileName)

	info, err := readLockInfo(lockPath)
	if err != nil || info == nil {
		return false, nil
	}

	// Check if the locking process is still alive
	if !isProcessAlive(info.PID) {
		// Stale lock — clean it up
		unlock, err := acquireFileLock(filepath.Join(profileDir, ".bpm.lock.guard"))
		if err == nil {
			current, _ := readLockInfo(lockPath)
			if current != nil && current.PID == info.PID && !isProcessAlive(current.PID) {
				os.Remove(lockPath)
			}
			unlock()
		}
		return false, nil
	}

	return true, info
}

func readLockInfo(lockPath string) (*LockInfo, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}
	var info LockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
