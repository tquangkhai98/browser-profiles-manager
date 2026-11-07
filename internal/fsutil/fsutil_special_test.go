//go:build !windows

package fsutil

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestCopyDirSkipsSymlinksAndSpecialFiles(t *testing.T) {
	base := t.TempDir()
	src, dst := filepath.Join(base, "src"), filepath.Join(base, "dst")
	if err := os.MkdirAll(src, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "Cookies"), []byte("db"), 0600); err != nil {
		t.Fatal(err)
	}
	// Chromium leaves a dangling SingletonLock symlink behind.
	if err := os.Symlink("host-12345", filepath.Join(src, "SingletonLock")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(src, "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "Cookies")); err != nil {
		t.Fatalf("regular file not copied: %v", err)
	}
	for _, name := range []string{"SingletonLock", "fifo"} {
		if _, err := os.Lstat(filepath.Join(dst, name)); !os.IsNotExist(err) {
			t.Errorf("%s should have been skipped, Lstat err = %v", name, err)
		}
	}
}

func TestCopyDirRejectsNonEmptyDest(t *testing.T) {
	base := t.TempDir()
	src, dst := filepath.Join(base, "src"), filepath.Join(base, "dst")
	if err := os.MkdirAll(src, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "stale"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := CopyDir(src, dst); err == nil {
		t.Fatal("expected error for non-empty destination")
	}
}

func TestCopyDirNestedDest(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(src, "bak")
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "f")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "bak")); !os.IsNotExist(err) {
		t.Fatalf("destination was copied into itself: %v", err)
	}
}
