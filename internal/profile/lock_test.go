package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireLock_Success(t *testing.T) {
	tmpDir := t.TempDir()

	release, err := AcquireLock(tmpDir, "test-caller")
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer release()

	locked, info := IsLocked(tmpDir)
	if !locked {
		t.Error("expected profile to be locked")
	}
	if info == nil {
		t.Fatal("expected lock info, got nil")
	}
	if info.Caller != "test-caller" {
		t.Errorf("Caller = %q, want %q", info.Caller, "test-caller")
	}
}

func TestAcquireLock_AlreadyLocked(t *testing.T) {
	tmpDir := t.TempDir()

	release, err := AcquireLock(tmpDir, "first")
	if err != nil {
		t.Fatalf("first AcquireLock() error = %v", err)
	}
	defer release()

	_, err = AcquireLock(tmpDir, "second")
	if err == nil {
		t.Fatal("expected error for double lock, got nil")
	}
}

func TestIsLocked_Free(t *testing.T) {
	tmpDir := t.TempDir()

	locked, info := IsLocked(tmpDir)
	if locked {
		t.Error("expected profile to be free")
	}
	if info != nil {
		t.Errorf("expected nil lock info, got %+v", info)
	}
}

func TestIsLocked_StaleLock(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a lock file with a PID that doesn't exist
	lockPath := filepath.Join(tmpDir, lockFileName)
	fakeData := []byte(`{"pid": 999999999, "caller": "dead-process", "locked_at": "2026-01-01T00:00:00Z"}`)
	if err := writeTestFile(t, lockPath, fakeData); err != nil {
		t.Fatalf("cannot write fake lock: %v", err)
	}

	locked, info := IsLocked(tmpDir)
	if locked {
		t.Error("stale lock should have been cleaned up, but IsLocked returned true")
	}
	if info != nil {
		t.Errorf("expected nil lock info after stale cleanup, got %+v", info)
	}
}

func TestReleaseLock(t *testing.T) {
	tmpDir := t.TempDir()

	release, err := AcquireLock(tmpDir, "release-test")
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}

	release()

	locked, _ := IsLocked(tmpDir)
	if locked {
		t.Error("expected profile to be free after release")
	}
}

func writeTestFile(t *testing.T, path string, data []byte) error {
	t.Helper()
	return os.WriteFile(path, data, 0600)
}
