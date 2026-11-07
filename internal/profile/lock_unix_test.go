//go:build !windows

package profile

import (
	"fmt"
	"os"
	"syscall"
	"testing"
)

func TestProcessAliveResult(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		alive bool
	}{
		{"success", nil, true}, {"permission", syscall.EPERM, true}, {"wrapped permission", fmt.Errorf("signal: %w", syscall.EPERM), true}, {"missing", syscall.ESRCH, false}, {"finished", os.ErrProcessDone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := processAliveResult(tc.err); got != tc.alive {
				t.Fatalf("got %v", got)
			}
		})
	}
}
