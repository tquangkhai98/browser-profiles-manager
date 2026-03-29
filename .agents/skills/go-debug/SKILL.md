---
name: go-debug
description: Debug Go applications using Delve (dlv). Covers breakpoints, stepping, variable inspection, goroutine debugging, conditional breakpoints, and common debug workflows for CLI and library code.
---

# Go Debug Skill — Using Delve (`dlv`)

> This skill provides instructions for debugging Go applications using the Delve debugger.

## Prerequisites

```bash
# Delve must be installed
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify
dlv version
```

---

## 1. Debug a Go CLI Application

### Start debug session for a CLI binary

```bash
# Debug the main package (equivalent to `go run .` but with debugger)
dlv debug .

# Debug with CLI arguments
dlv debug . -- create my-profile --browser chrome

# Debug a specific package/entry point
dlv debug ./cmd/bpm -- list

# Debug a specific main package in a subdirectory
dlv debug ./desktop -- 
```

### Build with debug symbols (no optimization)

```bash
# Build binary with full debug info (disables optimizations + inlining)
go build -gcflags="all=-N -l" -o app-debug .

# Then debug the compiled binary
dlv exec ./app-debug -- <args>
```

---

## 2. Core Delve Commands

Once inside a `dlv` session, use these commands:

### Breakpoints

```
# Set breakpoint by function name
break main.main
b main.main

# Set breakpoint by file:line
break cmd/create.go:25
b internal/profile/manager.go:42

# Set breakpoint by package.function
break github.com/tquangkhai98/browser-profiles-manager/internal/profile.(*Manager).Create

# Conditional breakpoint
break cmd/create.go:30
condition 1 name == "test-profile"

# List all breakpoints
breakpoints
bp

# Clear a breakpoint
clear 1
clearall
```

### Execution Control

```
# Continue execution until next breakpoint
continue
c

# Step over (next line, don't enter functions)
next
n

# Step into (enter function call)
step
s

# Step out (finish current function, return to caller)
stepout
so

# Restart the program
restart
r
```

### Inspecting Variables

```
# Print a variable
print name
p name

# Print with type info
whatis name

# Print struct with all fields
print *profile
p profile.Name

# Print a slice/map
print profiles
print settings["browser"]

# Print all local variables
locals

# Print function arguments
args

# Print all goroutine-local variables
vars

# Evaluate an expression
print len(profiles)
print err != nil
print fmt.Sprintf("name=%s", name)
```

### Call Stack

```
# Show current call stack
stack
bt

# Show stack with N frames
stack 20

# Switch to a different stack frame
frame 2

# Show goroutines
goroutines
goroutine 1
```

### Watchpoints (data breakpoints)

```
# Break when a variable changes value (Go 1.22+ / Delve 1.22+)
watch profile.IsLocked
watch -w write profile.Name
```

---

## 3. Debug Unit Tests

```bash
# Debug a specific test in a package
dlv test ./internal/profile -- -test.run TestCreateProfile

# Debug all tests in a package
dlv test ./internal/profile

# Debug tests with verbose output
dlv test ./internal/profile -- -test.v

# Debug a bench
dlv test ./internal/profile -- -test.bench BenchmarkCreate
```

---

## 4. Attach to Running Process

```bash
# Find the PID
pgrep -f bpm
# or
ps aux | grep bpm

# Attach debugger to running process
dlv attach <PID>

# Attach to a Wails dev app (useful for desktop debugging)
# First start: wails dev
# Then: dlv attach $(pgrep -f bpm-desktop)
```

---

## 5. Remote / Headless Debugging (for IDE integration)

```bash
# Start Delve in headless mode (DAP protocol for VS Code)
dlv dap --listen=:2345 --log

# Start Delve headless server (legacy JSON-RPC for GoLand)
dlv debug --headless --listen=:2345 --api-version=2 --log . -- <args>

# Connect from another terminal
dlv connect :2345
```

### VS Code launch.json configuration

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug CLI",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["list"]
    },
    {
      "name": "Debug Wails Desktop",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/desktop",
      "args": [],
      "buildFlags": "-tags dev"
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": 0
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "args": ["-test.run", "${selectedText}"]
    }
  ]
}
```

---

## 6. Common Debug Scenarios

### Scenario: Debug a Cobra CLI command

```bash
# Example: debug `bpm create test-profile --browser chrome`
dlv debug . -- create test-profile --browser chrome

# Inside dlv:
b cmd/create.go:25        # Break at RunE
c                          # Continue to breakpoint
p name                     # Print the profile name arg
p browser                  # Print the --browser flag value
n                          # Step through
```

### Scenario: Debug error handling

```bash
dlv debug . -- create existing-profile

# Inside dlv:
b internal/profile/manager.go:50   # Break at Create()
c
n                                   # Step until error path
p err                               # Inspect the error
p fmt.Sprintf("%+v", err)           # Full error chain
```

### Scenario: Inspect struct state

```bash
# Inside dlv at a breakpoint:
p *manager                # Print entire Manager struct
p manager.store            # Print store field
p profile.CreatedAt.Format("2006-01-02")  # Evaluate method
```

### Scenario: Debug goroutine / concurrency issues

```bash
# Inside dlv:
goroutines                  # List all goroutines
goroutine 5                 # Switch to goroutine 5
bt                          # Show stack for that goroutine
locals                      # Show local vars in that goroutine
goroutine 1                 # Switch back to main
```

---

## 7. Quick Debug with Print (when Delve is overkill)

Sometimes `fmt.Println` / `log.Printf` is faster for simple debugging:

```go
import "log"

// Quick debug print
log.Printf("[DEBUG] profile=%+v\n", profile)
log.Printf("[DEBUG] err=%v, name=%q\n", err, name)

// Print call stack
import "runtime/debug"
debug.PrintStack()

// Print type of a variable
log.Printf("[DEBUG] type=%T value=%v\n", x, x)
```

### Using `go vet` and `staticcheck` for static analysis

```bash
# Built-in vet (catches common mistakes)
go vet ./...

# Staticcheck (more thorough analysis)
staticcheck ./...
```

---

## 8. Useful Delve Shortcuts

| Command | Short | Description |
|---------|-------|-------------|
| `break` | `b` | Set breakpoint |
| `continue` | `c` | Continue execution |
| `next` | `n` | Step over |
| `step` | `s` | Step into |
| `stepout` | `so` | Step out of function |
| `print` | `p` | Print variable |
| `locals` | | Show local variables |
| `args` | | Show function arguments |
| `stack` | `bt` | Show call stack |
| `goroutines` | `grs` | List goroutines |
| `restart` | `r` | Restart program |
| `exit` | `q` | Quit debugger |
| `help` | `h` | Show help |

---

## 9. Troubleshooting

### "could not launch process: debugserver or lldb-server not found"
```bash
# macOS: Install Xcode command line tools
xcode-select --install
```

### "could not attach to pid: permission denied"
```bash
# macOS: Requires developer mode or codesigning
sudo dlv attach <PID>
# Or sign dlv binary (persistent fix):
# See: https://github.com/go-delve/delve/blob/master/Documentation/installation/osx/install.md
```

### Optimized binary (variables show as <optimized out>)
```bash
# Always build with debug flags:
go build -gcflags="all=-N -l" -o app .
dlv exec ./app
```
