package credential

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// credentialFiles are the Chromium database files to sync between profiles.
var credentialFiles = []string{
	"Cookies",
	"Cookies-journal",
	"Login Data",
	"Login Data-journal",
}

// Sync copies credential database files from source to target profile.
// It copies the files as-is, then encrypts any plaintext passwords
// so Chrome can read them (Chrome macOS requires v10-encrypted format).
// Files are always placed in the target's Default/ subdirectory, since Chromium
// with --user-data-dir reads from <profileDir>/Default/.
func Sync(srcDir, dstDir string) (int, error) {
	copied := 0

	for _, name := range credentialFiles {
		srcPath := findDBPath(srcDir, name)
		if srcPath == "" {
			continue
		}

		// Always write to Default/ in destination — Chromium reads from there
		dstPath := filepath.Join(dstDir, "Default", name)

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0700); err != nil {
			return copied, fmt.Errorf("cannot create directory for %s: %w", name, err)
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return copied, fmt.Errorf("cannot copy %s: %w", name, err)
		}
		copied++
	}

	// Encrypt any plaintext passwords in the copied Login Data
	loginDataPath := filepath.Join(dstDir, "Default", "Login Data")
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

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}
