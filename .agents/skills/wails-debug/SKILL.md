---
name: wails-debug
description: Debug Wails v2 desktop applications. Covers dev mode debugging, Go backend + frontend JS debugging, DevTools, build flags, and common Wails troubleshooting patterns.
---

# Wails Debug Skill — Debugging Wails v2 Desktop Apps

> This skill provides instructions for debugging Wails v2 applications that combine a Go backend with a web frontend.

## Architecture Overview

Wails apps have TWO layers to debug:
1. **Go backend** — business logic, data access, system calls
2. **Web frontend** — HTML/CSS/JS rendered in a webview

Each layer requires different tools. This skill covers both.

---

## 1. Dev Mode (Primary Debug Method)

```bash
# Start Wails in dev mode — enables hot-reload + DevTools
cd desktop
wails dev

# Dev mode with verbose logging
wails dev -loglevel debug

# Dev mode opening browser DevTools automatically
wails dev -browser

# Dev mode with custom frontend dev server (e.g., Vite)
wails dev -frontenddevserverurl http://localhost:5173
```

### What `wails dev` gives you:
- **Hot reload** for frontend changes
- **Go rebuild** on backend changes
- **DevTools access** (right-click → Inspect, or Cmd+Option+I on macOS)
- **Console logging** from both Go and JS sides

---

## 2. Debug Go Backend (with Delve)

### Method A: Use `wails dev` + attach Delve

```bash
# Terminal 1: Start Wails dev
cd desktop
wails dev

# Terminal 2: Find PID and attach Delve
dlv attach $(pgrep -f "bpm-desktop\|wails")
```

### Method B: Build debug binary + run with Delve

```bash
# Build with debug symbols and dev tag
cd desktop
go build -gcflags="all=-N -l" -tags dev -o bpm-debug .

# Run with Delve
dlv exec ./bpm-debug
```

### Method C: Direct dlv debug (simpler, may miss frontend assets)

```bash
cd desktop
dlv debug . -- 

# Set breakpoints on your App methods
b App.GetProfiles
b App.CreateProfile
c
```

### VS Code launch.json for Wails

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Wails App",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/desktop",
      "buildFlags": "-tags dev",
      "env": {
        "CGO_ENABLED": "1"
      }
    },
    {
      "name": "Attach to Wails Dev",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": 0
    }
  ]
}
```

---

## 3. Debug Frontend (JavaScript/WebView)

### Using DevTools in dev mode

```bash
wails dev
# Then in the app window:
# macOS: Cmd + Option + I  (or right-click → Inspect Element)
# Linux: Ctrl + Shift + I
```

### Console logging from Go → Frontend

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// Log to frontend DevTools console
runtime.LogDebug(a.ctx, "Debug message from Go")
runtime.LogInfo(a.ctx, "Info message")
runtime.LogWarning(a.ctx, "Warning message")
runtime.LogError(a.ctx, "Error message")

// Also visible in terminal running `wails dev`
```

### Console logging from Frontend → Go

```javascript
// In your frontend JS/TS code
console.log("Frontend debug:", data);
console.error("Frontend error:", err);

// These appear in DevTools Console tab
```

### Frontend calling Go methods (debug the bridge)

```javascript
// Auto-generated bindings are in frontend/wailsjs/go/
import { GetProfiles, CreateProfile } from '../wailsjs/go/main/App';

// Debug the call
try {
    const profiles = await GetProfiles();
    console.log("Profiles:", profiles);
} catch (err) {
    console.error("Go method failed:", err);
}
```

---

## 4. Debug Build Issues

### Common build flags

```bash
# Build with debug info
wails build -debug

# Build without UPX compression (easier to debug)
wails build -noupx

# Clean build (remove cached artifacts)
wails build -clean

# Build with specific tags
wails build -tags "debug,dev"

# Check wails doctor for environment issues
wails doctor
```

### Debug CGO issues (SQLite, system libs)

```bash
# Ensure CGO is enabled
export CGO_ENABLED=1

# macOS: Ensure Xcode CLI tools installed
xcode-select --install

# Check C compiler
go env CC

# Verbose build to see CGO compilation
go build -v -x ./desktop/
```

---

## 5. Runtime Events Debugging

```go
// desktop/app.go — add lifecycle hooks for debugging

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    log.Println("[APP] startup called")
    
    // Initialize with debug logging
    mgr, err := profile.NewManager()
    if err != nil {
        log.Printf("[APP] ERROR: manager init failed: %v\n", err)
        runtime.LogError(ctx, fmt.Sprintf("Init failed: %v", err))
    }
    a.manager = mgr
    log.Println("[APP] startup complete")
}

// Add more hooks in wails.Run options:
// OnStartup:    app.startup,
// OnDomReady:   app.domReady,
// OnShutdown:   app.shutdown,
// OnBeforeClose: app.beforeClose,

func (a *App) domReady(ctx context.Context) {
    log.Println("[APP] DOM ready — frontend loaded")
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
    log.Println("[APP] beforeClose called")
    return false
}

func (a *App) shutdown(ctx context.Context) {
    log.Println("[APP] shutdown — cleanup")
}
```

---

## 6. Debug Wails Events (Go ↔ Frontend)

### Go side: Emit events to frontend

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// Emit event to frontend
runtime.EventsEmit(a.ctx, "profile:updated", profile)
runtime.EventsEmit(a.ctx, "error:occurred", err.Error())

// Listen for events from frontend
runtime.EventsOn(a.ctx, "frontend:ready", func(data ...interface{}) {
    log.Printf("[EVENT] frontend ready: %v\n", data)
})
```

### Frontend side: Listen and emit

```javascript
import { EventsOn, EventsEmit } from '../wailsjs/runtime/runtime';

// Listen for Go events
EventsOn("profile:updated", (profile) => {
    console.log("[EVENT] profile updated:", profile);
});

// Emit event to Go
EventsEmit("frontend:ready", { timestamp: Date.now() });
```

---

## 7. Common Wails Debugging Issues

### Issue: App starts but shows blank/white screen

```bash
# Check frontend build
cd desktop/frontend
npm run build
ls dist/   # Should contain index.html

# Check embed directive in main.go
# Must have: //go:embed all:frontend/dist
# NOT:       //go:embed frontend/dist

# Try dev mode to see errors
wails dev -loglevel debug
```

### Issue: Go methods not available in frontend

```bash
# Regenerate bindings
cd desktop
wails generate module

# Check generated files
ls frontend/wailsjs/go/main/

# Ensure App is bound in main.go:
# Bind: []interface{}{app},
```

### Issue: "window.go is undefined" or similar

```javascript
// The bridge requires wails runtime. In dev mode it auto-injects.
// In production, make sure wailsjs is properly imported:
import '../wailsjs/runtime/runtime';
```

### Issue: CGO / SQLite build fails

```bash
# macOS ARM64
export CGO_ENABLED=1
export CC=clang

# If using modernc.org/sqlite (pure Go, no CGO needed):
# No special flags needed — it's a Go-native SQLite

# Verify
go build -v ./desktop/
```

### Issue: App crashes on startup (no error visible)

```bash
# Run directly to see stderr:
cd desktop
go run -tags dev .

# Or build and run:
go build -tags dev -o bpm-debug .
./bpm-debug 2>&1 | tee debug.log
```

---

## 8. Production Debug

### Enable debug mode in production build

```bash
# Build with -debug flag (includes DevTools in prod)
wails build -debug

# The resulting binary will have DevTools accessible
# Useful for debugging issues that only appear in production
```

### Log to file in production

```go
import (
    "log"
    "os"
)

func setupLogging() {
    f, err := os.OpenFile("bpm-debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening log file: %v", err)
    }
    log.SetOutput(f)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    log.Println("=== BPM Desktop started ===")
}
```

---

## 9. Quick Reference

| Task | Command |
|------|---------|
| Start dev mode | `wails dev` |
| Dev with debug logs | `wails dev -loglevel debug` |
| Open DevTools | `Cmd+Option+I` (macOS) |
| Build with debug | `wails build -debug` |
| Attach debugger | `dlv attach $(pgrep bpm)` |
| Regenerate bindings | `wails generate module` |
| Check environment | `wails doctor` |
| Clean build | `wails build -clean` |
