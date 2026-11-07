//go:build !windows

package runtime

import (
	"os/exec"
	"syscall"
)

func detach(cmd *exec.Cmd)        { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }
func signalProcess(pid int) error { return syscall.Kill(pid, syscall.SIGTERM) }
func signalGroup(pid int) error   { return syscall.Kill(-pid, syscall.SIGKILL) }
