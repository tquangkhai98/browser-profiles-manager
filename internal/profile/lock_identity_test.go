package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func writeLockInfo(t *testing.T, dir string, info LockInfo) {
	t.Helper()
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, lockFileName), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestProcessMatches(t *testing.T) {
	pid := os.Getpid()
	if !ProcessMatches(pid, 0) {
		t.Fatal("own process with unknown start time must match")
	}
	start := ProcessStartTime(pid)
	if runtime.GOOS != "linux" {
		if start != 0 {
			t.Fatalf("start time = %d, want 0 off Linux", start)
		}
		return
	}
	if start == 0 {
		t.Fatal("expected a start time on Linux")
	}
	if !ProcessMatches(pid, start) {
		t.Fatal("own process with its own start time must match")
	}
	if ProcessMatches(pid, start+1) {
		t.Fatal("a different start time means the PID was reused")
	}
}

func TestLockFromReusedPIDIsStale(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("process start time is only available on Linux")
	}
	dir := t.TempDir()
	// The PID is alive (it is this test), but the recorded instance is not.
	writeLockInfo(t, dir, LockInfo{PID: os.Getpid(), Caller: "launch", LockedAt: time.Now(), StartTime: ProcessStartTime(os.Getpid()) + 1})
	if locked, _ := IsLocked(dir); locked {
		t.Fatal("lock held by a reused PID must be stale")
	}
	release, err := AcquireLock(dir, "test")
	if err != nil {
		t.Fatalf("AcquireLock over a reused-PID lock: %v", err)
	}
	release()
}

func TestWriteLockRecordsStartTime(t *testing.T) {
	dir := t.TempDir()
	release, err := AcquireLock(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	info, err := readLockInfo(filepath.Join(dir, lockFileName))
	if err != nil {
		t.Fatal(err)
	}
	if info.StartTime != ProcessStartTime(os.Getpid()) {
		t.Fatalf("StartTime = %d, want %d", info.StartTime, ProcessStartTime(os.Getpid()))
	}
}

func TestHoldGuardBlocksAcquire(t *testing.T) {
	dir := t.TempDir()
	unlock, err := HoldGuard(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireLock(dir, "test"); err == nil {
		t.Fatal("AcquireLock must fail while the guard is held")
	}
	unlock()
	release, err := AcquireLock(dir, "test")
	if err != nil {
		t.Fatalf("AcquireLock after guard release: %v", err)
	}
	release()
}

func TestDeleteAndRenameRefuseLockedProfile(t *testing.T) {
	setupTestEnv(t)
	p, err := Create("busy", "chrome")
	if err != nil {
		t.Fatal(err)
	}
	release, err := AcquireLock(p.DataDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := Delete("busy", false); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("Delete of a locked profile: %v", err)
	}
	if err := Rename("busy", "other"); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("Rename of a locked profile: %v", err)
	}
	if _, err := os.Stat(p.DataDir); err != nil {
		t.Fatalf("data dir must survive: %v", err)
	}
	release()
	if err := Rename("busy", "other"); err != nil {
		t.Fatalf("Rename after release: %v", err)
	}
	q, err := Get("other")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(q.DataDir, lockFileName)); !os.IsNotExist(err) {
		t.Fatalf("rename must not leave lock metadata behind: %v", err)
	}
	if err := Delete("other", false); err != nil {
		t.Fatalf("Delete after release: %v", err)
	}
}

func TestDeleteForceIgnoresLock(t *testing.T) {
	setupTestEnv(t)
	p, err := Create("forced", "chrome")
	if err != nil {
		t.Fatal(err)
	}
	release, err := AcquireLock(p.DataDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := Delete("forced", true); err != nil {
		t.Fatalf("forced Delete: %v", err)
	}
	if _, err := os.Stat(p.DataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir must be removed: %v", err)
	}
}
