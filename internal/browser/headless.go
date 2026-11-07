package browser

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HasDisplay reports whether a GUI display is available.
func HasDisplay() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "windows" || os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

// ResolveHeadless resolves the explicit flag, environment, and display defaults.
func ResolveHeadless(flagSet, flagValue bool, envValue string, hasDisplay bool) bool {
	if flagSet {
		return flagValue
	}
	switch strings.ToLower(envValue) {
	case "1", "true", "yes", "on":
		return true
	}
	return !hasDisplay
}

// ResolveCDPEndpoint returns the CDP endpoint for a headless launch.
// An explicit port is checked over loopback; port zero uses DevToolsActivePort.
func ResolveCDPEndpoint(profileDir string, port int, timeout time.Duration) (string, error) {
	if port < 0 || port > 65535 {
		return "", fmt.Errorf("invalid CDP port %d", port)
	}
	if port == 0 {
		return WaitCDPEndpoint(profileDir, timeout)
	}

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
		if err == nil {
			conn.Close()
			return "http://" + address, nil
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", fmt.Errorf("waiting for CDP endpoint on port %d: %w", port, ctx.Err())
		case <-timer.C:
		}
	}
}

// WaitCDPEndpoint waits for Chromium to publish its debugging port.
func WaitCDPEndpoint(profileDir string, timeout time.Duration) (string, error) {
	path := filepath.Join(profileDir, "DevToolsActivePort")
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			line := strings.SplitN(string(data), "\n", 2)[0]
			port, parseErr := strconv.Atoi(strings.TrimSpace(line))
			if parseErr == nil && port > 0 && port <= 65535 {
				return fmt.Sprintf("http://127.0.0.1:%d", port), nil
			}
			err = fmt.Errorf("invalid debugging port %q", line)
		}
		lastErr = err
		if time.Now().After(deadline) {
			return "", fmt.Errorf("waiting for CDP endpoint: %w", lastErr)
		}
		time.Sleep(min(100*time.Millisecond, time.Until(deadline)))
	}
}
