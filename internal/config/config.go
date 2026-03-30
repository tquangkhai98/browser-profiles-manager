package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config is the top-level bpm configuration persisted to config.json.
type Config struct {
	DefaultBrowser    string    `json:"default_browser"`
	CustomProfilesDir string    `json:"custom_profiles_dir,omitempty"`
	Profiles          []Profile `json:"profiles"`
	Mappings          []Mapping `json:"mappings"`
}

// Profile represents a single isolated browser profile.
type Profile struct {
	Name      string  `json:"name"`
	Browser   string  `json:"browser"`
	DataDir   string  `json:"data_dir"`
	CreatedAt string  `json:"created_at"`
	LastUsed  *string `json:"last_used_at,omitempty"`
}

// Mapping links a project directory to a profile.
type Mapping struct {
	Directory string `json:"directory"`
	Profile   string `json:"profile"`
}

// ConfigDir returns the platform-specific config directory.
// Respects BPM_CONFIG_DIR env override (useful for testing).
//   - macOS/Linux: ~/.config/bpm/
//   - Windows:     %APPDATA%\bpm\
func ConfigDir() (string, error) {
	if dir := os.Getenv("BPM_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "bpm"), nil
	default: // darwin, linux
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", "bpm"), nil
	}
}

// DataDir returns the platform-specific data directory for profile storage.
// Respects BPM_DATA_DIR env override (useful for testing).
//   - macOS/Linux: ~/.local/share/bpm/
//   - Windows:     %LOCALAPPDATA%\bpm\
func DataDir() (string, error) {
	if dir := os.Getenv("BPM_DATA_DIR"); dir != "" {
		return dir, nil
	}
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		return filepath.Join(localAppData, "bpm"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".local", "share", "bpm"), nil
	}
}

// ProfilesDir returns the directory where all profile data directories live.
// If a custom directory is configured, it takes priority.
func ProfilesDir() (string, error) {
	cfg, err := Load()
	if err == nil && cfg.CustomProfilesDir != "" {
		return cfg.CustomProfilesDir, nil
	}
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "profiles"), nil
}

// ConfigPath returns the full path to config.json.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads config.json from disk.
// If the file does not exist, it creates a default config.
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return nil, fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		cfg := defaultConfig()
		if saveErr := Save(cfg); saveErr != nil {
			return nil, fmt.Errorf("cannot write default config: %w", saveErr)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config.json: %w", err)
	}
	return &cfg, nil
}

// Save atomically writes the config to disk using temp file + rename.
func Save(cfg *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	// Atomic write: temp file in same directory → rename
	tmpFile, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("cannot write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot rename temp file to config: %w", err)
	}

	return nil
}

// LoadWithLock loads config with file-level locking for concurrent access safety.
func LoadWithLock() (*Config, func(), error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return nil, nil, fmt.Errorf("cannot create config directory: %w", err)
	}

	lockPath := configPath + ".lock"
	unlock, err := lockFile(lockPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot acquire config lock: %w", err)
	}

	cfg, err := Load()
	if err != nil {
		unlock()
		return nil, nil, err
	}

	return cfg, unlock, nil
}

// SaveWithLock saves config inside an already-held lock.
// The caller is responsible for calling the unlock function returned by LoadWithLock.
func SaveWithLock(cfg *Config) error {
	return Save(cfg)
}

func defaultConfig() *Config {
	return &Config{
		DefaultBrowser: "",
		Profiles:       []Profile{},
		Mappings:       []Mapping{},
	}
}
