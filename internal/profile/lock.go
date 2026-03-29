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

// AcquireLock attempts to lock a profile directory.
// It writes a .bpm.lock file with metadata and acquires an OS-level file lock.
func AcquireLock(profileDir, caller string) (func(), error) {
	lockPath := filepath.Join(profileDir, lockFileName)

	// Check for stale lock first
	if existing, _ := readLockInfo(lockPath); existing != nil {
		if !isProcessAlive(existing.PID) {
			// Stale lock — remove it silently
			os.Remove(lockPath)
		} else {
			return nil, fmt.Errorf("profile is already locked by PID %d (%s) since %s",
				existing.PID, existing.Caller, existing.LockedAt.Format(time.RFC3339))
		}
	}

	// Acquire OS-level lock
	unlock, err := acquireFileLock(lockPath)
	if err != nil {
		return nil, fmt.Errorf("cannot acquire file lock: %w", err)
	}

	// Write lock metadata
	info := LockInfo{
		PID:      os.Getpid(),
		Caller:   caller,
		LockedAt: time.Now().UTC(),
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	if err := os.WriteFile(lockPath, data, 0600); err != nil {
		unlock()
		return nil, fmt.Errorf("cannot write lock file: %w", err)
	}

	release := func() {
		os.Remove(lockPath)
		unlock()
	}
	return release, nil
}

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
		os.Remove(lockPath)
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
