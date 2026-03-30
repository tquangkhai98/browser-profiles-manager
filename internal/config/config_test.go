package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupTestConfig creates a temp config environment for testing.
// Returns cleanup function.
func setupTestConfig(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	// Override config and data dirs via env vars
	t.Setenv("BPM_CONFIG_DIR", filepath.Join(tmpDir, "config"))
	t.Setenv("BPM_DATA_DIR", filepath.Join(tmpDir, "data"))
}

func TestLoad_DefaultConfig(t *testing.T) {
	setupTestConfig(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Profiles == nil {
		t.Error("Profiles should be initialized, not nil")
	}
	if cfg.Mappings == nil {
		t.Error("Mappings should be initialized, not nil")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	setupTestConfig(t)

	cfg := &Config{
		DefaultBrowser: "chrome",
		Profiles: []Profile{
			{Name: "test", Browser: "chrome", DataDir: "/tmp/test", CreatedAt: "2026-01-01T00:00:00Z"},
		},
		Mappings: []Mapping{},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was written
	configPath, _ := ConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("cannot read saved config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("saved config is not valid JSON: %v", err)
	}
	if loaded.DefaultBrowser != "chrome" {
		t.Errorf("DefaultBrowser = %q, want %q", loaded.DefaultBrowser, "chrome")
	}
	if len(loaded.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(loaded.Profiles))
	}
	if loaded.Profiles[0].Name != "test" {
		t.Errorf("profile name = %q, want %q", loaded.Profiles[0].Name, "test")
	}
}

func TestLoadWithLock_ReturnsUnlockFunc(t *testing.T) {
	setupTestConfig(t)

	cfg, unlock, err := LoadWithLock()
	if err != nil {
		t.Fatalf("LoadWithLock() error = %v", err)
	}
	defer unlock()

	if cfg == nil {
		t.Fatal("LoadWithLock() returned nil config")
	}
}

func TestSaveWithLock_PersistsChanges(t *testing.T) {
	setupTestConfig(t)

	cfg, unlock, err := LoadWithLock()
	if err != nil {
		t.Fatalf("LoadWithLock() error = %v", err)
	}
	defer unlock()

	cfg.Profiles = append(cfg.Profiles, Profile{
		Name:      "locked-profile",
		Browser:   "brave",
		DataDir:   "/tmp/locked",
		CreatedAt: "2026-01-01T00:00:00Z",
	})

	if err := SaveWithLock(cfg); err != nil {
		t.Fatalf("SaveWithLock() error = %v", err)
	}

	// Reload and verify
	reloaded, err := Load()
	if err != nil {
		t.Fatalf("reload error = %v", err)
	}
	if len(reloaded.Profiles) != 1 {
		t.Errorf("expected 1 profile after save, got %d", len(reloaded.Profiles))
	}
}

func TestConfigDir_ReturnsNonEmpty(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}
}

func TestProfilesDir_ReturnsNonEmpty(t *testing.T) {
	dir, err := ProfilesDir()
	if err != nil {
		t.Fatalf("ProfilesDir() error = %v", err)
	}
	if dir == "" {
		t.Error("ProfilesDir() returned empty string")
	}
}
