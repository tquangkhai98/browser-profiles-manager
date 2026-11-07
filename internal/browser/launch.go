package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// LaunchOptions controls optional Chromium launch arguments.
type LaunchOptions struct {
	Headless  bool
	CDPPort   int
	ExtraArgs []string
}

var isRoot = func() bool { return os.Geteuid() == 0 }

// BuildArgs builds Chromium arguments for a profile.
func BuildArgs(profileDir string, opts LaunchOptions) []string {
	args := []string{fmt.Sprintf("--user-data-dir=%s", profileDir), "--no-first-run", "--no-default-browser-check"}
	if runtime.GOOS == "linux" {
		args = append(args, "--password-store=basic")
	}
	args = append(args, opts.ExtraArgs...)
	if opts.Headless {
		args = append(args, "--headless=new", "--disable-gpu", "--hide-scrollbars", "--mute-audio", fmt.Sprintf("--remote-debugging-port=%d", opts.CDPPort))
		if isRoot() {
			args = append(args, "--no-sandbox")
		}
	}
	return args
}

// Launch starts a browser with the given profile directory and returns its command.
func Launch(browserPath, profileDir string, opts LaunchOptions) (*exec.Cmd, error) {
	cmd := exec.Command(browserPath, BuildArgs(profileDir, opts)...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot launch browser: %w", err)
	}
	return cmd, nil
}
