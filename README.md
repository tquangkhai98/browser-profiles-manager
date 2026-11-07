# bpm — Browser Profiles Manager

[![Release](https://img.shields.io/github/v/release/tquangkhai98/browser-profiles-manager?color=blue)](https://github.com/tquangkhai98/browser-profiles-manager/releases/latest)
[![Build](https://github.com/tquangkhai98/browser-profiles-manager/actions/workflows/release.yml/badge.svg)](https://github.com/tquangkhai98/browser-profiles-manager/actions/workflows/release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/tquangkhai98/browser-profiles-manager)](https://goreportcard.com/report/github.com/tquangkhai98/browser-profiles-manager)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)]()

> Centralized browser profile management for AI IDEs

**bpm** is a CLI tool + MCP server that manages isolated Chromium browser profiles across AI-powered development environments like Claude Code, Cursor, and Antigravity.

---

## Why bpm?

AI IDEs spawn their own Chromium instances with ephemeral profiles:

| Pain Point | Impact |
|-----------|--------|
| 🔑 Login sessions lost between runs | Re-authenticate repeatedly (OAuth, 2FA) |
| 💥 Profile conflicts across parallel agents | Browser crashes, corrupt data |
| 🔄 No credential sync across tools | Each tool starts from scratch |
| 📁 Profiles scattered across filesystem | Hard to audit, clean up, back up |

**bpm is a profile provider for AI agents and testing** — real Chrome, no fingerprint spoofing.
It composes with Playwright MCP / Chrome DevTools MCP rather than driving the browser itself.
Anti-detect managers like [FoxProfile](https://github.com/chienbm98/foxprofile) exist for spoofed-fingerprint/proxy use cases;
Chrome DevTools MCP's `--isolated` option covers throwaway profiles.

---

## Features

| Feature | Description |
|---------|-------------|
| **Profile CRUD** | Create, list, delete isolated browser profiles |
| **Browser Launch** | Launch Chromium with `--user-data-dir` + file lock |
| **Directory Mapping** | Map project dirs to profiles for auto-resolution |
| **Credential Sync** | Inspect & sync cookies/logins between profiles |
| **Import/Export** | Import existing Chrome profiles, export for backup |
| **MCP Server** | Expose all features to any AI IDE via MCP protocol |
| **Desktop App** | Simple GUI for visual management (Wails) |

---

## Quick Start

```bash
# Install (requires Go 1.24+)
go install github.com/tquangkhai98/browser-profiles-manager@latest

# Or download a pre-built binary from GitHub Releases:
# https://github.com/tquangkhai98/browser-profiles-manager/releases

# Create a profile
bpm create work-staging

# Launch browser with profile
bpm use work-staging

# Check what credentials exist
bpm creds work-staging

# Sync credentials to another profile
bpm sync work-staging personal

# Import existing Chrome profile
bpm import ~/Library/Application\ Support/Google/Chrome/Default my-chrome

# Map project directory to profile
bpm map ~/projects/my-app work-staging
```

---

## Linux (headless / server)

Download `bpm-<version>-linux-amd64.tar.gz` or `bpm-<version>-linux-arm64.tar.gz` from [GitHub Releases](https://github.com/tquangkhai98/browser-profiles-manager/releases), extract it and put `bpm` on your PATH, or install with `go install github.com/tquangkhai98/browser-profiles-manager@latest`. Install Chromium or Chrome on the host.

```bash
bpm create work
bpm launch work --headless --json
```

The launch returns one JSON object, for example:

```json
{"profile":"work","browser":"chrome","pid":1234,"port":41234,"cdp_http":"http://127.0.0.1:41234","webSocketDebuggerUrl":"ws://127.0.0.1:41234/devtools/browser/example","headless":true}
```

Connect from Playwright on the same host:

```javascript
const browser = await chromium.connectOverCDP("http://127.0.0.1:<port>");
```

Stop the browser with `bpm stop work`. CDP listens on loopback; remote clients need a secure tunnel. The CLI and MCP server are one CGO-free binary. `profile_use` starts the same detached browser and returns its CDP state; `profile_list` reports running state and `profile_stop` stops it.

Only one live browser can use a profile. `bpm launch work --wait 30s` waits for its lock; stale locks are cleaned automatically. Set `BPM_CONFIG_DIR` to relocate the configuration, run state, and logs. Headless mode is selected automatically when no display is available, or with `BPM_HEADLESS=1`. Use `--headless=false` to override it and `--cdp-port 9222` to select a port. When running as root, bpm adds `--no-sandbox` to headless Chromium.

On Linux every browser launch uses `--password-store=basic` so profiles can move between machines without depending on a host keyring. Cookies are encrypted with a fixed key, so protect the profile directory from unauthorized access.

---

## Move a profile between machines

Stop the browser before exporting, then copy the archive to the destination machine:

```bash
bpm stop work
bpm export work ~/work.tar.gz
# Copy ~/work.tar.gz to the destination machine.
bpm import ~/work.tar.gz --name work
bpm launch work --headless --json
```

Linux launches always pass `--password-store=basic`, so cookies created with that setting are not sealed to the host keyring. Cookies created with another operating system or keyring may not be portable. The archive contains session cookies; treat it as a secret. A copied Chrome user-data directory, such as `/config/chrome` from a linuxserver/chrome container using the basic password store, can also be imported directly.

## MCP Integration

Add to your AI IDE's MCP config (Claude Code, Cursor, Antigravity):

```json
{
  "mcpServers": {
    "bpm": {
      "command": "bpm",
      "args": ["serve"]
    }
  }
}
```

### MCP Tools

| Tool | Description |
|------|-------------|
| `profile_create` | Create a new isolated browser profile |
| `profile_list` | List all profiles with status |
| `profile_use` | Launch browser with profile |
| `profile_status` | Check lock status |
| `profile_delete` | Delete a profile |
| `mapping_set` | Map directory to profile |
| `mapping_get` | Resolve profile for a directory |
| `creds_inspect` | List credentials in a profile |
| `creds_sync` | Sync credentials between profiles |
| `browser_detect` | List installed browsers |

---

## AI Agent Integration

bpm profiles work with browser automation MCP servers, letting AI agents control browsers **with persistent login sessions**. Login once manually, then let your AI agent automate with full authentication — no re-login needed.

### Use Case: Multi-Role Testing

```bash
# Step 1: Create profiles for each role
bpm create lms-admin
bpm create lms-teacher
bpm create lms-student

# Step 2: Login once per profile (manual)
bpm use lms-admin      # → login as admin
bpm use lms-teacher    # → login as teacher
bpm use lms-student    # → login as student

# Step 3: AI agent uses profiles with persistent sessions ✅
```

### 🥇 Playwright MCP — Best Overall

Stable, official, uses accessibility tree for reliable element interaction.

```json
{
  "mcpServers": {
    "browser-admin": {
      "command": "npx",
      "args": [
        "@playwright/mcp@latest",
        "--user-data-dir", "~/.local/share/bpm/profiles/lms-admin"
      ]
    },
    "browser-teacher": {
      "command": "npx",
      "args": [
        "@playwright/mcp@latest",
        "--user-data-dir", "~/.local/share/bpm/profiles/lms-teacher"
      ]
    }
  }
}
```

### 🥈 Chrome DevTools MCP — Debug & Inspect

Connect to a running Chrome instance launched by bpm. Provides DOM inspection, Network monitoring, and Console access.

First launch Chrome manually with remote debugging enabled:

```bash
# macOS example — launch Chrome with remote debugging on port 9222
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --user-data-dir="$(bpm list --json | jq -r '.[] | select(.name=="lms-admin") | .path')" \
  --remote-debugging-port=9222
```

```json
{
  "mcpServers": {
    "chrome-devtools": {
      "command": "npx",
      "args": ["chrome-devtools-mcp@latest", "--browserUrl=http://127.0.0.1:9222"]
    }
  }
}
```

### 🥉 Browser Use MCP — Natural Language

AI-native browser automation — describe actions in plain language instead of CSS selectors.

```json
{
  "mcpServers": {
    "browser-use": {
      "command": "npx",
      "args": [
        "browser-use-mcp@latest",
        "--user-data-dir", "~/.local/share/bpm/profiles/lms-admin"
      ]
    }
  }
}
```

### Comparison

| Tool | Session Persist | AI Control | Best For |
|------|:-:|:-:|----------|
| **Playwright MCP** | ✅ | ✅ | E2E testing, form automation |
| **Chrome DevTools MCP** | ✅ | ✅ | Debugging, network inspection |
| **Browser Use MCP** | ✅ | ✅ | Natural language automation |
| bpm CLI only | ✅ | ❌ | Manual testing |

> **⚠️ Important:** Never open the same profile in two browsers simultaneously — bpm uses file locks to prevent profile corruption.

---

## CLI Commands

```
bpm create <name>          Create isolated profile
bpm list [--json]          List profiles with status
bpm delete <name>          Delete profile + data
bpm status <name> [--json] Check lock/usage status
bpm use <name>             Launch browser with profile
bpm detect [--json]        List installed browsers
bpm map <dir> <profile>    Map project dir → profile
bpm map --auto [--json]    Auto-resolve profile for cwd
bpm map --list [--json]    Show all mappings
bpm creds <name> [--json]  Inspect credentials in profile
bpm sync <src> <dst>       Sync credentials between profiles
bpm import <path> [name]  Import a Chrome profile or archive
bpm export <name> <path>   Export a profile archive
bpm serve                  Start MCP server
```

### JSON Output

All read commands support `--json` for scripting and piping:

```bash
$ bpm list --json
[
  {
    "name": "work-staging",
    "browser": "chrome",
    "locked": false,
    "created_at": "2026-03-29T12:00:00Z"
  }
]

$ bpm detect --json
[
  {
    "name": "Google Chrome",
    "id": "chrome",
    "exe_path": "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
  }
]

$ bpm creds work-staging --json
{
  "profile_name": "work-staging",
  "sites": [
    {"domain": "github.com", "cookie_count": 12, "login_count": 1}
  ],
  "total_cookies": 12,
  "total_logins": 1
}
```

---

## How Credential Sync Works

When you login to GitHub in one browser profile, that login is stored as cookies and saved passwords. **Credential sync** copies those databases to another profile:

1. Reads Chromium SQLite databases (`Cookies`, `Login Data`) from source
2. Backs up target profile's databases
3. Copies databases to target via atomic write (temp file → rename)
4. ⚠️ **bpm never decrypts passwords** — they work because Chromium uses OS-level keychain (macOS Keychain / Windows DPAPI)

---

## Screenshots

<p align="center">
  <img src="docs/screenshots/01-profile-list.png" width="800" alt="Profile List" />
  <br><em>Profile List — manage all browser profiles</em>
</p>

<p align="center">
  <img src="docs/screenshots/02-ai-integration.png" width="800" alt="AI Agent Integration" />
  <br><em>AI Agent Integration — connect profiles to MCP automation tools</em>
</p>

<p align="center">
  <img src="docs/screenshots/03-settings.png" width="800" alt="Settings" />
  <br><em>Settings — configure defaults, MCP server, and one-click IDE install</em>
</p>

---

## Development

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | ≥ 1.24 | `brew install go` |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Node.js | ≥ 18 | `brew install node` (for desktop frontend) |

### Build & Test

```bash
make build          # Build CLI with version info
make test           # Run all tests
make cover          # Generate coverage report
make lint           # Run staticcheck
make install        # Install to $GOPATH/bin
make desktop        # Build desktop app
make dev            # Desktop dev mode with hot-reload
make clean          # Remove build artifacts
```

### Project Structure

```
browser-profiles-manager/
├── main.go                 # CLI entry point
├── cmd/                    # CLI commands (Cobra)
├── internal/               # Core business logic
│   ├── profile/            #   Profile CRUD + locking
│   ├── browser/            #   Browser detection & launch
│   ├── credential/         #   Credential read/sync (SQLite)
│   ├── mapping/            #   Directory ↔ profile mapping
│   ├── config/             #   Configuration management
│   └── mcp/                #   MCP server implementation
├── desktop/                # Wails desktop app
├── Makefile                # Build automation
└── docs/                   # Documentation
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------:|
| Core + CLI | Go + Cobra |
| MCP Server | mcp-go (stdio) |
| Desktop App | Wails v2 |
| Storage | JSON config + filesystem |
| Credential read | modernc.org/sqlite (pure Go) |

## Platform Support

| Platform | Status |
|----------|--------|
| macOS | ✅ Supported |
| Windows | ✅ Supported |
| Linux | CLI and headless MCP server (amd64/arm64); desktop (amd64) |

---

## License

MIT

---

<p align="center">
  If you find bpm useful, consider giving it a ⭐ — it helps others discover this project!
</p>
