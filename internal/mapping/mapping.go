package mapping

import (
	"fmt"
	"path/filepath"

	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
)

// Set maps a project directory to a profile.
func Set(dir, profileName string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot resolve path %q: %w", dir, err)
	}

	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	// Verify profile exists
	found := false
	for _, p := range cfg.Profiles {
		if p.Name == profileName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile %q not found", profileName)
	}

	// Update existing mapping or add new one
	for i, m := range cfg.Mappings {
		if m.Directory == absDir {
			cfg.Mappings[i].Profile = profileName
			return config.SaveWithLock(cfg)
		}
	}

	cfg.Mappings = append(cfg.Mappings, config.Mapping{
		Directory: absDir,
		Profile:   profileName,
	})
	return config.SaveWithLock(cfg)
}

// Get resolves the profile for a given directory.
// Returns the profile name or empty string if no mapping exists.
func Get(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path %q: %w", dir, err)
	}

	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	// Exact match first
	for _, m := range cfg.Mappings {
		if m.Directory == absDir {
			return m.Profile, nil
		}
	}

	// Walk up parent directories for closest match
	current := absDir
	for {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		for _, m := range cfg.Mappings {
			if m.Directory == parent {
				return m.Profile, nil
			}
		}
		current = parent
	}

	return "", nil
}

// List returns all directory → profile mappings.
func List() ([]config.Mapping, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return cfg.Mappings, nil
}

// Remove removes a mapping for a directory.
func Remove(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot resolve path %q: %w", dir, err)
	}

	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	found := false
	clean := make([]config.Mapping, 0, len(cfg.Mappings))
	for _, m := range cfg.Mappings {
		if m.Directory == absDir {
			found = true
		} else {
			clean = append(clean, m)
		}
	}

	if !found {
		return fmt.Errorf("no mapping found for %q", absDir)
	}

	cfg.Mappings = clean
	return config.SaveWithLock(cfg)
}
