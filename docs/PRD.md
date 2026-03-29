# Browser Profiles Manager (bpm) — PRD

> **Version**: 1.1 | **Date**: 2026-03-29 | **Status**: Draft

---

## 1. Summary

**bpm** is a single Go binary (CLI + MCP server) with a simple desktop app that centrally manages isolated Chromium browser profiles for AI development environments. It solves lost login sessions, profile conflicts, and scattered credentials across AI IDEs like Claude Code, Cursor, and Antigravity.

**Platforms**: macOS and Windows.

---

## 2. Problem

| Pain Point | Impact |
|-----------|--------|
| AI IDEs spawn ephemeral browser profiles — login sessions lost every run | Re-authenticate repeatedly (OAuth, 2FA) |
| Parallel AI agents conflict on the same profile directory | Browser crashes, corrupt data |
| No way to check or sync credentials across profiles | Each tool starts from scratch |
| Browser profile directories scattered across filesystem | Hard to audit, clean up, back up |

### What is "Credential Sync"?

When you login to GitHub in one browser profile, that login is stored as cookies and saved passwords in that profile's database files. **Credential sync** means copying those cookies/passwords from one profile to another — so you don't have to login again in the second profile. Think of it as "clone my login sessions to another profile".

### Market Gap

No existing tool provides centralized browser profile management across AI IDEs:
- **Browser MCP** → automation, not profile management
- **browser-use** → per-agent config, not centralized
- **AdsPower MCP** → anti-detect, closed ecosystem
- **chrome-cli** → controls tabs/windows, not profiles

---

## 3. Target Users

**Primary**: Developers using AI IDEs (Claude Code, Cursor, Antigravity) who need persistent, isolated browser sessions.

---

## 4. Core Features

### F1: Profile CRUD
Create, list, delete isolated browser profiles. Each profile = separate `--user-data-dir` directory.

| Command | Behavior |
|---------|----------|
| `bpm create <name>` | Creates profile directory |
| `bpm list [--json]` | Shows all profiles: name, browser, last used, status |
| `bpm delete <name>` | Removes profile. Confirms unless `--force` |
| `bpm status <name>` | Shows lock status (who's using it) |

### F2: Browser Launch + Lock
Launch Chromium browser with isolated profile. File lock prevents two agents using same profile.

| Command | Behavior |
|---------|----------|
| `bpm use <name>` | Launches browser with `--user-data-dir`, acquires lock |
| `bpm detect` | Lists installed Chromium browsers |

### F3: Directory Mapping
Map project directories to profiles → auto-resolve which profile to use.

| Command | Behavior |
|---------|----------|
| `bpm map <dir> <profile>` | Creates mapping |
| `bpm map --auto` | Resolves profile for current directory |
| `bpm map --list` | Shows all mappings |

### F4: Credential Sync & Import
Check what logins exist in a profile, sync them to another profile, or import an existing browser profile.

| Command | What it does |
|---------|-------------|
| `bpm creds <name>` | Lists domains with cookies/saved logins (e.g., github.com: 12 cookies, 1 login) |
| `bpm sync <src> <dst>` | Copies cookies + login databases from source → target profile |
| `bpm import <path> <name>` | Imports existing Chrome/Brave profile folder into bpm |
| `bpm export <name> <path>` | Exports profile for backup or sharing |

**How sync works:**
1. Reads Chromium SQLite databases (`Cookies`, `Login Data`) from source profile
2. Backs up target profile's databases
3. Copies databases to target profile
4. ⚠️ Passwords remain encrypted — bpm never decrypts them. They work because Chromium uses OS-level keychain (macOS Keychain / Windows DPAPI)

**How import works:**
- Point to an existing Chrome profile (e.g., `~/Library/Application Support/Google/Chrome/Default/`)
- bpm copies it into `~/.local/share/bpm/profiles/<name>/`
- You now have your existing authenticated sessions managed by bpm

### F5: MCP Server
Expose all features as MCP tools. Any AI IDE can connect.

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

| Tool | Description |
|------|-------------|
| `profile_create` | Create profile |
| `profile_list` | List all profiles |
| `profile_use` | Launch browser with profile |
| `profile_status` | Check lock status |
| `mapping_get` | Resolve profile for directory |
| `creds_inspect` | List credentials in profile |
| `creds_sync` | Sync credentials between profiles |
| `browser_detect` | List installed browsers |

### F6: Desktop App (Simple)
A lightweight desktop app for users who prefer a GUI over CLI.

**Screens:**
1. **Profile List** — see all profiles with status, one-click launch
2. **Create Profile** — simple form: name + browser selection
3. **Credential View** — see which sites have cookies/logins in a profile
4. **Sync Flow** — select source profile → target profile → sync button
5. **Import** — pick existing Chrome profile folder → import into bpm

**Style**: Clean, simple, functional. Dark theme. No fancy dashboard or analytics.

---

## 5. CLI UX Examples

```bash
$ bpm list
NAME           BROWSER   LAST USED     STATUS
work-staging   chrome    2h ago        free
personal       brave     3 days ago    free
ci-scraper     chrome    12h ago       locked (PID: 58291)

$ bpm creds work-staging
DOMAIN                    COOKIES   LOGINS
github.com                12        1
accounts.google.com       8         1
app.staging.example.com   5         0

$ bpm sync work-staging personal
✓ Backed up target profile "personal"
✓ Synced 25 cookies from work-staging → personal
✓ Synced 2 logins from work-staging → personal
Done in 0.3s

$ bpm import ~/Library/Application\ Support/Google/Chrome/Default my-chrome
✓ Imported Chrome profile as "my-chrome"
✓ Size: 245 MB
✓ Found: 48 cookies, 12 logins

$ bpm detect
BROWSER          VERSION   PATH
Google Chrome    126.0     /Applications/Google Chrome.app
Brave Browser    1.67      /Applications/Brave Browser.app
```

---

## 6. Non-Functional Requirements

| Area | Requirement |
|------|------------|
| Performance | CLI < 200ms, browser launch < 1s |
| Security | Profile dirs 0700, bpm never decrypts passwords |
| Platform | macOS + Windows |
| Distribution | Single binary: `go install`, GitHub Releases, Homebrew (macOS) |

---

## 7. Release Plan

| Phase | Scope | Timeline |
|-------|-------|----------|
| Phase 1 | Profile CRUD + Browser Launch + Lock | ~3 days |
| Phase 2 | Credentials + Mapping + MCP Server | ~3 days |
| Phase 3 | Desktop App (simple) | ~3 days |
| Phase 4 | Polish + README + Testing | ~1 day |
| **Total** | **Full v1.0** | **~10 days** |

---

## 8. Future (NOT v1)

- Linux support
- Cloud sync between machines
- Browser extension
- Fine-grained credential sync (specific cookies/domains only)
- Profile templates

---

*End of PRD v1.1*
