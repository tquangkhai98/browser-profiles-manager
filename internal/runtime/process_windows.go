//go:build windows

package runtime

import (
	"os"
	"os/exec"
)

func detach(cmd *exec.Cmd)        {}
func signalProcess(pid int) error { return signalGroup(pid) }
func signalGroup(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
