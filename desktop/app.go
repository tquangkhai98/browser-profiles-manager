package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

// App is the Wails application backend.
// All exported methods are automatically available to the frontend via window.go.main.App.
type App struct {
	ctx context.Context
}

// NewApp creates the application struct.
func NewApp() *App {
	return &App{}
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
}

// GetSettings returns current settings.
func (a *App) GetSettings() (*SettingsInfo, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	profilesDir, _ := config.ProfilesDir()

	return &SettingsInfo{
		DefaultBrowser: cfg.DefaultBrowser,
		ProfilesDir:    profilesDir,
		Version:        "1.0.0",
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

// GetMCPConfig returns the MCP server JSON configuration snippet.
func (a *App) GetMCPConfig() string {
	return `{
  "mcpServers": {
    "bpm": {
      "command": "bpm",
      "args": ["serve"]
    }
  }
}`
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
