package credential

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

// credentialFiles are the Chromium database files to sync between profiles.
var credentialFiles = []string{
	"Cookies",
	"Cookies-journal",
	"Login Data",
	"Login Data-journal",
}

// sqliteSidecarSuffixes are the SQLite side files that belong to a database.
// A leftover one in the target must not be paired with a freshly copied DB:
// SQLite would roll a hot journal (or replay a WAL) onto the new file.
var sqliteSidecarSuffixes = []string{"-journal", "-wal", "-shm"}

// Sync copies credential database files from source to target profile.
// It copies the files as-is, then encrypts any plaintext passwords
// so Chrome can read them (Chrome macOS requires v10-encrypted format).
// Files are always placed in the target's Default/ subdirectory, since Chromium
// with --user-data-dir reads from <profileDir>/Default/.
//
// Both profiles are locked for the duration of the copy, so Sync fails if
// either browser is running. Sync also refuses to copy between profiles of
// different browsers: cookie and password blobs are encrypted with a
// per-browser key and would not decrypt in the target.
func Sync(srcDir, dstDir string) (int, error) {
	if filepath.Clean(srcDir) == filepath.Clean(dstDir) {
		return 0, fmt.Errorf("source and target are the same profile")
	}
	if err := checkSameBrowser(srcDir, dstDir); err != nil {
		return 0, err
	}

	releaseSrc, err := profile.AcquireLock(srcDir, "bpm-sync")
	if err != nil {
		return 0, fmt.Errorf("source profile is in use, close the browser first: %w", err)
	}
	defer releaseSrc()
	releaseDst, err := profile.AcquireLock(dstDir, "bpm-sync")
	if err != nil {
		return 0, fmt.Errorf("target profile is in use, close the browser first: %w", err)
	}
	defer releaseDst()

	copied := 0
	dstDefault := filepath.Join(dstDir, "Default")

	for _, name := range credentialFiles {
		srcPath := findDBPath(srcDir, name)
		if srcPath == "" {
			continue
		}

		// Always write to Default/ in destination — Chromium reads from there
		dstPath := filepath.Join(dstDefault, name)

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0700); err != nil {
			return copied, fmt.Errorf("cannot create directory for %s: %w", name, err)
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return copied, fmt.Errorf("cannot copy %s: %w", name, err)
		}
		copied++

		// Remove stale side files in the target unless the matching source
		// side file exists and is copied too (only -journal is in the list).
		if isSidecar(name) {
			continue
		}
		srcBase := filepath.Join(filepath.Dir(srcPath), name)
		for _, suffix := range sqliteSidecarSuffixes {
			if _, err := os.Lstat(srcBase + suffix); err == nil && isCredentialFile(name+suffix) {
				continue
			}
			stale := dstPath + suffix
			if err := os.Remove(stale); err != nil && !os.IsNotExist(err) {
				return copied, fmt.Errorf("cannot remove stale %s: %w", stale, err)
			}
		}
	}

	// Encrypt any plaintext passwords in the copied Login Data
	loginDataPath := filepath.Join(dstDefault, "Login Data")
	if _, err := os.Stat(loginDataPath); err == nil {
		n, err := EncryptLoginDB(loginDataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot encrypt passwords: %v\n", err)
		} else if n > 0 {
			fmt.Fprintf(os.Stderr, "Encrypted %d plaintext password(s) for Chrome compatibility\n", n)
		}
	}

	return copied, nil
}

func isSidecar(name string) bool {
	for _, suffix := range sqliteSidecarSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func isCredentialFile(name string) bool {
	for _, f := range credentialFiles {
		if f == name {
			return true
		}
	}
	return false
}

// checkSameBrowser returns an error when both directories belong to known bpm
// profiles that use different browsers. Directories that aren't registered
// profiles are not checked.
func checkSameBrowser(srcDir, dstDir string) error {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	srcBrowser, dstBrowser := "", ""
	for _, p := range cfg.Profiles {
		switch filepath.Clean(p.DataDir) {
		case filepath.Clean(srcDir):
			srcBrowser = p.Browser
		case filepath.Clean(dstDir):
			dstBrowser = p.Browser
		}
	}
	if srcBrowser != "" && dstBrowser != "" && srcBrowser != dstBrowser {
		return fmt.Errorf("cannot sync credentials from a %s profile to a %s profile: each browser encrypts cookies and passwords with its own key, so the copied data would not decrypt in the target", srcBrowser, dstBrowser)
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic)
	dir := filepath.Dir(dst)
	tmpFile, err := os.CreateTemp(dir, "bpm-sync-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	if err := tmpFile.Chmod(srcInfo.Mode()); err != nil {
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

	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if runtime.GOOS != "windows" {
		if d, err := os.Open(dir); err == nil {
			_ = d.Sync()
			d.Close()
		}
	}
	return nil
}
