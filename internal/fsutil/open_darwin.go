//go:build darwin

package fsutil

import "os/exec"

// OpenPath opens a path or URL with the system default application.
func OpenPath(path string) error { return exec.Command("open", path).Start() }
