package browser

import (
	"fmt"
	"os"
)

// DetectBrowsers scans for installed Chromium-based browsers on the system.
func DetectBrowsers() []BrowserInfo {
	var found []BrowserInfo
	seen := make(map[string]bool)

	for _, b := range registry() {
		if _, err := os.Stat(b.ExePath); err == nil {
			// Deduplicate by ID — keep first found
			if !seen[b.ID] {
				seen[b.ID] = true
				found = append(found, b)
			}
		}
	}
	return found
}

// FindBrowser finds a specific browser by its short ID (e.g., "chrome", "brave").
func FindBrowser(id string) (*BrowserInfo, error) {
	for _, b := range DetectBrowsers() {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("browser %q not found. Run 'bpm detect' to see installed browsers", id)
}

// DefaultBrowser returns the first detected browser, or an error if none found.
func DefaultBrowser() (*BrowserInfo, error) {
	browsers := DetectBrowsers()
	if len(browsers) == 0 {
		return nil, fmt.Errorf("no supported Chromium browser found. Install Chrome, Brave, or Edge")
	}
	return &browsers[0], nil
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
