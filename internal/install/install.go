// Package install handles MCP server configuration for AI IDEs.
// It reads/writes IDE-specific config files to register bpm as an MCP server.
package install

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// IDE represents a supported AI IDE.
type IDE struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ConfigPath string `json:"config_path"`
	Installed  bool   `json:"installed"`
	BPMEnabled bool   `json:"bpm_enabled"`
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
	servers, err := serversForWrite(cfg, target.ConfigPath)
	if err != nil {
		return nil, err
	}
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

	servers, err := serversForWrite(cfg, target.ConfigPath)
	if err != nil {
		return nil, err
	}
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

// serversForWrite returns cfg's mcpServers map for modification. It refuses
// to continue when mcpServers exists but is not a JSON object, so a value bpm
// doesn't understand is never silently replaced.
func serversForWrite(cfg map[string]interface{}, path string) (map[string]interface{}, error) {
	raw, ok := cfg["mcpServers"]
	if !ok || raw == nil {
		return make(map[string]interface{}), nil
	}
	servers, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s: \"mcpServers\" is not a JSON object; fix it by hand before running bpm install", path)
	}
	return servers, nil
}

// writeConfig atomically replaces the IDE config at path with cfg.
//
// A symlinked config is written at its real target so dotfile-manager links
// survive. An existing config is first copied to <real path>.bpm.bak. The new
// file keeps the original file's permissions (0600 for a new file) and is
// fsynced before the rename; the parent directory is fsynced afterwards.
func writeConfig(path string, cfg map[string]interface{}) error {
	realPath, err := resolveConfigPath(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(realPath)

	// Ensure the (real) parent directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	mode := os.FileMode(0600)
	if info, err := os.Stat(realPath); err == nil {
		mode = info.Mode().Perm()
		if err := backupFile(realPath, realPath+".bpm.bak", mode); err != nil {
			return fmt.Errorf("cannot back up %s: %w", realPath, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Atomic write: temp file + fsync + rename
	tmpFile, err := os.CreateTemp(dir, ".bpm-mcp-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Chmod(mode); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, realPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	syncDir(dir)
	return nil
}

// resolveConfigPath follows symlinks at path. A missing file resolves to
// itself (with symlinks in its parent directory resolved when possible); a
// dangling symlink resolves to the link's target.
func resolveConfigPath(path string) (string, error) {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return real, nil
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if parent, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
			return filepath.Join(parent, filepath.Base(path)), nil
		}
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return path, nil
	}
	// Dangling symlink: write where it points so the link becomes valid.
	target, err := os.Readlink(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return target, nil
}

// backupFile copies src to dst (replacing dst) with the given mode.
func backupFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// syncDir fsyncs a directory so a rename into it is durable. Best effort;
// skipped on Windows, where directories can't be opened for Sync.
func syncDir(dir string) {
	if runtime.GOOS == "windows" {
		return
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		d.Close()
	}
}

// resolveBPMPath returns the absolute path of the bpm CLI to register as the
// MCP command. It prefers the running binary when that is the bpm CLI (the
// desktop app is a different binary and has no "serve" command), then bpm on
// PATH, and only falls back to a bare "bpm" as a last resort.
func resolveBPMPath() string {
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		name := strings.TrimSuffix(strings.ToLower(filepath.Base(exe)), ".exe")
		if name == "bpm" {
			return exe
		}
	}
	if resolved, err := exec.LookPath("bpm"); err == nil {
		if abs, err := filepath.Abs(resolved); err == nil {
			return abs
		}
		return resolved
	}
	return "bpm"
}
