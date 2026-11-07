package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEscalation(t *testing.T) {
	for _, tc := range []struct {
		name             string
		elapsed, timeout time.Duration
		want             bool
	}{
		{"before timeout", time.Second, 2 * time.Second, false},
		{"at timeout", 2 * time.Second, 2 * time.Second, true},
		{"after timeout", 3 * time.Second, 2 * time.Second, true},
		{"zero timeout", 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := escalation(tc.elapsed, tc.timeout); got != tc.want {
				t.Errorf("escalation(%v, %v) = %v, want %v", tc.elapsed, tc.timeout, got, tc.want)
			}
		})
	}
}

func TestStopRejectsNegativeTimeout(t *testing.T) {
	if stopped, err := Stop("", -time.Second); err == nil || stopped {
		t.Fatalf("Stop with negative timeout = %v, %v; want false and an error", stopped, err)
	}
}

func TestWaitEndpointFallsBackToActivePort(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "DevToolsActivePort"), []byte("43210\n/devtools/browser/example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	port, ws, err := waitEndpoint(dir, 0, time.Second)
	if err != nil || port != 43210 || ws != "ws://127.0.0.1:43210/devtools/browser/example" {
		t.Fatalf("port=%d ws=%q error=%v", port, ws, err)
	}
}
