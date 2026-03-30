package mapping

import (
	"path/filepath"
	"testing"

	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
)

// setupTestEnv sets up isolated config and data dirs for testing.
func setupTestEnv(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("BPM_CONFIG_DIR", filepath.Join(tmpDir, "config"))
	t.Setenv("BPM_DATA_DIR", filepath.Join(tmpDir, "data"))
}

// createTestProfile adds a profile directly to config for mapping tests.
func createTestProfile(t *testing.T, name string) {
	t.Helper()
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		t.Fatalf("LoadWithLock() error = %v", err)
	}
	defer unlock()

	cfg.Profiles = append(cfg.Profiles, config.Profile{
		Name:      name,
		Browser:   "chrome",
		DataDir:   filepath.Join(t.TempDir(), name),
		CreatedAt: "2026-01-01T00:00:00Z",
	})
	if err := config.SaveWithLock(cfg); err != nil {
		t.Fatalf("SaveWithLock() error = %v", err)
	}
}

func TestSet_NewMapping(t *testing.T) {
	setupTestEnv(t)
	createTestProfile(t, "work")

	dir := t.TempDir()
	if err := Set(dir, "work"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify mapping was saved
	mappings, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(mappings))
	}
}

func TestSet_UpdateExisting(t *testing.T) {
	setupTestEnv(t)
	createTestProfile(t, "old")
	createTestProfile(t, "new")

	dir := t.TempDir()
	if err := Set(dir, "old"); err != nil {
		t.Fatalf("Set(old) error = %v", err)
	}
	if err := Set(dir, "new"); err != nil {
		t.Fatalf("Set(new) error = %v", err)
	}

	// Should still be 1 mapping, not 2
	mappings, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping after update, got %d", len(mappings))
	}
	if mappings[0].Profile != "new" {
		t.Errorf("mapping profile = %q, want %q", mappings[0].Profile, "new")
	}
}

func TestSet_ProfileNotFound(t *testing.T) {
	setupTestEnv(t)

	if err := Set("/tmp/some-dir", "nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
}

func TestGet_ExactMatch(t *testing.T) {
	setupTestEnv(t)
	createTestProfile(t, "exact")

	dir := t.TempDir()
	if err := Set(dir, "exact"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	profile, err := Get(dir)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile != "exact" {
		t.Errorf("Get() = %q, want %q", profile, "exact")
	}
}

func TestGet_ParentMatch(t *testing.T) {
	setupTestEnv(t)
	createTestProfile(t, "parent")

	parentDir := t.TempDir()
	if err := Set(parentDir, "parent"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	childDir := filepath.Join(parentDir, "subdir", "deep")

	profile, err := Get(childDir)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile != "parent" {
		t.Errorf("Get(child) = %q, want %q (parent match)", profile, "parent")
	}
}

func TestGet_NoMatch(t *testing.T) {
	setupTestEnv(t)

	profile, err := Get("/tmp/unmapped-dir")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile != "" {
		t.Errorf("Get() = %q, want empty string", profile)
	}
}

func TestRemove_Success(t *testing.T) {
	setupTestEnv(t)
	createTestProfile(t, "removable")

	dir := t.TempDir()
	if err := Set(dir, "removable"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if err := Remove(dir); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	mappings, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(mappings) != 0 {
		t.Errorf("expected 0 mappings after remove, got %d", len(mappings))
	}
}

func TestRemove_NotFound(t *testing.T) {
	setupTestEnv(t)

	err := Remove("/tmp/no-such-mapping")
	if err == nil {
		t.Fatal("expected error for removing nonexistent mapping, got nil")
	}
}
