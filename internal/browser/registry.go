package browser

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

// BrowserInfo describes a supported Chromium-based browser.
type BrowserInfo struct {
	Name    string `json:"name"`
	ID      string `json:"id"`       // Short identifier: chrome, brave, edge, arc
	ExePath string `json:"exe_path"` // Full path to executable
}

// registry returns all known browser paths for the current OS.
func registry() []BrowserInfo {
	switch runtime.GOOS {
	case "darwin":
		return darwinBrowsers()
	case "windows":
		return windowsBrowsers()
	default:
		return linuxBrowsers()
	}
}

func darwinBrowsers() []BrowserInfo {
	return []BrowserInfo{
		{
			Name:    "Google Chrome",
			ID:      "chrome",
			ExePath: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		},
		{
			Name:    "Brave Browser",
			ID:      "brave",
			ExePath: "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		},
		{
			Name:    "Microsoft Edge",
			ID:      "edge",
			ExePath: "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		},
		{
			Name:    "Arc",
			ID:      "arc",
			ExePath: "/Applications/Arc.app/Contents/MacOS/Arc",
		},
	}
}

func windowsBrowsers() []BrowserInfo {
	programFiles := envOrDefault("ProgramFiles", `C:\Program Files`)
	programFilesX86 := envOrDefault("ProgramFiles(x86)", `C:\Program Files (x86)`)
	localAppData := envOrDefault("LOCALAPPDATA", "")

	browsers := []BrowserInfo{
		{
			Name:    "Google Chrome",
			ID:      "chrome",
			ExePath: programFiles + `\Google\Chrome\Application\chrome.exe`,
		},
		{
			Name:    "Brave Browser",
			ID:      "brave",
			ExePath: programFiles + `\BraveSoftware\Brave-Browser\Application\brave.exe`,
		},
		{
			Name:    "Microsoft Edge",
			ID:      "edge",
			ExePath: programFilesX86 + `\Microsoft\Edge\Application\msedge.exe`,
		},
	}

	// Chrome may also be in LocalAppData
	if localAppData != "" {
		browsers = append(browsers, BrowserInfo{
			Name:    "Google Chrome (User)",
			ID:      "chrome",
			ExePath: localAppData + `\Google\Chrome\Application\chrome.exe`,
		})
	}

	return browsers
}

func linuxBrowsers() []BrowserInfo {
	browsers := []BrowserInfo{
		{Name: "Google Chrome", ID: "chrome", ExePath: "/usr/bin/google-chrome"},
		{Name: "Google Chrome", ID: "chrome", ExePath: "/usr/bin/google-chrome-stable"},
		{Name: "Google Chrome", ID: "chrome", ExePath: "/usr/local/bin/google-chrome"},
		{Name: "Google Chrome", ID: "chrome", ExePath: "/opt/google/chrome/chrome"},
		{Name: "Google Chrome", ID: "chrome", ExePath: "/snap/bin/chromium"},
		{Name: "Chromium", ID: "chromium", ExePath: "/usr/bin/chromium"},
		{Name: "Chromium", ID: "chromium", ExePath: "/usr/bin/chromium-browser"},
		{Name: "Chromium", ID: "chromium", ExePath: "/usr/local/bin/chromium"},
		{Name: "Chromium", ID: "chromium", ExePath: "/snap/bin/chromium"},
		{Name: "Brave Browser", ID: "brave", ExePath: "/usr/bin/brave-browser"},
		{Name: "Brave Browser", ID: "brave", ExePath: "/usr/local/bin/brave-browser"},
		{Name: "Brave Browser", ID: "brave", ExePath: "/snap/bin/brave"},
		{Name: "Microsoft Edge", ID: "edge", ExePath: "/usr/bin/microsoft-edge"},
		{Name: "Microsoft Edge", ID: "edge", ExePath: "/usr/bin/microsoft-edge-stable"},
	}
	seen := make(map[string]bool)
	key := func(b BrowserInfo) string {
		p, err := filepath.EvalSymlinks(b.ExePath)
		if err != nil {
			p = b.ExePath
		}
		return b.ID + "\x00" + p
	}
	for _, b := range browsers {
		seen[key(b)] = true
	}
	for _, c := range []struct{ name, id, executable string }{
		{"Google Chrome", "chrome", "google-chrome"}, {"Google Chrome", "chrome", "google-chrome-stable"},
		{"Chromium", "chromium", "chromium"}, {"Chromium", "chromium", "chromium-browser"},
		{"Brave Browser", "brave", "brave-browser"}, {"Microsoft Edge", "edge", "microsoft-edge"},
	} {
		if path, err := exec.LookPath(c.executable); err == nil {
			b := BrowserInfo{Name: c.name, ID: c.id, ExePath: path}
			if !seen[key(b)] {
				browsers = append(browsers, b)
				seen[key(b)] = true
			}
		}
	}
	return browsers
}
