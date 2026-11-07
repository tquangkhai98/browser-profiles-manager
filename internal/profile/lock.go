package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	lockFileName  = ".bpm.lock"
	guardFileName = ".bpm.lock.guard"
)

// LockInfo contains metadata about who holds a lock on a profile.
type LockInfo struct {
	PID      int       `json:"pid"`
	Caller   string    `json:"caller"`
	LockedAt time.Time `json:"locked_at"`
	// StartTime identifies the PID's process instance where the OS exposes it
	// (Linux). Zero means unknown; liveness then relies on the PID alone.
	StartTime uint64 `json:"start_time,omitempty"`
}

// inUseError formats the error returned when a live process holds the lock.
func inUseError(info *LockInfo) error {
	return fmt.Errorf("profile is already in use by pid %d (%s) since %s", info.PID, info.Caller, info.LockedAt.Format(time.RFC3339))
}

// AcquireLock holds the acquisition guard until release.
func AcquireLock(profileDir, caller string) (func(), error) {
	unlock, err := acquireFileLock(filepath.Join(profileDir, guardFileName))
	if err != nil {
		if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); lockAlive(info) {
			return nil, inUseError(info)
		}
		return nil, fmt.Errorf("cannot acquire file lock: %w", err)
	}
	path := filepath.Join(profileDir, lockFileName)
	if existing, _ := readLockInfo(path); lockAlive(existing) {
		unlock()
		return nil, inUseError(existing)
	}
	if err := WriteLock(profileDir, os.Getpid(), caller); err != nil {
		unlock()
		return nil, err
	}
	return func() { os.Remove(path); unlock() }, nil
}

// WriteLock updates lock metadata while the acquisition guard is held.
func WriteLock(profileDir string, pid int, caller string) error {
	start, _ := processStartTime(pid)
	data, err := json.MarshalIndent(LockInfo{PID: pid, Caller: caller, LockedAt: time.Now().UTC(), StartTime: start}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profileDir, lockFileName), data, 0600)
}

// AcquireForHandoff acquires a lock and exposes a handoff that preserves metadata.
func AcquireForHandoff(profileDir, caller string) (release func(), handoff func(int) error, err error) {
	unlock, err := acquireFileLock(filepath.Join(profileDir, guardFileName))
	if err != nil {
		if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); lockAlive(info) {
			return nil, nil, inUseError(info)
		}
		return nil, nil, err
	}
	path := filepath.Join(profileDir, lockFileName)
	if info, _ := readLockInfo(path); lockAlive(info) {
		unlock()
		return nil, nil, inUseError(info)
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

// HoldGuard takes the acquisition guard without writing lock metadata. While
// it is held no bpm process can acquire, hand off, or release the profile
// lock. Every bpm command that owns a lock itself (use, export, creds sync,
// desktop, a launch before handoff) holds the guard for as long as it owns the
// lock, so a free guard plus a live lock means a detached browser owns it.
func HoldGuard(profileDir string) (func(), error) {
	return acquireFileLock(filepath.Join(profileDir, guardFileName))
}

// LiveLock returns the lock metadata when its holder is still alive, without
// touching the lock file.
func LiveLock(profileDir string) *LockInfo {
	info, _ := readLockInfo(filepath.Join(profileDir, lockFileName))
	if !lockAlive(info) {
		return nil
	}
	return info
}

// ProcessAlive reports whether a PID still exists.
func ProcessAlive(pid int) bool { return pid > 0 && isProcessAlive(pid) && !processZombie(pid) }

// ProcessStartTime returns an identifier for the current instance of pid, or
// zero when the platform does not expose one.
func ProcessStartTime(pid int) uint64 {
	start, _ := processStartTime(pid)
	return start
}

// ProcessMatches reports whether pid is alive and, when start is known, is
// still the same process instance that start was recorded for.
func ProcessMatches(pid int, start uint64) bool {
	if !ProcessAlive(pid) {
		return false
	}
	if start != 0 {
		if current, ok := processStartTime(pid); ok && current != start {
			return false
		}
	}
	return true
}

func lockAlive(info *LockInfo) bool {
	// ProcessMatches also rejects zombies and PIDs reused by another process.
	return info != nil && ProcessMatches(info.PID, info.StartTime)
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
	if !lockAlive(info) {
		// Stale lock — clean it up
		unlock, err := acquireFileLock(filepath.Join(profileDir, guardFileName))
		if err == nil {
			current, _ := readLockInfo(lockPath)
			if current != nil && current.PID == info.PID && !lockAlive(current) {
				os.Remove(lockPath)
			}
			unlock()
		}
		return false, nil
	}

	return true, info
}

// lockForMutation claims the profile for a destructive operation (delete,
// rename). Unlike IsLocked it holds the guard, so no launch can start between
// the check and the change. A live holder is returned as holder (and the guard
// is not held) unless force is set. The returned unlock never removes lock
// metadata.
func lockForMutation(profileDir string, force bool) (unlock func(), holder *LockInfo, err error) {
	unlock, err = acquireFileLock(filepath.Join(profileDir, guardFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || force {
			return func() {}, nil, nil
		}
		if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); lockAlive(info) {
			return nil, info, nil
		}
		return nil, nil, fmt.Errorf("profile is busy: %w", err)
	}
	if info, _ := readLockInfo(filepath.Join(profileDir, lockFileName)); lockAlive(info) && !force {
		unlock()
		return nil, info, nil
	}
	return unlock, nil, nil
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
