package runtime

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
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

func TestWaitEndpointIgnoresForeignListener(t *testing.T) {
	// Another process already serves CDP on the requested port. Without this
	// browser's DevToolsActivePort the endpoint must not be adopted.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"ws://127.0.0.1/devtools/browser/foreign"}`))
	}))
	defer srv.Close()
	port := srv.Listener.Addr().(*net.TCPAddr).Port
	dir := t.TempDir()
	if _, ws, err := waitEndpoint(dir, port, 300*time.Millisecond); err == nil {
		t.Fatalf("adopted foreign endpoint %q", ws)
	}
	// A DevToolsActivePort for a different port than requested is not ours either.
	if err := os.WriteFile(filepath.Join(dir, "DevToolsActivePort"), []byte("1\n/devtools/browser/x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, ws, err := waitEndpoint(dir, port, 300*time.Millisecond); err == nil {
		t.Fatalf("adopted mismatched endpoint %q", ws)
	}
	if err := os.WriteFile(filepath.Join(dir, "DevToolsActivePort"), []byte(strconv.Itoa(port)+"\n/devtools/browser/own\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, ws, err := waitEndpoint(dir, port, time.Second); err != nil || got != port || ws == "" {
		t.Fatalf("port=%d ws=%q err=%v", got, ws, err)
	}
}

func TestStopRefusesWhileAnotherBPMProcessHoldsProfile(t *testing.T) {
	t.Setenv("BPM_CONFIG_DIR", t.TempDir())
	t.Setenv("BPM_DATA_DIR", t.TempDir())
	p, err := profile.Create("held", "chrome")
	if err != nil {
		t.Fatal(err)
	}
	// e.g. a launch that has not handed off yet, or `bpm use`: the lock PID
	// is a bpm process, which Stop must not signal (here: the test itself).
	release, err := profile.AcquireLock(p.DataDir, "launch")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	stopped, err := Stop("held", 0)
	if err == nil || stopped || !strings.Contains(err.Error(), "not by a detached browser") {
		t.Fatalf("Stop = %v, %v", stopped, err)
	}
	if locked, _ := profile.IsLocked(p.DataDir); !locked {
		t.Fatal("Stop must leave the lock in place")
	}
}

func TestStopWithoutBrowserReportsNotRunning(t *testing.T) {
	t.Setenv("BPM_CONFIG_DIR", t.TempDir())
	t.Setenv("BPM_DATA_DIR", t.TempDir())
	if _, err := profile.Create("idle", "chrome"); err != nil {
		t.Fatal(err)
	}
	stopped, err := Stop("idle", 0)
	if err != nil || stopped {
		t.Fatalf("Stop = %v, %v", stopped, err)
	}
}
