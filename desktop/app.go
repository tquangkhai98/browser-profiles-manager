package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/install"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

// App is the Wails application backend.
// All exported methods are automatically available to the frontend via window.go.main.App.
type App struct {
	ctx     context.Context
	version string
	commit  string
	date    string
}

// NewApp creates the application struct.
func NewApp() *App {
	return &App{version: "dev", commit: "none", date: "unknown"}
}

// SetBuildInfo injects build-time variables from ldflags.
func (a *App) SetBuildInfo(version, commit, date string) {
	a.version = version
	a.commit = commit
	a.date = date
}

// startup is called when the app starts. The context is saved for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Data types returned to frontend ---

// ProfileInfo is a frontend-friendly representation of a profile with status.
type ProfileInfo struct {
	Name      string `json:"name"`
	Browser   string `json:"browser"`
	DataDir   string `json:"data_dir"`
	CreatedAt string `json:"created_at"`
	LastUsed  string `json:"last_used"`
	Locked    bool   `json:"locked"`
	LockPID   int    `json:"lock_pid"`
	LockBy    string `json:"lock_by"`
}

// BrowserItem is a frontend-friendly browser description.
type BrowserItem struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	ExePath string `json:"exe_path"`
}

// CredentialSite is a frontend-friendly credential entry.
type CredentialSite struct {
	Domain  string `json:"domain"`
	Cookies int    `json:"cookies"`
	Logins  int    `json:"logins"`
}

// CredentialResult holds credential inspection results.
type CredentialResult struct {
	ProfileName  string           `json:"profile_name"`
	Sites        []CredentialSite `json:"sites"`
	TotalCookies int              `json:"total_cookies"`
	TotalLogins  int              `json:"total_logins"`
}

// SyncResult holds the result of a credential sync operation.
type SyncResult struct {
	FilesCopied int    `json:"files_copied"`
	Message     string `json:"message"`
}

// --- Profile operations ---

// ListProfiles returns all profiles with their lock status.
func (a *App) ListProfiles() ([]ProfileInfo, error) {
	statuses, err := profile.List()
	if err != nil {
		return nil, err
	}

	infos := make([]ProfileInfo, 0, len(statuses))
	for _, s := range statuses {
		info := ProfileInfo{
			Name:      s.Name,
			Browser:   s.Browser,
			DataDir:   s.DataDir,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
			Locked:    s.Locked,
		}
		if s.LastUsed != nil {
			info.LastUsed = s.LastUsed.Format(time.RFC3339)
		}
		if s.LockInfo != nil {
			info.LockPID = s.LockInfo.PID
			info.LockBy = s.LockInfo.Caller
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// CreateProfile creates a new isolated browser profile.
func (a *App) CreateProfile(name, browserID string) error {
	_, err := profile.Create(name, browserID)
	return err
}

// DeleteProfile removes a profile and its data.
func (a *App) DeleteProfile(name string, force bool) error {
	return profile.Delete(name, force)
}

// RenameProfile changes a profile's name.
func (a *App) RenameProfile(oldName, newName string) error {
	return profile.Rename(oldName, newName)
}

// OpenProfileDir opens the profile's data directory in the OS file manager.
func (a *App) OpenProfileDir(name string) error {
	p, err := profile.Get(name)
	if err != nil {
		return err
	}
	return exec.Command("open", p.DataDir).Start()
}

// --- Browser operations ---

// DetectBrowsers returns all installed Chromium-based browsers.
func (a *App) DetectBrowsers() []BrowserItem {
	browsers := browser.DetectBrowsers()
	items := make([]BrowserItem, 0, len(browsers))
	for _, b := range browsers {
		items = append(items, BrowserItem{
			Name:    b.Name,
			ID:      b.ID,
			ExePath: b.ExePath,
		})
	}
	return items
}

// LaunchBrowser launches a browser with the specified profile.
func (a *App) LaunchBrowser(name string) error {
	p, err := profile.Get(name)
	if err != nil {
		return err
	}

	// Find the browser executable
	var browserPath string
	if p.Browser != "" {
		b, err := browser.FindBrowser(p.Browser)
		if err != nil {
			return err
		}
		browserPath = b.ExePath
	} else {
		b, err := browser.DefaultBrowser()
		if err != nil {
			return err
		}
		browserPath = b.ExePath
	}

	// Acquire lock
	release, err := profile.AcquireLock(p.DataDir, "bpm-desktop")
	if err != nil {
		return err
	}

	// Launch browser
	cmd, err := browser.Launch(browserPath, p.DataDir)
	if err != nil {
		release()
		return err
	}

	// Update last used
	profile.UpdateLastUsed(name)

	// Wait for browser to exit in background, then release lock
	go func() {
		cmd.Wait()
		release()
	}()

	return nil
}

// --- Credential operations ---

// InspectCredentials reads cookies/logins for a profile.
func (a *App) InspectCredentials(name string) (*CredentialResult, error) {
	p, err := profile.Get(name)
	if err != nil {
		return nil, err
	}

	result, err := credential.Inspect(p.DataDir, name)
	if err != nil {
		return nil, err
	}

	cr := &CredentialResult{
		ProfileName:  result.ProfileName,
		TotalCookies: result.TotalCookies,
		TotalLogins:  result.TotalLogins,
	}
	for _, s := range result.Sites {
		cr.Sites = append(cr.Sites, CredentialSite{
			Domain:  s.Domain,
			Cookies: s.CookieCount,
			Logins:  s.LoginCount,
		})
	}
	return cr, nil
}

// SyncCredentials copies credential database files from source to target profile.
func (a *App) SyncCredentials(srcName, dstName string) (*SyncResult, error) {
	srcProfile, err := profile.Get(srcName)
	if err != nil {
		return nil, fmt.Errorf("source profile: %w", err)
	}
	dstProfile, err := profile.Get(dstName)
	if err != nil {
		return nil, fmt.Errorf("target profile: %w", err)
	}

	copied, err := credential.Sync(srcProfile.DataDir, dstProfile.DataDir)
	if err != nil {
		return nil, err
	}

	return &SyncResult{
		FilesCopied: copied,
		Message:     fmt.Sprintf("Synced %d credential files from %q to %q", copied, srcName, dstName),
	}, nil
}

// --- Import operations ---

// SelectDirectory opens a native directory picker dialog.
func (a *App) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Chrome Profile Directory",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// ImportProfile imports an existing Chrome profile directory into bpm.
func (a *App) ImportProfile(srcPath, name string) error {
	// Verify source exists
	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("source path %q not found: %w", srcPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path %q is not a directory", srcPath)
	}

	// Create the profile first (validates name, creates dir)
	p, err := profile.Create(name, "chrome")
	if err != nil {
		return err
	}

	// Copy contents
	if err := copyDir(srcPath, p.DataDir); err != nil {
		profile.Delete(name, true)
		return fmt.Errorf("import failed: %w", err)
	}

	return nil
}

// --- Settings operations ---

// SettingsInfo holds settings data for the frontend.
type SettingsInfo struct {
	DefaultBrowser string `json:"default_browser"`
	ProfilesDir    string `json:"profiles_dir"`
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	BuildDate      string `json:"build_date"`
	ConfigDir      string `json:"config_dir"`
}

// GetSettings returns current settings.
func (a *App) GetSettings() (*SettingsInfo, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	profilesDir, _ := config.ProfilesDir()
	configDir, _ := config.ConfigDir()

	return &SettingsInfo{
		DefaultBrowser: cfg.DefaultBrowser,
		ProfilesDir:    profilesDir,
		Version:        a.version,
		Commit:         a.commit,
		BuildDate:      a.date,
		ConfigDir:      configDir,
	}, nil
}

// SaveDefaultBrowser updates the default browser setting.
func (a *App) SaveDefaultBrowser(browserID string) error {
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	cfg.DefaultBrowser = browserID
	return config.SaveWithLock(cfg)
}

// ChangeProfilesDir opens a directory picker and sets it as the custom profiles directory.
// Returns the selected path, or empty string if cancelled.
func (a *App) ChangeProfilesDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Profiles Directory",
	})
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", nil // User cancelled
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create directory: %w", err)
	}

	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return "", err
	}
	defer unlock()

	cfg.CustomProfilesDir = dir
	if err := config.SaveWithLock(cfg); err != nil {
		return "", err
	}
	return dir, nil
}

// ResetSettings resets all settings to defaults while preserving profiles.
func (a *App) ResetSettings() error {
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	cfg.DefaultBrowser = ""
	cfg.CustomProfilesDir = ""
	cfg.Mappings = nil
	return config.SaveWithLock(cfg)
}

// OpenConfigDir opens the config directory in the OS file manager.
func (a *App) OpenConfigDir() error {
	configDir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	return exec.Command("open", configDir).Start()
}

// GetMCPConfig returns the MCP server JSON configuration snippet.
// It auto-detects the bpm binary path for accuracy.
func (a *App) GetMCPConfig() string {
	bpmPath := "bpm"
	if resolved, err := exec.LookPath("bpm"); err == nil {
		bpmPath = resolved
	}

	cfg := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"bpm": map[string]interface{}{
				"command": bpmPath,
				"args":    []string{"serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// ReorderProfiles saves a new profile ordering.
// names must contain exactly the same profile names as currently exist, in the desired order.
func (a *App) ReorderProfiles(names []string) error {
	cfg, unlock, err := config.LoadWithLock()
	if err != nil {
		return err
	}
	defer unlock()

	if len(names) != len(cfg.Profiles) {
		return fmt.Errorf("expected %d profile names, got %d", len(cfg.Profiles), len(names))
	}

	// Build lookup by name
	byName := make(map[string]config.Profile, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		byName[p.Name] = p
	}

	// Rebuild in requested order
	reordered := make([]config.Profile, 0, len(names))
	for _, name := range names {
		p, ok := byName[name]
		if !ok {
			return fmt.Errorf("profile %q not found", name)
		}
		reordered = append(reordered, p)
	}

	cfg.Profiles = reordered
	return config.SaveWithLock(cfg)
}

// ExportAllProfiles exports all profiles to the selected directory.
func (a *App) ExportAllProfiles() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Export Directory",
	})
	if err != nil || dir == "" {
		return "", err
	}

	statuses, err := profile.List()
	if err != nil {
		return "", err
	}

	exported := 0
	for _, s := range statuses {
		dstPath := filepath.Join(dir, s.Name)
		if err := copyDir(s.DataDir, dstPath); err != nil {
			return "", fmt.Errorf("failed to export %q: %w", s.Name, err)
		}
		exported++
	}

	return fmt.Sprintf("Exported %d profiles to %s", exported, dir), nil
}

// --- MCP Install operations ---

// IDEInfo is a frontend-friendly IDE status.
type IDEInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ConfigPath string `json:"config_path"`
	Installed  bool   `json:"installed"`
	BPMEnabled bool   `json:"bpm_enabled"`
}

// InstallResult holds the outcome of an install operation.
type InstallResult struct {
	IDE        string `json:"ide"`
	ConfigPath string `json:"config_path"`
	Action     string `json:"action"`
	Message    string `json:"message"`
}

// ListIDEs returns all supported AI IDEs with their install status.
func (a *App) ListIDEs() []IDEInfo {
	ides := install.SupportedIDEs()
	result := make([]IDEInfo, 0, len(ides))
	for _, ide := range ides {
		result = append(result, IDEInfo{
			ID:         ide.ID,
			Name:       ide.Name,
			ConfigPath: ide.ConfigPath,
			Installed:  ide.Installed,
			BPMEnabled: ide.BPMEnabled,
		})
	}
	return result
}

// InstallMCP adds bpm MCP config to the specified IDE.
func (a *App) InstallMCP(ideID string) (*InstallResult, error) {
	result, err := install.Install(ideID)
	if err != nil {
		return nil, err
	}
	return &InstallResult{
		IDE:        result.IDE,
		ConfigPath: result.ConfigPath,
		Action:     result.Action,
		Message:    result.Message,
	}, nil
}

// UninstallMCP removes bpm MCP config from the specified IDE.
func (a *App) UninstallMCP(ideID string) (*InstallResult, error) {
	result, err := install.Uninstall(ideID)
	if err != nil {
		return nil, err
	}
	return &InstallResult{
		IDE:        result.IDE,
		ConfigPath: result.ConfigPath,
		Action:     result.Action,
		Message:    result.Message,
	}, nil
}

// --- Helper functions ---

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode()|0700)
		}

		// Skip very large files (cache) to speed up import
		if info.Size() > 500*1024*1024 {
			return nil
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
