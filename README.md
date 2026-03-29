# bpm — Browser Profiles Manager

> Centralized browser profile management for AI IDEs

**bpm** is a CLI tool + MCP server that manages isolated Chromium browser profiles across AI-powered development environments like Claude Code, Cursor, and Antigravity.

## Problem

AI IDEs spawn their own Chromium instances with ephemeral profiles:
- 🔑 Login sessions lost between runs
- 💥 Profile conflicts when parallel agents use the same directory
- 🔄 No way to sync credentials across tools
- 🤖 No profile awareness for headless/cron agents

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

## Quick Start

```bash
# Install
go install github.com/tquangkhai98/browser-profiles-manager@latest

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

## MCP Integration

Add to your AI IDE's MCP config (Claude Code, Cursor, Antigravity, etc.):

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
| `mapping_get` | Resolve profile for a directory |
| `creds_inspect` | List credentials in a profile |
| `creds_sync` | Sync credentials between profiles |
| `browser_detect` | List installed browsers |

## CLI Commands

```
bpm create <name>          Create isolated profile
bpm list [--json]          List profiles with status
bpm delete <name>          Delete profile + data
bpm status <name>          Check lock/usage status
bpm use <name>             Launch browser with profile
bpm detect                 List installed browsers
bpm map <dir> <profile>    Map project dir → profile
bpm map --auto             Auto-resolve profile for cwd
bpm map --list             Show all mappings
bpm creds <name>           Inspect credentials in profile
bpm sync <src> <dst>       Sync credentials between profiles
bpm import <path> <name>   Import existing Chrome profile
bpm export <name> <path>   Export profile for backup
bpm serve                  Start MCP server
```

## Development

> 📖 Chưa biết Go? Xem [Go Quick Guide](docs/GO_GUIDE.md) — hướng dẫn cơ bản dành cho developer mới.

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | ≥ 1.25 | `brew install go` |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Node.js | ≥ 18 | `brew install node` (for desktop frontend) |

### Project Structure

```
browser-profiles-manager/
├── main.go                 # CLI entry point
├── go.mod                  # Go module & dependencies
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go             #   Root command setup
│   ├── create.go           #   bpm create
│   ├── list.go             #   bpm list
│   ├── use.go              #   bpm use
│   ├── serve.go            #   bpm serve (MCP server)
│   └── ...                 #   Other subcommands
├── internal/               # Core business logic (private packages)
│   ├── profile/            #   Profile CRUD
│   ├── browser/            #   Browser detection & launch
│   ├── credential/         #   Credential read (SQLite)
│   ├── mapping/            #   Directory ↔ profile mapping
│   ├── config/             #   Configuration management
│   └── mcp/                #   MCP server implementation
├── desktop/                # Wails desktop app (separate binary)
│   ├── main.go             #   Desktop entry point
│   ├── app.go              #   App struct (Go ↔ JS bridge)
│   ├── wails.json          #   Wails configuration
│   └── frontend/           #   HTML/CSS/JS frontend
├── build/                  # Build output
└── docs/                   # Documentation
```

### Build & Run

```bash
# ────────────────────────────────────────────
# 🔨  Build CLI binary
# ────────────────────────────────────────────
go build -o bpm .                   # Build → ./bpm
go build -o bpm -v .                # Build verbose (show packages)

# ────────────────────────────────────────────
# ▶️  Run without building
# ────────────────────────────────────────────
go run .                            # Run CLI directly
go run . list                       # Run a specific command
go run . create test-profile        # Run with arguments

# ────────────────────────────────────────────
# 🖥️  Desktop App (Wails)
# ────────────────────────────────────────────
cd desktop
wails dev                           # Dev mode with hot-reload
wails build                         # Production build → build/bin/
wails build -debug                  # Build with DevTools enabled

# ────────────────────────────────────────────
# 📦  Install globally
# ────────────────────────────────────────────
go install .                        # Install to $GOPATH/bin/bpm
```

### Debug

```bash
# ────────────────────────────────────────────
# 🐛  Debug with Delve (Go debugger)
# ────────────────────────────────────────────

# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug CLI
dlv debug .                         # Start debugger on main.go
dlv debug . -- list                 # Debug with CLI arguments
dlv debug . -- create my-profile    # Debug specific command

# Debug with breakpoint
dlv debug . -- use test             # Then in dlv console:
                                    #   break cmd/use.go:25
                                    #   continue
                                    #   print profileName
                                    #   next / step / continue

# Debug desktop app
cd desktop
dlv debug .                         # Debug Wails app

# ────────────────────────────────────────────
# 🔍  Build with debug symbols (no optimization)
# ────────────────────────────────────────────
go build -gcflags="all=-N -l" -o bpm .    # Full debug info
```

### Logging & Troubleshooting

```bash
# ────────────────────────────────────────────
# 📋  Print debug info in code
# ────────────────────────────────────────────
# Use fmt.Printf / fmt.Println for quick debug output:
#   fmt.Printf("[DEBUG] profile: %+v\n", profile)
#   fmt.Printf("[DEBUG] err: %v\n", err)

# Use log package for structured logging:
#   log.Printf("loading config from %s", path)
#   log.Fatalf("critical error: %v", err)   ← exits program

# ────────────────────────────────────────────
# 🔎  Verbose run (see what Go does)
# ────────────────────────────────────────────
go build -v .                       # Show packages being compiled
go build -x .                       # Show ALL build commands executed
go run -race .                      # Run with race condition detector

# ────────────────────────────────────────────
# 📊  Profile & trace
# ────────────────────────────────────────────
go test -cpuprofile cpu.prof -bench .   # CPU profiling
go tool pprof cpu.prof                  # Analyze profile
go test -trace trace.out                # Execution trace
go tool trace trace.out                 # View trace in browser
```

### Testing

```bash
# ────────────────────────────────────────────
# 🧪  Run tests
# ────────────────────────────────────────────
go test ./...                       # Run ALL tests in project
go test ./internal/profile/         # Test specific package
go test ./internal/profile/ -v      # Verbose (show each test name)
go test ./internal/profile/ -run TestCreate  # Run specific test
go test ./... -count=1              # No cache, fresh run
go test ./... -cover                # Show code coverage %
go test ./... -coverprofile=coverage.out     # Generate coverage file
go tool cover -html=coverage.out             # View coverage in browser
```

### Dependency Management

```bash
# ────────────────────────────────────────────
# 📦  Manage Go modules
# ────────────────────────────────────────────
go mod tidy                         # Clean unused deps, add missing ones
go mod download                     # Download all dependencies
go mod vendor                       # Copy deps to vendor/ folder
go mod graph                        # Show dependency graph
go get github.com/some/package      # Add a new dependency
go get -u ./...                     # Update all dependencies
go list -m all                      # List all modules
```

### Clean & Reset

```bash
# ────────────────────────────────────────────
# 🧹  Cleanup
# ────────────────────────────────────────────
go clean                            # Remove build cache for this package
go clean -cache                     # Clear entire build cache
go clean -testcache                 # Clear test result cache
rm -f bpm                           # Remove built binary
rm -rf build/                       # Remove build output
```

### Common Development Workflow

```bash
# 1. Pull latest & sync deps
git pull && go mod tidy

# 2. Make changes to code...

# 3. Quick test run
go run . list

# 4. Run tests
go test ./...

# 5. Build final binary
go build -o bpm .

# 6. Test the binary
./bpm list
./bpm create my-test

# 7. Commit
git add . && git commit -m "feat: add new feature"
```

---

## Screenshots

<p align="center">
  <img src="docs/wireframe/01-profile-list.png" width="800" alt="Profile List" />
  <br><em>Profile List — manage all browser profiles</em>
</p>

<p align="center">
  <img src="docs/wireframe/03-credential-view.png" width="800" alt="Credential View" />
  <br><em>Credential View — inspect & sync cookies/logins</em>
</p>

<p align="center">
  <img src="docs/wireframe/04-settings.png" width="800" alt="Settings" />
  <br><em>Settings — configure defaults and MCP</em>
</p>

## Tech Stack

| Layer | Technology |
|-------|-----------|
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
| Linux | 🔜 Planned |

## Documentation

- [PLAN.md](PLAN.md) — Technical implementation plan
- [docs/PRD.md](docs/PRD.md) — Product requirements document
- [docs/STITCH.md](docs/STITCH.md) — Design wireframes (Stitch MCP)
- [docs/GO_GUIDE.md](docs/GO_GUIDE.md) — Go quick guide for beginners

## License

MIT
