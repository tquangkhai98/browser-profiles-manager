//go:build linux

package profile

import (
	"bytes"
	"os"
	"strconv"
)

// processStartTime returns the start time of pid in clock ticks since boot
// (field 22 of /proc/<pid>/stat). It lets a lock or runtime record tell its
// own process apart from an unrelated process that later reused the PID.
func processStartTime(pid int) (uint64, bool) {
	start, _, ok := processStat(pid)
	return start, ok
}

// processZombie reports whether pid has exited but has not been reaped yet.
// kill(pid, 0) still succeeds for a zombie, so it must not count as alive.
func processZombie(pid int) bool {
	_, zombie, ok := processStat(pid)
	return ok && zombie
}

func processStat(pid int) (start uint64, zombie, ok bool) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0, false, false
	}
	// comm (field 2) may contain spaces and parentheses; fields after the last
	// ')' start at field 3 (state).
	i := bytes.LastIndexByte(data, ')')
	if i < 0 {
		return 0, false, false
	}
	fields := bytes.Fields(data[i+1:])
	const startTimeIndex = 22 - 3
	if len(fields) <= startTimeIndex {
		return 0, false, false
	}
	v, err := strconv.ParseUint(string(fields[startTimeIndex]), 10, 64)
	if err != nil {
		return 0, false, false
	}
	return v, string(fields[0]) == "Z", true
}
