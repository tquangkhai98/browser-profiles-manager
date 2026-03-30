package browser

import (
	"testing"
)

func TestDetectBrowsers_ReturnsSlice(t *testing.T) {
	browsers := DetectBrowsers()

	// Should return a non-nil slice (even if empty on CI)
	if browsers == nil {
		t.Error("DetectBrowsers() returned nil, expected non-nil slice")
	}
}

func TestDetectBrowsers_NoDuplicates(t *testing.T) {
	browsers := DetectBrowsers()

	seen := make(map[string]bool)
	for _, b := range browsers {
		if seen[b.ID] {
			t.Errorf("duplicate browser ID: %q", b.ID)
		}
		seen[b.ID] = true
	}
}

func TestFindBrowser_NotFound(t *testing.T) {
	_, err := FindBrowser("nonexistent-browser-xyz")
	if err == nil {
		t.Fatal("expected error for unknown browser, got nil")
	}
}

func TestDefaultBrowser(t *testing.T) {
	b, err := DefaultBrowser()

	// On CI/containers this might fail (no browsers installed)
	// On dev machines it should succeed
	if err != nil {
		t.Skipf("no browser found (expected on CI): %v", err)
	}

	if b.Name == "" {
		t.Error("DefaultBrowser().Name is empty")
	}
	if b.ID == "" {
		t.Error("DefaultBrowser().ID is empty")
	}
	if b.ExePath == "" {
		t.Error("DefaultBrowser().ExePath is empty")
	}
}

func TestRegistry_ReturnsEntries(t *testing.T) {
	entries := registry()
	if len(entries) == 0 {
		t.Error("registry() returned no entries")
	}

	for _, e := range entries {
		if e.Name == "" {
			t.Error("browser entry has empty Name")
		}
		if e.ID == "" {
			t.Error("browser entry has empty ID")
		}
		if e.ExePath == "" {
			t.Error("browser entry has empty ExePath")
		}
	}
}
