---
name: bpm-mcp-setup
description: Setting up bpm MCP server in AI IDEs (Claude Code, Cursor, Antigravity). Configuration paths, installation commands, and troubleshooting.
---

# BPM MCP Server Setup

> How to configure bpm as an MCP server across AI IDEs.

---

## 1. Overview

bpm exposes all profile management features via MCP (Model Context Protocol) over stdio transport. Any MCP-compatible AI IDE can call bpm tools directly.

### Available MCP Tools

| Tool | Purpose |
|------|---------|
| `profile_create` | Create isolated browser profile |
| `profile_list` | List all profiles with lock status |
| `profile_use` | Launch browser with profile (acquires lock) |
| `profile_status` | Check lock status of a profile |
| `profile_delete` | Delete profile and data |
| `mapping_set` | Map project directory → profile |
| `mapping_get` | Resolve profile for a directory |
| `creds_inspect` | List credential domains (never decrypts) |
| `creds_sync` | Copy credential DBs between profiles |
| `browser_detect` | List installed Chromium browsers |

---

## 2. Installation

### Prerequisites

```bash
# Go ≥ 1.25 required
go install github.com/tquangkhai98/browser-profiles-manager@latest
```

### Auto-Install to AI IDE

```bash
# Install into a specific IDE's MCP config
bpm install claude-code
bpm install claude-desktop
bpm install cursor
bpm install antigravity

# Check which IDEs are configured
bpm install --list
```

---

## 3. IDE Configuration Paths

### Claude Code (CLI)

| Platform | Config Path |
|----------|-------------|
| macOS | `~/.claude/mcp_servers.json` |
| Windows | `%USERPROFILE%\.claude\mcp_servers.json` |

### Claude Desktop App

| Platform | Config Path |
|----------|-------------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |

### Cursor

| Platform | Config Path |
|----------|-------------|
| macOS | `~/.cursor/mcp.json` |
| Windows | `%USERPROFILE%\.cursor\mcp.json` |

### Antigravity

| Platform | Config Path |
|----------|-------------|
| macOS | `~/.gemini/settings.json` |
| Windows | `%USERPROFILE%\.gemini\settings.json` |

---

## 4. Manual Configuration

### MCP Config JSON

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

### With Full Path (if bpm not in PATH)

```json
{
  "mcpServers": {
    "bpm": {
      "command": "/usr/local/bin/bpm",
      "args": ["serve"]
    }
  }
}
```

---

## 5. Verification

```bash
# Test MCP server starts correctly
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | bpm serve

# Check bpm binary location
which bpm

# Verify installed version
bpm --version
```

---

## 6. Troubleshooting

| Issue | Solution |
|-------|----------|
| `command not found: bpm` | Add `$GOPATH/bin` to PATH |
| IDE can't connect | Check config JSON syntax, restart IDE |
| Profile locked | Another process has the lock — check `bpm status <name>` |
| Config file not found | Create parent directories first |
| Permission denied | Check file permissions on config file |

---

## 7. Architecture Notes

- Transport: **stdio** (stdin/stdout JSON-RPC)
- Library: `mcp-go` by mark3labs
- Server entry: `internal/mcp/server.go`
- Tool handlers: `internal/mcp/tools.go`
- CLI command: `cmd/serve.go`

> **Remember:** The MCP server runs as a child process of the IDE. It must exit cleanly when stdin closes.
