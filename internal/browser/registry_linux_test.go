package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxBrowserCandidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chromium-browser")
	if err := os.WriteFile(path, []byte(""), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	candidates := linuxBrowsers()
	expected := []struct{ id, path string }{
		{"chrome", "/usr/bin/google-chrome"}, {"chrome", "/usr/bin/google-chrome-stable"},
		{"chrome", "/usr/local/bin/google-chrome"}, {"chrome", "/opt/google/chrome/chrome"},
		{"chrome", "/snap/bin/chromium"}, {"chromium", "/usr/bin/chromium"},
		{"chromium", "/usr/bin/chromium-browser"}, {"chromium", "/usr/local/bin/chromium"},
		{"chromium", "/snap/bin/chromium"}, {"brave", "/usr/bin/brave-browser"},
		{"brave", "/usr/local/bin/brave-browser"}, {"brave", "/snap/bin/brave"},
		{"edge", "/usr/bin/microsoft-edge"}, {"edge", "/usr/bin/microsoft-edge-stable"},
		{"chromium", path},
	}
	if len(candidates) != len(expected) {
		t.Fatalf("got %d candidates, want %d", len(candidates), len(expected))
	}
	for i, want := range expected {
		if candidates[i].ID != want.id || candidates[i].ExePath != want.path {
			t.Errorf("candidate %d: %+v, want %+v", i, candidates[i], want)
		}
	}
}
