---
name: chromium-profiles
description: Chromium browser profile internals — user-data-dir structure, credential databases (Cookies, Login Data), SQLite schemas, OS-level encryption, and atomic file operations. Use when working with profile CRUD, credential sync, import/export, or debugging browser data issues.
---

# Chromium Profile Internals

> Understanding Chromium's profile structure, credential storage, and how bpm manages them.

---

## 1. User Data Directory

Chromium's `--user-data-dir` flag specifies the root directory for all profile data.

### Directory Structure

```
<user-data-dir>/
├── Default/                    # Active profile subdirectory
│   ├── Cookies                 # SQLite — session cookies
│   ├── Login Data              # SQLite — saved passwords (encrypted)
│   ├── History                 # SQLite — browsing history
│   ├── Bookmarks               # JSON — bookmarks
│   ├── Preferences             # JSON — browser settings
│   ├── Web Data                # SQLite — autofill data
│   ├── Extension Cookies       # SQLite — extension cookies
│   └── Extensions/             # Installed extensions
├── Local State                 # JSON — encryption keys, profiles metadata
├── First Run                   # Marker file
└── SingletonLock               # Lock file (exclusive access)
```

> **Critical:** Chromium reads credential data from `Default/` subdirectory, not the root. bpm's credential sync must target this path.

---

## 2. Credential Databases

### Cookies (SQLite)

```sql
-- Schema: Default/Cookies
CREATE TABLE cookies (
    creation_utc     INTEGER NOT NULL,
    host_key         TEXT NOT NULL,
    name             TEXT NOT NULL,
    value            TEXT NOT NULL,
    path             TEXT NOT NULL,
    expires_utc      INTEGER NOT NULL,
    is_secure        INTEGER NOT NULL,
    is_httponly       INTEGER NOT NULL,
    last_access_utc  INTEGER NOT NULL,
    has_expires      INTEGER NOT NULL DEFAULT 1,
    is_persistent    INTEGER NOT NULL DEFAULT 1,
    priority         INTEGER NOT NULL DEFAULT 1,
    encrypted_value  BLOB NOT NULL DEFAULT '',
    samesite         INTEGER NOT NULL DEFAULT -1,
    source_scheme    INTEGER NOT NULL DEFAULT 0,
    source_port      INTEGER NOT NULL DEFAULT -1,
    last_update_utc  INTEGER NOT NULL DEFAULT 0
);
```

### Login Data (SQLite)

```sql
-- Schema: Default/Login Data
CREATE TABLE logins (
    origin_url         TEXT NOT NULL,
    action_url         TEXT,
    username_element   TEXT,
    username_value     TEXT,
    password_element   TEXT,
    password_value     BLOB,        -- OS-encrypted blob
    signon_realm       TEXT NOT NULL,
    date_created       INTEGER NOT NULL,
    blacklisted_by_user INTEGER NOT NULL,
    scheme             INTEGER NOT NULL,
    password_type      INTEGER,
    times_used         INTEGER,
    date_last_used     INTEGER,
    date_password_modified INTEGER DEFAULT 0
);
```

---

## 3. OS-Level Encryption

### macOS
- Uses **macOS Keychain** to store encryption key
- Key stored under `Chromium Safe Storage` or `Chrome Safe Storage`
- Encrypted with PBKDF2 + AES-128-CBC
- Prefix: `v10` (3 bytes) before encrypted data

### Windows
- Uses **DPAPI** (Data Protection API)
- Key stored in `Local State` JSON (`os_crypt.encrypted_key`)
- Prefix: `v10` or `DPAPI` before encrypted data

> **bpm never decrypts passwords.** Credential sync copies the encrypted SQLite files. They only work on the same OS/user because the keychain/DPAPI keys are tied to the OS user account.

---

## 4. bpm Profile Management

### Profile Storage

```
# Default paths
macOS:   ~/.local/share/bpm/profiles/<name>/
Windows: %LOCALAPPDATA%\bpm\profiles\<name>\

# Each profile directory IS a user-data-dir
~/.local/share/bpm/profiles/work-staging/
├── Default/
│   ├── Cookies
│   ├── Login Data
│   └── ...
├── Local State
└── ...
```

### File Locking

| Platform | Mechanism |
|----------|-----------|
| macOS/Linux | `flock()` on `.bpm.lock` file |
| Windows | `LockFileEx()` on `.bpm.lock` file |

```go
// Pattern: always defer release
releaseLock, err := profile.AcquireLock(dataDir, "caller-id")
if err != nil {
    return err
}
defer releaseLock()
```

### Credential Sync Flow

1. Validate source and target profiles exist
2. Close any browsers using target profile
3. Backup target's credential DBs
4. Copy `Cookies` and `Login Data` from source `Default/` to target `Default/`
5. Use atomic write: temp file → rename

```go
// Internal flow (credential.Sync)
copied, err := credential.Sync(srcDataDir, dstDataDir)
```

### Atomic File Operations

All config writes use temp file + rename pattern:

```go
// Write to temp file first
tmpFile, _ := os.CreateTemp(filepath.Dir(target), ".bpm-tmp-*")
tmpFile.Write(data)
tmpFile.Close()
// Atomic rename
os.Rename(tmpFile.Name(), target)
```

---

## 5. Import/Export

### Import Existing Chrome Profile

```bash
# Source: Chrome's default profile directory
# macOS: ~/Library/Application Support/Google/Chrome/Default
# Windows: %LOCALAPPDATA%\Google\Chrome\User Data\Default

bpm import "~/Library/Application Support/Google/Chrome/Default" my-chrome
```

Import copies the entire profile directory into bpm's managed storage.

### Export Profile for Backup

```bash
bpm export work-staging ~/Desktop/work-staging-backup
```

---

## 6. Common Pitfalls

| Pitfall | Explanation |
|---------|-------------|
| Credentials in root, not `Default/` | Chromium only reads from `Default/` subdirectory |
| Plaintext passwords don't work | Must be encrypted with OS keychain/DPAPI (`v10` format) |
| Cross-OS sync fails | Encryption keys are tied to OS user account |
| Profile corruption | Opening same profile in 2 browsers simultaneously |
| SQLite WAL mode | Must open databases read-only (`?mode=ro`) for inspection |

---

## 7. Inspection Queries

```sql
-- Count cookies per domain
SELECT host_key, COUNT(*) as cookie_count
FROM cookies GROUP BY host_key ORDER BY cookie_count DESC;

-- Count saved logins per site
SELECT signon_realm, COUNT(*) as login_count
FROM logins GROUP BY signon_realm ORDER BY login_count DESC;

-- Check if cookies exist for a domain
SELECT COUNT(*) FROM cookies WHERE host_key LIKE '%github.com%';
```

> **Always open databases read-only:** `sqlite3 "file:Cookies?mode=ro"`

---

> **Remember:** bpm is a profile manager, not a password manager. It moves encrypted blobs around — never looks inside them.
