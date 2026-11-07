// Package runtime manages detached browser processes and their CDP state.
package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

type State struct {
	Profile   string    `json:"profile"`
	Browser   string    `json:"browser"`
	PID       int       `json:"pid"`
	Port      int       `json:"port"`
	CDPHTTP   string    `json:"cdp_http"`
	WebSocket string    `json:"webSocketDebuggerUrl"`
	Headless  bool      `json:"headless"`
	StartedAt time.Time `json:"started_at,omitempty"`
}

type Options struct {
	Browser  string
	Headless bool
	Port     int
	Wait     time.Duration
}

func statePath(name string) (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "run", name+".json"), nil
}
func Read(name string) (*State, error) {
	path, err := statePath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err = json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if !profile.ProcessAlive(s.PID) {
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	return &s, nil
}
func Launch(name string, opts Options) (*State, error) {
	if opts.Port < 0 || opts.Port > 65535 || opts.Wait < 0 {
		return nil, fmt.Errorf("invalid port or wait duration")
	}
	p, err := profile.Get(name)
	if err != nil {
		return nil, err
	}
	id := opts.Browser
	if id == "" {
		id = p.Browser
	}
	if id == "" {
		b, e := browser.DefaultBrowser()
		if e != nil {
			return nil, e
		}
		id = b.ID
	}
	b, err := browser.FindBrowser(id)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(opts.Wait)
	var release func()
	var handoff func(int) error
	for {
		release, handoff, err = profile.AcquireForHandoff(p.DataDir, "launch")
		if err == nil {
			break
		}
		if opts.Wait == 0 || !time.Now().Before(deadline) {
			return nil, fmt.Errorf("profile %q: %w", name, err)
		}
		time.Sleep(min(100*time.Millisecond, time.Until(deadline)))
	}
	held := true
	defer func() {
		if held {
			release()
		}
	}()
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	logdir := filepath.Join(dir, "logs")
	if err = os.MkdirAll(logdir, 0700); err != nil {
		return nil, err
	}
	log, err := os.OpenFile(filepath.Join(logdir, name+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	if err = os.Remove(filepath.Join(p.DataDir, "DevToolsActivePort")); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	args := browser.BuildArgs(p.DataDir, browser.LaunchOptions{Headless: opts.Headless, CDPPort: opts.Port})
	if !opts.Headless {
		args = append(args, fmt.Sprintf("--remote-debugging-port=%d", opts.Port))
	}
	cmd := exec.Command(b.ExePath, args...)
	cmd.Stdout = log
	cmd.Stderr = log
	detach(cmd)
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot launch browser: %w", err)
	}
	pid := cmd.Process.Pid
	failed := true
	defer func() {
		if failed {
			_ = signalGroup(pid)
			_ = cmd.Wait()
		}
	}()
	port, ws, err := waitEndpoint(p.DataDir, opts.Port, 10*time.Second)
	if err != nil {
		return nil, err
	}
	s := &State{Profile: name, Browser: id, PID: pid, Port: port, CDPHTTP: fmt.Sprintf("http://127.0.0.1:%d", port), WebSocket: ws, Headless: opts.Headless, StartedAt: time.Now().UTC()}
	path, err := statePath(name)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		return nil, err
	}
	if err = handoff(pid); err != nil {
		os.Remove(path)
		return nil, err
	}
	held = false
	if err = cmd.Process.Release(); err != nil {
		return nil, err
	}
	failed = false
	_ = profile.UpdateLastUsed(name)
	return s, nil
}
func waitEndpoint(dir string, requested int, timeout time.Duration) (int, string, error) {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(filepath.Join(dir, "DevToolsActivePort"))
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		port := requested
		path := ""
		if len(lines) >= 2 {
			if n, e := strconv.Atoi(strings.TrimSpace(lines[0])); e == nil && n > 0 && n <= 65535 {
				port = n
				path = strings.TrimSpace(lines[1])
			}
		}
		if port > 0 {
			url := fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
			resp, e := client.Get(url)
			if e == nil {
				var v struct {
					WebSocket string `json:"webSocketDebuggerUrl"`
				}
				e = json.NewDecoder(resp.Body).Decode(&v)
				resp.Body.Close()
				if e == nil && v.WebSocket != "" {
					return port, v.WebSocket, nil
				}
			}
			if path != "" {
				return port, fmt.Sprintf("ws://127.0.0.1:%d%s", port, path), nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return 0, "", fmt.Errorf("waiting for DevTools endpoint timed out")
}
func escalation(elapsed, timeout time.Duration) bool { return elapsed >= timeout }

func Stop(name string, timeout time.Duration) (bool, error) {
	if timeout < 0 {
		return false, fmt.Errorf("invalid timeout")
	}
	p, err := profile.Get(name)
	if err != nil {
		return false, err
	}
	s, err := Read(name)
	if err != nil {
		return false, err
	}
	pid := 0
	if s != nil {
		pid = s.PID
	} else if locked, info := profile.IsLocked(p.DataDir); locked {
		pid = info.PID
	}
	if pid == 0 {
		_ = profile.ReleaseLock(p.DataDir)
		return false, nil
	}
	// A group-wide SIGTERM prevents Chromium's graceful shutdown and can lose
	// unflushed cookies and session state. Signal only the browser first; use
	// SIGKILL on its group only if it remains alive after the timeout.
	if err = signalProcess(pid); err != nil && profile.ProcessAlive(pid) {
		return false, fmt.Errorf("cannot stop browser: %w", err)
	}
	start := time.Now()
	for profile.ProcessAlive(pid) && !escalation(time.Since(start), timeout) {
		time.Sleep(min(100*time.Millisecond, timeout-time.Since(start)))
	}
	if profile.ProcessAlive(pid) {
		if err = signalGroup(pid); err != nil && profile.ProcessAlive(pid) {
			return false, err
		}
	}
	path, err := statePath(name)
	if err != nil {
		return false, err
	}
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err = profile.ReleaseLock(p.DataDir); err != nil {
		return false, err
	}
	return true, nil
}
