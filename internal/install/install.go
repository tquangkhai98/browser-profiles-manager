// Package install handles MCP server configuration for AI IDEs.
// It reads/writes IDE-specific config files to register bpm as an MCP server.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// IDE represents a supported AI IDE.
type IDE struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ConfigPath  string `json:"config_path"`
	Installed   bool   `json:"installed"`
	BPMEnabled  bool   `json:"bpm_enabled"`
}

// Result holds the outcome of an install/uninstall operation.
type Result struct {
	IDE        string `json:"ide"`
	ConfigPath string `json:"config_path"`
	Action     string `json:"action"` // "installed" | "already_installed" | "uninstalled" | "not_installed"
	Message    string `json:"message"`
}

// SupportedIDEs returns all known IDE targets with their config paths.
func SupportedIDEs() []IDE {
	home, _ := os.UserHomeDir()
	ides := []IDE{
		{
			ID:         "claude-code",
			Name:       "Claude Code",
			ConfigPath: filepath.Join(home, ".claude", "mcp_servers.json"),
		},
		{
			ID:         "claude-desktop",
			Name:       "Claude Desktop",
			ConfigPath: claudeDesktopConfigPath(home),
		},
		{
			ID:         "cursor",
			Name:       "Cursor",
			ConfigPath: filepath.Join(home, ".cursor", "mcp.json"),
		},
		{
			ID:         "antigravity",
			Name:       "Antigravity",
			ConfigPath: filepath.Join(home, ".gemini", "settings.json"),
		},
	}

	// Check installed status and bpm config presence
	for i := range ides {
		ides[i].Installed = configDirExists(ides[i].ConfigPath)
		ides[i].BPMEnabled = hasBPMConfig(ides[i].ConfigPath)
	}

	return ides
}

// Install adds bpm MCP server config to the specified IDE.
func Install(ideID string) (*Result, error) {
	ides := SupportedIDEs()
	var target *IDE
	for i, ide := range ides {
		if ide.ID == ideID {
			target = &ides[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("unknown IDE %q. Supported: claude-code, claude-desktop, cursor, antigravity", ideID)
	}

	// Check if already configured
	if target.BPMEnabled {
		return &Result{
			IDE:        target.Name,
			ConfigPath: target.ConfigPath,
			Action:     "already_installed",
			Message:    fmt.Sprintf("bpm MCP is already configured in %s", target.Name),
		}, nil
	}

	// Read or create config
	cfg, err := readOrCreateConfig(target.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Add bpm entry
	bpmPath := resolveBPMPath()
	servers := getOrCreateServers(cfg)
	servers["bpm"] = map[string]interface{}{
		"command": bpmPath,
		"args":    []interface{}{"serve"},
	}
	cfg["mcpServers"] = servers

	// Write config
	if err := writeConfig(target.ConfigPath, cfg); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	return &Result{
		IDE:        target.Name,
		ConfigPath: target.ConfigPath,
		Action:     "installed",
		Message:    fmt.Sprintf("✅ bpm MCP installed to %s", target.Name),
	}, nil
}

// Uninstall removes bpm MCP server config from the specified IDE.
func Uninstall(ideID string) (*Result, error) {
	ides := SupportedIDEs()
	var target *IDE
	for i, ide := range ides {
		if ide.ID == ideID {
			target = &ides[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("unknown IDE %q", ideID)
	}

	if !target.BPMEnabled {
		return &Result{
			IDE:        target.Name,
			ConfigPath: target.ConfigPath,
			Action:     "not_installed",
			Message:    fmt.Sprintf("bpm MCP is not configured in %s", target.Name),
		}, nil
	}

	cfg, err := readOrCreateConfig(target.ConfigPath)
	if err != nil {
		return nil, err
	}

	servers := getOrCreateServers(cfg)
	delete(servers, "bpm")
	cfg["mcpServers"] = servers

	if err := writeConfig(target.ConfigPath, cfg); err != nil {
		return nil, err
	}

	return &Result{
		IDE:        target.Name,
		ConfigPath: target.ConfigPath,
		Action:     "uninstalled",
		Message:    fmt.Sprintf("🗑️ bpm MCP removed from %s", target.Name),
	}, nil
}

// --- helpers ---

func claudeDesktopConfigPath(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appdata, "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(home, ".config", "claude", "claude_desktop_config.json")
	}
}

func configDirExists(configPath string) bool {
	dir := filepath.Dir(configPath)
	_, err := os.Stat(dir)
	return err == nil
}

func hasBPMConfig(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	servers := getOrCreateServers(cfg)
	_, exists := servers["bpm"]
	return exists
}

func readOrCreateConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	return cfg, nil
}

func getOrCreateServers(cfg map[string]interface{}) map[string]interface{} {
	raw, ok := cfg["mcpServers"]
	if !ok {
		return make(map[string]interface{})
	}
	servers, ok := raw.(map[string]interface{})
	if !ok {
		return make(map[string]interface{})
	}
	return servers
}

func writeConfig(path string, cfg map[string]interface{}) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file + rename
	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".bpm-mcp-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path)
}

func resolveBPMPath() string {
	if resolved, err := exec.LookPath("bpm"); err == nil {
		return resolved
	}
	return "bpm"
}
