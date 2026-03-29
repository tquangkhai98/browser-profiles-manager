//go:build !windows

package config

import (
	"fmt"
	"os"
	"syscall"
)

// lockFile acquires an exclusive file lock using syscall.Flock.
// Returns an unlock function to release the lock.
func lockFile(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file %s: %w", path, err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("cannot acquire flock on %s: %w", path, err)
	}

	unlock := func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		os.Remove(path)
	}

	return unlock, nil
}
