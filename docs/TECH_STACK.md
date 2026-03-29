# Tech Stack — Browser Profiles Manager (bpm)

## Overview

bpm is built as a **single Go binary** that serves as both a CLI tool and MCP server, with an optional desktop GUI built using Wails. The stack prioritizes simplicity, cross-platform support, and zero external runtime dependencies.

---

## Core

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | [Go](https://go.dev/) | 1.22+ | Single binary, cross-compilation, syscall access |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) | v1.8+ | Industry-standard Go CLI framework with subcommands, flags, help generation |
| MCP SDK | [mcp-go](https://github.com/mark3labs/mcp-go) | latest | Official Go SDK for Model Context Protocol (stdio transport) |

### Why Go?
- **Single binary** — no runtime, no dependencies, just copy and run
- **Cross-compilation** — `GOOS=darwin GOARCH=arm64 go build` for macOS, `GOOS=windows` for Windows
- **Syscall access** — native file locking via `syscall.Flock` (POSIX) and `LockFileEx` (Windows)
- **Fast startup** — CLI commands respond in < 200ms
- **Strong concurrency** — goroutines for background tasks, channels for communication

---

## Storage

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Config | JSON file | `config.json` — profiles, mappings, settings |
| Profile data | Filesystem | Each profile = isolated directory under `profiles/<name>/` |
| Credential read | [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure Go SQLite — read Chromium cookie/login databases |

### Storage Paths

| Platform | Config | Profile Data |
|----------|--------|-------------|
| macOS | `~/.config/bpm/config.json` | `~/.local/share/bpm/profiles/<name>/` |
| Windows | `%APPDATA%\bpm\config.json` | `%LOCALAPPDATA%\bpm\profiles\<name>\` |

### Why JSON (not SQLite/YAML)?
- Human-readable and editable
- No driver dependency for config
- Atomic writes via temp file + `os.Rename`
- Good enough for <100 profiles (typical use case)

### Why modernc.org/sqlite?
- **Pure Go** — no CGO, no system SQLite dependency
- **Read-only** — bpm only reads Chromium's `Cookies` and `Login Data` SQLite DBs
- **Cross-platform** — works on macOS and Windows without native libs

---

## Concurrency & Locking

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Profile lock | `syscall.Flock` (macOS) / `LockFileEx` (Windows) | Prevent two agents using same profile |
| Config lock | Same mechanism | Atomic config read-modify-write |
| Atomic writes | `os.CreateTemp` + `os.Rename` | Crash-safe file writes |

### Lock File Format

```json
// ~/.local/share/bpm/profiles/<name>/.bpm.lock
{
  "pid": 12345,
  "caller": "claude-code",
  "locked_at": "2026-03-29T10:00:00Z"
}
```

### Stale Lock Recovery
1. Read lock file → get PID
2. `os.FindProcess(pid)` → check if process alive
3. If dead → remove lock → retry acquisition

---

## Desktop App

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Framework | [Wails](https://wails.io/) | v2 | Go backend + web frontend in native window |
| Frontend | Vanilla HTML/CSS/JS | — | Simple UI, no framework overhead |

### Why Wails over Tauri/Electron?

| Criteria | Wails | Tauri | Electron |
|----------|-------|-------|----------|
| Backend language | **Go** ✅ (same as core) | Rust | Node.js |
| Binary size | ~10MB | ~5MB | ~150MB |
| Complexity | Low | Medium (need Rust) | Low |
| Cross-platform | macOS + Windows + Linux | Same | Same |
| Performance | Native WebView | Native WebView | Chromium |

**Key advantage**: Wails uses the same Go codebase as the CLI. The desktop app's backend directly imports `internal/profile`, `internal/browser`, etc. — no IPC, no sidecar binary, no duplicate logic.

---

## MCP Protocol

| Aspect | Detail |
|--------|--------|
| Transport | stdio (stdin/stdout JSON-RPC) |
| SDK | [mcp-go](https://github.com/mark3labs/mcp-go) |
| Tools exposed | 8 tools (profile CRUD, mapping, credentials, browser detection) |

### How AI IDEs Connect

```
┌──────────────────┐     stdio      ┌──────────────┐
│  AI IDE          │ ──────────────▶│  bpm serve   │
│  (Claude Code,   │  JSON-RPC      │  (MCP server)│
│   Cursor, etc.)  │◀──────────────│              │
└──────────────────┘                └──────────────┘
```

The AI IDE spawns `bpm serve` as a subprocess. Communication happens over stdin/stdout using JSON-RPC 2.0 messages per the MCP specification.

---

## Browser Integration

### Chromium Flags Used

| Flag | Purpose |
|------|---------|
| `--user-data-dir=<path>` | Point browser to isolated profile directory |
| `--no-first-run` | Skip first-run setup wizard |
| `--no-default-browser-check` | Skip default browser prompt |

### Supported Browsers

| Browser | macOS Path | Windows Path |
|---------|-----------|-------------|
| Chrome | `/Applications/Google Chrome.app` | `%ProgramFiles%\Google\Chrome\Application\chrome.exe` |
| Brave | `/Applications/Brave Browser.app` | `%ProgramFiles%\BraveSoftware\Brave-Browser\Application\brave.exe` |
| Edge | `/Applications/Microsoft Edge.app` | `%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe` |
| Arc | `/Applications/Arc.app` | — (macOS only) |

---

## Credential Handling

### Chromium SQLite Databases

| Database | File | Contains |
|----------|------|----------|
| Cookies | `Cookies` | Domain, cookie name, value (encrypted), expiry |
| Logins | `Login Data` | URL, username, password (encrypted) |

### Security Model

| Aspect | Approach |
|--------|----------|
| Password decryption | ❌ **Never** — bpm does not decrypt passwords |
| Cookie values | ❌ Read domain/count only, not values |
| Sync operation | Copies entire DB files (encrypted data stays encrypted) |
| OS keychain | Passwords remain tied to OS keychain (macOS Keychain / Windows DPAPI) |
| File permissions | Profile directories created with `0700` |

---

## Build & Distribution

| Method | Command / Detail |
|--------|-----------------|
| Dev build | `go build -o bpm .` |
| Cross-compile macOS | `GOOS=darwin GOARCH=arm64 go build -o bpm-darwin-arm64 .` |
| Cross-compile Windows | `GOOS=windows GOARCH=amd64 go build -o bpm-windows-amd64.exe .` |
| Install from source | `go install github.com/tquangkhai98/browser-profiles-manager@latest` |
| Desktop app | `wails build` → produces `.app` (macOS) / `.exe` (Windows) |
| Homebrew (planned) | `brew install bpm` |
| GitHub Releases | Signed binaries for macOS/Windows |

---

## Dependencies Summary

| Package | Purpose | License |
|---------|---------|---------|
| `github.com/spf13/cobra` | CLI framework | Apache 2.0 |
| `github.com/mark3labs/mcp-go` | MCP server SDK | MIT |
| `modernc.org/sqlite` | Pure Go SQLite (read credentials) | BSD |
| `github.com/wailsapp/wails/v2` | Desktop app framework | MIT |

**Total direct dependencies: 4** — intentionally minimal.

---

## Development Environment

| Tool | Purpose |
|------|---------|
| Go 1.22+ | Language runtime |
| Git | Version control |
| `gh` CLI | GitHub operations |
| Wails CLI | Desktop app dev (`wails dev`, `wails build`) |
| VS Code / AI IDE | Development |
