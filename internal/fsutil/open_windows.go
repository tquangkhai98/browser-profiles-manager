//go:build windows

package fsutil

import "os/exec"

// OpenPath opens a path or URL with the system default application.
func OpenPath(path string) error { return exec.Command("cmd", "/c", "start", "", path).Start() }
