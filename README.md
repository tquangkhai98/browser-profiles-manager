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

## License

MIT
