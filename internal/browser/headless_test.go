package browser

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestResolveHeadless(t *testing.T) {
	for _, tc := range []struct {
		name          string
		set, value    bool
		env           string
		display, want bool
	}{
		{"explicit false", true, false, "true", false, false},
		{"explicit true", true, true, "", true, true},
		{"environment", false, false, "YeS", true, true},
		{"no display", false, false, "false", false, true},
		{"display", false, false, "off", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveHeadless(tc.set, tc.value, tc.env, tc.display); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveCDPEndpoint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) {
			t.Skip("sandbox disallows TCP sockets")
		}
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	endpoint, err := ResolveCDPEndpoint(t.TempDir(), port, time.Second)
	if err != nil || endpoint != fmt.Sprintf("http://127.0.0.1:%d", port) {
		t.Fatalf("endpoint %q: %v", endpoint, err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveCDPEndpoint(t.TempDir(), port, 100*time.Millisecond); err == nil {
		t.Fatal("expected timeout for unused port")
	}
	for _, port := range []int{-1, 70000} {
		if _, err := ResolveCDPEndpoint(t.TempDir(), port, time.Second); err == nil {
			t.Errorf("expected error for invalid port %d", port)
		}
	}
}

func TestBuildArgs(t *testing.T) {
	old := isRoot
	defer func() { isRoot = old }()
	isRoot = func() bool { return false }
	base := []string{"--user-data-dir=/profile", "--no-first-run", "--no-default-browser-check"}
	if runtime.GOOS == "linux" {
		base = append(base, "--password-store=basic")
	}
	if got := BuildArgs("/profile", LaunchOptions{}); !reflect.DeepEqual(got, base) {
		t.Fatalf("base: %v", got)
	}
	want := append(append([]string{}, base...), "--extra", "--headless=new", "--disable-gpu", "--hide-scrollbars", "--mute-audio", "--remote-debugging-port=9222")
	if got := BuildArgs("/profile", LaunchOptions{Headless: true, CDPPort: 9222, ExtraArgs: []string{"--extra"}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("headless: %v", got)
	}
	isRoot = func() bool { return true }
	got := BuildArgs("/profile", LaunchOptions{Headless: true})
	if got[len(got)-2] != "--remote-debugging-port=0" || got[len(got)-1] != "--no-sandbox" {
		t.Fatalf("root: %v", got)
	}
}

func TestWaitCDPEndpoint(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "DevToolsActivePort"), []byte("9222\n/devtools/browser/id\n"), 0600); err != nil {
		t.Fatal(err)
	}
	endpoint, err := WaitCDPEndpoint(dir, time.Second)
	if err != nil || endpoint != "http://127.0.0.1:9222" {
		t.Fatalf("endpoint %q: %v", endpoint, err)
	}
	if _, err := WaitCDPEndpoint(t.TempDir(), 0); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestBuildArgsModes(t *testing.T) {
	old := isRoot
	defer func() { isRoot = old }()
	for _, tc := range []struct {
		name           string
		headless, root bool
	}{
		{"headed user", false, false}, {"headed root", false, true}, {"headless user", true, false}, {"headless root", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isRoot = func() bool { return tc.root }
			args := BuildArgs("/profile", LaunchOptions{Headless: tc.headless, CDPPort: 9222})
			has := func(want string) bool {
				for _, arg := range args {
					if arg == want {
						return true
					}
				}
				return false
			}
			for flag, want := range map[string]bool{"--password-store=basic": runtime.GOOS == "linux", "--headless=new": tc.headless, "--remote-debugging-port=9222": tc.headless, "--no-sandbox": tc.headless && tc.root} {
				if has(flag) != want {
					t.Errorf("flag %s: got %v, want %v", flag, has(flag), want)
				}
			}
		})
	}
}
