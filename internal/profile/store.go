package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// Create creates a new isolated browser profile.
func Create(name, browser string) (*Profile, error) {
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("invalid profile name %q: must start with alphanumeric and contain only alphanumeric, hyphens, underscores", name)
	}

	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}
	defer unlock()

	// Check for duplicate name
	for _, p := range cfg.Profiles {
		if p.Name == name {
			return nil, fmt.Errorf("profile %q already exists", name)
		}
	}

	// Determine profile data directory
	profilesDir, err := config.ProfilesDir()
	if err != nil {
		return nil, err
	}
	dataDir := filepath.Join(profilesDir, name)

	// Create profile directory with restricted permissions
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create profile directory: %w", err)
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Add to config
	configProfile := config.Profile{
		Name:      name,
		Browser:   browser,
		DataDir:   dataDir,
		CreatedAt: nowStr,
	}
	cfg.Profiles = append(cfg.Profiles, configProfile)

	if err := config.SaveWithLock(cfg); err != nil {
		// Rollback: remove created directory
		os.RemoveAll(dataDir)
		return nil, fmt.Errorf("cannot save config: %w", err)
	}

	profile := &Profile{
		Name:      name,
		Browser:   browser,
		DataDir:   dataDir,
		CreatedAt: now,
	}
	return profile, nil
}

// List returns all profiles with their lock status.
func List() ([]ProfileStatus, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	statuses := make([]ProfileStatus, 0, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)
		var lastUsed *time.Time
		if p.LastUsed != nil {
			t, _ := time.Parse(time.RFC3339, *p.LastUsed)
			lastUsed = &t
		}

		profile := Profile{
			Name:      p.Name,
			Browser:   p.Browser,
			DataDir:   p.DataDir,
			CreatedAt: createdAt,
			LastUsed:  lastUsed,
		}

		locked, lockInfo := IsLocked(p.DataDir)
		statuses = append(statuses, ProfileStatus{
			Profile:  profile,
			Locked:   locked,
			LockInfo: lockInfo,
		})
	}
	return statuses, nil
}

// Get finds a profile by name.
func Get(name string) (*Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load config: %w", err)
	}

	for _, p := range cfg.Profiles {
		if p.Name == name {
			createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)
			var lastUsed *time.Time
			if p.LastUsed != nil {
				t, _ := time.Parse(time.RFC3339, *p.LastUsed)
				lastUsed = &t
			}
			return &Profile{
				Name:      p.Name,
				Browser:   p.Browser,
				DataDir:   p.DataDir,
				CreatedAt: createdAt,
				LastUsed:  lastUsed,
			}, nil
		}
	}
	return nil, fmt.Errorf("profile %q not found", name)
}

// Delete removes a profile from config and optionally deletes its data directory.
func Delete(name string, force bool) error {
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	defer unlock()

	idx := -1
	for i, p := range cfg.Profiles {
		if p.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("profile %q not found", name)
	}

	p := cfg.Profiles[idx]

	// Check if locked
	locked, lockInfo := IsLocked(p.DataDir)
	if locked && !force {
		return fmt.Errorf("profile %q is locked by PID %d (%s). Use --force to delete anyway", name, lockInfo.PID, lockInfo.Caller)
	}

	// Remove data directory
	if err := os.RemoveAll(p.DataDir); err != nil {
		if !force {
			return fmt.Errorf("cannot remove profile directory: %w", err)
		}
		// If force, continue even if directory removal fails
		fmt.Fprintf(os.Stderr, "Warning: cannot remove profile directory %s: %v\n", p.DataDir, err)
	}

	// Remove from config
	cfg.Profiles = append(cfg.Profiles[:idx], cfg.Profiles[idx+1:]...)

	// Also clean up any mappings pointing to this profile
	cleanMappings := make([]config.Mapping, 0, len(cfg.Mappings))
	for _, m := range cfg.Mappings {
		if m.Profile != name {
			cleanMappings = append(cleanMappings, m)
		}
	}
	cfg.Mappings = cleanMappings

	if err := config.SaveWithLock(cfg); err != nil {
		return fmt.Errorf("cannot save config: %w", err)
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp for a profile.
func UpdateLastUsed(name string) error {
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	for i, p := range cfg.Profiles {
		if p.Name == name {
			nowStr := time.Now().UTC().Format(time.RFC3339)
			cfg.Profiles[i].LastUsed = &nowStr
			return config.SaveWithLock(cfg)
		}
	}
	return fmt.Errorf("profile %q not found", name)
}
