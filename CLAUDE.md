# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

**bpm** (Browser Profiles Manager) — a CLI + MCP server + desktop app for managing isolated Chromium browser profiles across AI IDEs (Claude Code, Cursor, Antigravity). Prevents login session loss and profile conflicts when multiple AI agents spawn browser instances.

## Build & Run Commands

```bash
# CLI
go build -o bpm .            # Build binary
go run . <command>            # Run without building (e.g., go run . list)

# Desktop (Wails v2)
cd desktop && wails dev       # Dev mode with hot-reload
cd desktop && wails build     # Production build → build/bin/

# Tests
go test ./...                 # All tests
go test ./internal/profile/   # Single package
go test -run TestFoo ./...    # Single test by name

# Maintenance
go mod tidy                   # Clean/add dependencies
go vet ./...                  # Lint
```

## Architecture

### Three interfaces, one core:
- **CLI** (`main.go` → `cmd/`): Cobra-based commands. Each file in `cmd/` is one subcommand.
- **MCP Server** (`cmd/serve.go` → `internal/mcp/`): Exposes profile/mapping/credential tools over stdio using `mcp-go`.
- **Desktop** (`desktop/`): Wails v2 app. `app.go` is the Go backend bound to a vanilla JS frontend (no npm, no build step). Auto-generated `wailsjs/` bindings are gitignored.

### Core packages (`internal/`):
- **profile**: CRUD + file-based locking (flock on Unix, LockFileEx on Windows). `AcquireLock()` returns a release function — always `defer` it.
- **browser**: Detects installed Chromium browsers by platform-specific paths (`registry.go`), launches with `--user-data-dir`.
- **config**: Single JSON file (`~/.config/bpm/config.json`) holds profiles + mappings. Uses atomic writes (temp file + rename). `LoadWithLock`/`SaveWithLock` for concurrent safety.
- **credential**: Read-only SQLite inspection of Cookies/Login Data (never decrypts). `Sync` copies DB files between profiles.
- **mapping**: Maps directories to profiles with parent-directory fallback resolution.

### Platform abstraction:
Build-tag files (`_unix.go` / `_windows.go`) handle OS-specific paths, file locking, and process checking.

### Key patterns:
- Atomic file writes everywhere (temp + rename)
- Profile directories created with 0700, lock files with 0600
- Credentials opened read-only (`?mode=ro`)
- Errors wrapped with `fmt.Errorf("context: %w", err)`
