# Browser Profiles Manager (bpm)

## Problem

AI IDEs (Claude Code, Cursor, Antigravity...) each spawn their own Chromium instances with separate/ephemeral profiles. This causes:
- Login sessions lost between runs
- Tab conflicts when parallel agents use the same profile
- No way to check or sync credentials across tools
- No profile awareness for headless/cron agents

**No existing tool solves centralized browser profile management across AI IDEs.**

**Goal**: `bpm` — a single Go binary (CLI + MCP server) with a simple desktop app for visual management. Supports macOS and Windows.

---

## Architecture

```
┌──────────────────────────────────────┐
│  AI IDEs (Claude Code, Cursor, ...)  │
│  └─ MCP config → bpm serve          │
├──────────────────────────────────────┤
│                                      │
│  ┌───────────┐  ┌──────────────┐    │
│  │  bpm CLI   │  │  bpm serve   │    │
│  │  (cobra)   │  │  (MCP/stdio) │    │
│  └─────┬──────┘  └──────┬──────┘    │
│        └────────┬────────┘           │
│        ┌────────▼────────┐           │
│        │   Go Core Lib    │           │
│        │ profile/ config/ │           │
│        │ browser/ creds/  │           │
│        └────────┬────────┘           │
│        ┌────────▼────────┐           │
│        │ Config + Profiles│           │
│        │ (JSON + filesystem)         │
│        └─────────────────┘           │
│                                      │
│  ┌──────────────────────────────┐    │
│  │       Desktop App (Wails)     │    │
│  │    Simple UI — calls bpm CLI  │    │
│  └──────────────────────────────┘    │
└──────────────────────────────────────┘
```

**Key design**: Each profile = its own `--user-data-dir` directory. Full filesystem-level isolation.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Core + CLI | Go + Cobra |
| MCP Server | mcp-go (stdio) |
| Desktop App | Wails v2 (Go backend + Web frontend) |
| Frontend | Vanilla HTML/CSS/JS (simple, no framework) |
| Storage | JSON config + filesystem |
| Locking | flock (macOS) / LockFileEx (Windows) |
| Credential read | modernc.org/sqlite (pure Go, read-only) |

> **Why Wails over Tauri?** Same language as core (Go), no need for Rust. Single codebase, simpler build. Cross-platform macOS + Windows.

---

## Core Features

### 1. Profile CRUD
- `bpm create <name> [--browser chrome]` — Create isolated profile
- `bpm list [--json]` — List all profiles with status (free/locked)
- `bpm delete <name> [--force]` — Delete profile + data
- `bpm status <name>` — Check lock status

### 2. Browser Launch + Lock
- `bpm use <name>` — Launch browser with `--user-data-dir`
- File-based lock prevents 2 agents using same profile
- Auto-detect installed Chromium browsers
- `bpm detect` — List installed browsers

### 3. Directory Mapping
- `bpm map <dir> <profile>` — Map project dir to profile
- `bpm map --auto` — Auto-resolve profile for current directory
- `bpm map --list` — Show all mappings

### 4. Credential Sync & Import
- `bpm creds <name>` — Inspect what sites have cookies/logins in a profile
- `bpm sync <source> <target>` — Copy cookies/logins from one profile to another
- `bpm import <path> <name>` — Import an existing Chrome profile into bpm
- `bpm export <name> <path>` — Export a bpm profile for backup

### 5. MCP Server
- `bpm serve` — Start MCP server (stdio transport)
- Any AI IDE adds to MCP config and it just works

### 6. Desktop App (Simple)
- Profile list with status (free/locked)
- Create / delete profiles
- One-click browser launch
- View credentials per profile
- Sync credentials between profiles (select source → target)
- Import existing Chrome profile
- Simple, functional — not a fancy dashboard

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `bpm create <name>` | Create isolated profile |
| `bpm list [--json]` | List profiles with status |
| `bpm delete <name> [--force]` | Delete profile + data |
| `bpm status <name>` | Check lock/usage status |
| `bpm use <name>` | Launch browser with profile |
| `bpm detect` | List installed browsers |
| `bpm map <dir> <profile>` | Map project dir → profile |
| `bpm map --auto` | Auto-resolve profile for cwd |
| `bpm map --list` | Show all mappings |
| `bpm creds <name>` | Inspect credentials in profile |
| `bpm sync <src> <dst>` | Sync credentials between profiles |
| `bpm import <path> <name>` | Import existing Chrome profile |
| `bpm export <name> <path>` | Export profile for backup |
| `bpm serve` | Start MCP server |

---

## Data Model

```go
type Profile struct {
    Name      string     `json:"name"`
    Browser   string     `json:"browser"`
    DataDir   string     `json:"data_dir"`
    CreatedAt time.Time  `json:"created_at"`
    LastUsed  *time.Time `json:"last_used_at"`
}

type Mapping struct {
    Directory string `json:"directory"`
    Profile   string `json:"profile"`
}

type Config struct {
    DefaultBrowser string    `json:"default_browser"`
    Profiles       []Profile `json:"profiles"`
    Mappings       []Mapping `json:"mappings"`
}
```

**Storage paths:**
- macOS: `~/.config/bpm/config.json` + `~/.local/share/bpm/profiles/<name>/`
- Windows: `%APPDATA%\bpm\config.json` + `%LOCALAPPDATA%\bpm\profiles\<name>\`

---

## Project Structure

```
browser-profiles-manager/
├── go.mod
├── main.go
├── cmd/                        # CLI commands (Cobra)
│   ├── root.go
│   ├── create.go
│   ├── list.go
│   ├── delete.go
│   ├── use.go
│   ├── status.go
│   ├── detect.go
│   ├── map.go
│   ├── creds.go
│   ├── sync.go
│   ├── importcmd.go
│   ├── export.go
│   └── serve.go
├── internal/
│   ├── profile/
│   │   ├── model.go            # Profile struct
│   │   ├── store.go            # CRUD operations
│   │   └── lock.go             # File-based locking
│   ├── browser/
│   │   ├── detect.go           # Find installed browsers
│   │   ├── launch.go           # Launch with --user-data-dir
│   │   └── registry.go         # Browser paths per OS
│   ├── credential/
│   │   ├── inspect.go          # Read cookies/logins from profile
│   │   └── sync.go             # Copy credential DBs between profiles
│   ├── mapping/
│   │   └── mapping.go          # Dir → profile mapping
│   ├── config/
│   │   └── config.go           # Load/save config.json
│   └── mcp/
│       ├── server.go           # MCP server setup
│       └── tools.go            # MCP tool handlers
├── desktop/                    # Wails desktop app
│   ├── app.go                  # Go backend (calls core lib)
│   └── frontend/
│       ├── index.html
│       ├── style.css
│       └── main.js
├── docs/
│   └── PRD.md
└── README.md
```

---

## Browser Detection

### macOS
| Browser | App Path |
|---------|----------|
| Chrome | `/Applications/Google Chrome.app` |
| Brave | `/Applications/Brave Browser.app` |
| Edge | `/Applications/Microsoft Edge.app` |
| Arc | `/Applications/Arc.app` |

### Windows
| Browser | Registry / Path |
|---------|----------------|
| Chrome | `%ProgramFiles%\Google\Chrome\Application\chrome.exe` |
| Brave | `%ProgramFiles%\BraveSoftware\Brave-Browser\Application\brave.exe` |
| Edge | `%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe` |

---

## Implementation Phases

### Phase 1: CLI Core (~3 days)
- Go module + Cobra skeleton
- config.go — load/save with file locking (cross-platform)
- Profile CRUD (create, list, delete, status)
- Browser detect + launch with `--user-data-dir`
- File lock (flock on macOS, LockFileEx on Windows)

### Phase 2: Credential + Mapping + MCP (~3 days)
- Credential inspect (read SQLite cookie/login DBs)
- Credential sync (copy DBs between profiles)
- Import/export profile directories
- Directory → profile mapping
- MCP server with all tools

### Phase 3: Desktop App (~3 days)
- Wails v2 project setup
- Simple UI: profile list, create/delete, launch, creds view
- Sync flow: select source → target → sync
- Import existing Chrome profile
- Build for macOS + Windows

### Phase 4: Polish (~1 day)
- `--json` output for all commands
- Error messages with actionable hints
- README with install + MCP config examples
- Test with Claude Code, Cursor, Antigravity

**Total: ~10 days**

---

## Concurrency & Safety

- **Profile lock**: flock (macOS) / LockFileEx (Windows) + PID metadata
- **Config lock**: atomic read-modify-write
- **Atomic writes**: write temp file → rename
- **Stale lock cleanup**: check if PID alive, auto-remove if dead

---

## Verification

1. `bpm create test && bpm list && bpm use test` → browser opens with clean profile
2. Create 2 profiles, login in one, `bpm creds` shows sites, `bpm sync` copies to other
3. `bpm import` existing Chrome profile → works immediately
4. MCP config in Claude Code → verify tools work
5. `bpm use` on locked profile → proper error
6. Desktop app: create, launch, sync, import — all work
7. Build and test on both macOS and Windows

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/mark3labs/mcp-go` | MCP server SDK |
| `modernc.org/sqlite` | Read cookie/login DBs (pure Go) |
| `github.com/wailsapp/wails/v2` | Desktop app framework |
