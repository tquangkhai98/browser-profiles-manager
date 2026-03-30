package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
)

// setupTestEnv sets up isolated config and data dirs for testing.
func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("BPM_CONFIG_DIR", filepath.Join(tmpDir, "config"))
	t.Setenv("BPM_DATA_DIR", filepath.Join(tmpDir, "data"))
	return tmpDir
}

func TestCreate_Success(t *testing.T) {
	setupTestEnv(t)

	p, err := Create("test-profile", "chrome")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if p.Name != "test-profile" {
		t.Errorf("Name = %q, want %q", p.Name, "test-profile")
	}
	if p.Browser != "chrome" {
		t.Errorf("Browser = %q, want %q", p.Browser, "chrome")
	}

	// Verify directory was created
	if _, err := os.Stat(p.DataDir); os.IsNotExist(err) {
		t.Error("Profile data directory was not created")
	}

	// Verify config was updated
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile in config, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Name != "test-profile" {
		t.Errorf("config profile name = %q, want %q", cfg.Profiles[0].Name, "test-profile")
	}
}

func TestCreate_DuplicateName(t *testing.T) {
	setupTestEnv(t)

	_, err := Create("dup", "chrome")
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = Create("dup", "chrome")
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestCreate_InvalidName(t *testing.T) {
	setupTestEnv(t)

	tests := []struct {
		name string
	}{
		{""},
		{"-starts-with-dash"},
		{"has spaces"},
		{"special!chars"},
		{"../path-traversal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Create(tt.name, "chrome")
			if err == nil {
				t.Errorf("Create(%q) expected error, got nil", tt.name)
			}
		})
	}
}

func TestList_Empty(t *testing.T) {
	setupTestEnv(t)

	profiles, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestList_MultipleProfiles(t *testing.T) {
	setupTestEnv(t)

	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		if _, err := Create(name, "chrome"); err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
	}

	profiles, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(profiles))
	}

	// All should be unlocked
	for _, p := range profiles {
		if p.Locked {
			t.Errorf("profile %q should not be locked", p.Name)
		}
	}
}

func TestGet_Found(t *testing.T) {
	setupTestEnv(t)

	_, err := Create("findme", "brave")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	p, err := Get("findme")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if p.Name != "findme" {
		t.Errorf("Name = %q, want %q", p.Name, "findme")
	}
	if p.Browser != "brave" {
		t.Errorf("Browser = %q, want %q", p.Browser, "brave")
	}
}

func TestGet_NotFound(t *testing.T) {
	setupTestEnv(t)

	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
}

func TestDelete_Success(t *testing.T) {
	setupTestEnv(t)

	p, err := Create("deleteme", "chrome")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	dataDir := p.DataDir

	if err := Delete("deleteme", false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify directory removed
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Error("Profile data directory was not removed")
	}

	// Verify removed from config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles after delete, got %d", len(cfg.Profiles))
	}
}

func TestDelete_CleansUpMappings(t *testing.T) {
	setupTestEnv(t)

	_, err := Create("mapped-profile", "chrome")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Add a mapping manually
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		t.Fatalf("LoadWithLock() error = %v", err)
	}
	cfg.Mappings = append(cfg.Mappings, config.Mapping{
		Directory: "/tmp/test-project",
		Profile:   "mapped-profile",
	})
	if err := config.SaveWithLock(cfg); err != nil {
		unlock()
		t.Fatalf("SaveWithLock() error = %v", err)
	}
	unlock()

	// Delete the profile
	if err := Delete("mapped-profile", true); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify mapping was cleaned up
	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if len(cfg.Mappings) != 0 {
		t.Errorf("expected 0 mappings after profile delete, got %d", len(cfg.Mappings))
	}
}

func TestDelete_NotFound(t *testing.T) {
	setupTestEnv(t)

	err := Delete("ghost", false)
	if err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
}

func TestUpdateLastUsed(t *testing.T) {
	setupTestEnv(t)

	_, err := Create("useme", "chrome")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := UpdateLastUsed("useme"); err != nil {
		t.Fatalf("UpdateLastUsed() error = %v", err)
	}

	p, err := Get("useme")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if p.LastUsed == nil {
		t.Error("LastUsed should be set after UpdateLastUsed")
	}
}
