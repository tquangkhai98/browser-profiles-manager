---
name: browser-automation
description: Browser automation integration patterns with bpm profiles. Covers Playwright MCP, Chrome DevTools MCP, Browser Use MCP for multi-role testing, CI/CD, web scraping, and data entry automation.
---

# Browser Automation Integration

> Using bpm profiles with browser automation MCP servers for AI agent workflows.

---

## 1. Core Concept

bpm manages **persistent browser sessions**. Automation MCPs control **browser actions**. Together they enable:

```
Login once (manual) → AI agent automates with full authentication → No re-login needed
```

---

## 2. Integration Options

### 🥇 Playwright MCP — Best Overall

Stable, official. Uses accessibility tree for reliable element interaction.

```json
{
  "mcpServers": {
    "browser-admin": {
      "command": "npx",
      "args": [
        "@playwright/mcp@latest",
        "--user-data-dir", "~/.local/share/bpm/profiles/lms-admin"
      ]
    }
  }
}
```

**Best for:** E2E testing, form automation, multi-role testing.

### 🥈 Chrome DevTools MCP — Debug & Inspect

Connect to a running Chrome with remote debugging.

```bash
# Step 1: Launch with debug port
bpm use lms-admin --debug-port=9222
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

**Best for:** DOM inspection, network monitoring, console access.

### 🥉 Browser Use MCP — Natural Language

AI-native — describe actions in plain language.

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

**Best for:** Natural language automation, non-technical workflows.

---

## 3. Use Case Patterns

### Multi-Role Testing (LMS/CRM)

```bash
# Create role-specific profiles
bpm create lms-admin
bpm create lms-teacher
bpm create lms-student

# Login once per role (manual)
bpm use lms-admin      # → login as admin
bpm use lms-teacher    # → login as teacher
bpm use lms-student    # → login as student

# AI agent uses profiles with persistent sessions ✅
```

MCP config for all roles:
```json
{
  "mcpServers": {
    "browser-admin": {
      "command": "npx",
      "args": ["@playwright/mcp@latest", "--user-data-dir", "~/.local/share/bpm/profiles/lms-admin"]
    },
    "browser-teacher": {
      "command": "npx",
      "args": ["@playwright/mcp@latest", "--user-data-dir", "~/.local/share/bpm/profiles/lms-teacher"]
    },
    "browser-student": {
      "command": "npx",
      "args": ["@playwright/mcp@latest", "--user-data-dir", "~/.local/share/bpm/profiles/lms-student"]
    }
  }
}
```

### CI/CD Pipeline

```bash
# Pre-authenticated profile for staging tests
bpm create staging-qa
bpm use staging-qa  # login once

# CI agent runs E2E tests with auth
bpm use staging-qa --headless
```

### Web Scraping with Auth

```bash
# Access authenticated dashboards
bpm create dashboard-monitor
bpm use dashboard-monitor  # login to dashboard

# AI scrapes with full session
```

### Data Entry Automation

```bash
# AI fills forms across multiple platforms
bpm create erp-data-entry
bpm use erp-data-entry  # login to ERP

# AI agent automates form filling with persistent session
```

---

## 4. Profile Path Resolution

```
# Default profile storage
macOS:   ~/.local/share/bpm/profiles/<name>/
Windows: %LOCALAPPDATA%\bpm\profiles\<name>\

# Custom path via config
bpm config set profiles-dir /path/to/custom/dir
```

When configuring `--user-data-dir` in MCP configs, use the full path to the profile directory.

---

## 5. Safety Rules

| Rule | Reason |
|------|--------|
| Never open same profile in 2 browsers | File lock prevents corruption |
| Use separate profiles per role | Avoid session conflicts |
| Don't decrypt credentials | bpm never accesses encrypted passwords |
| Sync credentials via `bpm sync` | Atomic copy with backup |

---

## 6. Comparison Table

| Tool | Session Persist | AI Control | Multi-Role | Best For |
|------|:-:|:-:|:-:|----------|
| **Playwright MCP** | ✅ | ✅ | ✅ | E2E testing, form automation |
| **Chrome DevTools MCP** | ✅ | ✅ | ✅ | Debugging, network inspection |
| **Browser Use MCP** | ✅ | ✅ | ✅ | Natural language automation |
| bpm CLI only | ✅ | ❌ | ✅ | Manual testing |

---

> **Key insight:** bpm is the **session persistence layer**. Automation MCPs are the **action layer**. Combine them for authenticated AI automation.
