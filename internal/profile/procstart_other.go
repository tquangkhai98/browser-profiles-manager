//go:build !linux

package profile

// processStartTime is not implemented on this platform; PID reuse is not
// detected and liveness falls back to the PID check alone.
func processStartTime(int) (uint64, bool) { return 0, false }

func processZombie(int) bool { return false }
