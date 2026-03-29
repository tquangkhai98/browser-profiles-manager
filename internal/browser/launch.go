package browser

import (
	"fmt"
	"os/exec"
)

// Launch starts a browser with the given profile directory.
// It uses --user-data-dir for profile isolation.
// Returns the exec.Cmd so the caller can manage the process.
func Launch(browserPath, profileDir string) (*exec.Cmd, error) {
	args := []string{
		fmt.Sprintf("--user-data-dir=%s", profileDir),
		"--no-first-run",
		"--no-default-browser-check",
	}

	cmd := exec.Command(browserPath, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot launch browser: %w", err)
	}

	return cmd, nil
}
