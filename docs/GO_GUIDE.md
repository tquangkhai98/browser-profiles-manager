# Go Quick Guide — Dành cho developer mới

> Hướng dẫn cơ bản về Go, tập trung vào các pattern được sử dụng trong project **bpm**.

---

## Table of Contents

- [1. Setup](#1-setup)
- [2. Go trong 5 phút](#2-go-trong-5-phút)
- [3. Struct & Method](#3-struct--method)
- [4. Error Handling](#4-error-handling)
- [5. Package & Import](#5-package--import)
- [6. Interface](#6-interface)
- [7. Goroutine & Concurrency](#7-goroutine--concurrency)
- [8. Go Module System](#8-go-module-system)
- [9. Pattern trong project bpm](#9-pattern-trong-project-bpm)
- [10. Cheat Sheet](#10-cheat-sheet)

---

## 1. Setup

```bash
# Install Go (macOS)
brew install go

# Verify
go version
# → go version go1.25.5 darwin/arm64

# GOPATH — nơi Go lưu trữ binaries & packages
echo $GOPATH
# → ~/go (default)

# Đảm bảo GOPATH/bin trong PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Editor Setup

- **VS Code**: Cài extension "Go" (by Go Team at Google) → tự động format, autocomplete, linting.
- **GoLand** (JetBrains): Full IDE, có sẵn mọi thứ.

---

## 2. Go trong 5 phút

### Hello World

```go
package main          // Mỗi file Go thuộc 1 package. Chương trình khởi chạy từ "package main"

import "fmt"          // Import standard library package

func main() {         // Entry point — giống main() trong C/Java
    fmt.Println("Hello, World!")
}
```

### Biến & Kiểu dữ liệu

```go
// Khai báo dài (explicit type)
var name string = "bpm"
var count int = 42
var active bool = true

// Khai báo ngắn (short declaration — phổ biến nhất)
name := "bpm"         // Go tự infer kiểu dữ liệu
count := 42
active := true

// Constants
const version = "0.1.0"
const maxProfiles = 100
```

### Kiểu dữ liệu phổ biến

```go
// Basic types
string      // "hello"
int         // 42
float64     // 3.14
bool        // true / false
byte        // alias of uint8

// Composite types
[]string            // slice (dynamic array) — ["a", "b", "c"]
map[string]int      // map (dictionary)      — {"age": 25}
[3]int              // array (fixed size)     — [1, 2, 3]
```

### Slice & Map

```go
// Slice (dùng rất nhiều, giống array nhưng dynamic)
profiles := []string{"work", "personal", "staging"}
profiles = append(profiles, "new-one")    // Thêm phần tử
fmt.Println(profiles[0])                   // → "work"
fmt.Println(len(profiles))                 // → 4

// Map
settings := map[string]string{
    "browser": "chrome",
    "theme":   "dark",
}
settings["language"] = "en"               // Thêm key
value, exists := settings["browser"]      // Check tồn tại
delete(settings, "theme")                 // Xóa key
```

### If / For / Switch

```go
// If — KHÔNG cần parentheses ()
if count > 10 {
    fmt.Println("nhiều")
} else if count > 5 {
    fmt.Println("vừa")
} else {
    fmt.Println("ít")
}

// If với short statement (rất phổ biến trong Go)
if err := doSomething(); err != nil {
    fmt.Println("Lỗi:", err)
}

// For — Go chỉ có FOR, không có while
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// For range (iterate slice/map)
for index, name := range profiles {
    fmt.Printf("%d: %s\n", index, name)
}

// For vô hạn (như while true)
for {
    // do something forever
    break  // thoát khi cần
}

// Switch
switch browser {
case "chrome":
    fmt.Println("Chromium-based")
case "firefox":
    fmt.Println("Mozilla")
default:
    fmt.Println("Unknown")
}
```

### Hàm (Function)

```go
// Hàm cơ bản
func greet(name string) string {
    return "Hello, " + name
}

// Multiple return values (ĐẶC TRƯNG CỦA GO)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide by zero")
    }
    return a / b, nil    // nil = no error
}

// Gọi hàm
result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)

// Hàm không return
func logMessage(msg string) {
    fmt.Println("[LOG]", msg)
}
```

---

## 3. Struct & Method

Struct là cách Go thay thế class trong OOP.

```go
// Định nghĩa struct
type Profile struct {
    Name      string    `json:"name"`       // json tag cho serialization
    Browser   string    `json:"browser"`
    CreatedAt time.Time `json:"created_at"`
    IsLocked  bool      `json:"is_locked"`
}

// Tạo instance
p := Profile{
    Name:    "work-staging",
    Browser: "chrome",
}

// Truy cập field
fmt.Println(p.Name)     // → "work-staging"
p.IsLocked = true

// Pointer to struct (phổ biến)
p2 := &Profile{Name: "personal"}
```

### Method (receiver function)

```go
// Value receiver — KHÔNG thay đổi struct gốc
func (p Profile) DisplayName() string {
    return fmt.Sprintf("[%s] %s", p.Browser, p.Name)
}

// Pointer receiver — CÓ THỂ thay đổi struct gốc
func (p *Profile) Lock() {
    p.IsLocked = true
}

// Gọi method
profile := Profile{Name: "work", Browser: "chrome"}
fmt.Println(profile.DisplayName())   // → "[chrome] work"
profile.Lock()
fmt.Println(profile.IsLocked)        // → true
```

> **Quy ước Go**: Nếu method cần modify struct → dùng pointer receiver `*T`.
> Nếu chỉ read → dùng value receiver `T`.
> Trong practice, hầu hết dùng pointer receiver cho consistency.

---

## 4. Error Handling

Go KHÔNG có try/catch. Thay vào đó, lỗi được trả về như một giá trị bình thường.

```go
// Pattern cơ bản — THẤY RẤT NHIỀU TRONG CODE GO
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do X: %w", err)  // Wrap error
}
// tiếp tục xử lý result...

// Tạo error
import "errors"
var ErrNotFound = errors.New("profile not found")

// Dùng fmt.Errorf để tạo error với context
func getProfile(name string) (*Profile, error) {
    profile, exists := profiles[name]
    if !exists {
        return nil, fmt.Errorf("profile %q not found", name)
    }
    return profile, nil
}

// Error wrapping với %w (Go 1.13+)
if err := saveConfig(); err != nil {
    return fmt.Errorf("saving config: %w", err)
}

// Check wrapped error
if errors.Is(err, ErrNotFound) {
    fmt.Println("Profile does not exist")
}
```

### Pattern trong project bpm

```go
// cmd/create.go — ví dụ thực tế
RunE: func(cmd *cobra.Command, args []string) error {
    name := args[0]
    mgr := profile.NewManager()

    if err := mgr.Create(name, browser); err != nil {
        return fmt.Errorf("failed to create profile: %w", err)
    }

    fmt.Printf("✅ Profile %q created\n", name)
    return nil
},
```

---

## 5. Package & Import

### Cách Go tổ chức code

```
project/
├── main.go           // package main
├── cmd/              // package cmd
│   ├── root.go
│   └── create.go
└── internal/         // private packages
    └── profile/      // package profile
        └── manager.go
```

```go
// internal/profile/manager.go
package profile                    // Package được xác định bởi folder name

type Manager struct { ... }        // Exported (viết HOA chữ đầu)
type config struct { ... }         // Unexported (viết thường chữ đầu)

func NewManager() *Manager { ... } // Exported function
func helper() { ... }              // Unexported function
```

> **Quy tắc quan trọng nhất**:
> - **Viết HOA chữ đầu** → Public (exported) — các package khác có thể truy cập
> - **Viết thường chữ đầu** → Private (unexported) — chỉ dùng trong package hiện tại

### Import

```go
import (
    // Standard library
    "fmt"
    "os"
    "path/filepath"

    // Third-party (tự động theo module path)
    "github.com/spf13/cobra"

    // Internal project packages (dùng module path đầy đủ)
    "github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)
```

### `internal/` — thư mục đặc biệt

Go có rule: các package trong `internal/` **chỉ có thể được import bởi code trong parent directory** của `internal/`. Không ai bên ngoài project có thể dùng.

---

## 6. Interface

Interface trong Go là **implicit** — không cần khai báo `implements`.

```go
// Định nghĩa interface
type ProfileStore interface {
    Save(profile *Profile) error
    Load(name string) (*Profile, error)
    Delete(name string) error
}

// BẤT KỲ struct nào có đủ 3 method trên → tự động thỏa mãn interface
type FileStore struct {
    basePath string
}

func (fs *FileStore) Save(profile *Profile) error { ... }
func (fs *FileStore) Load(name string) (*Profile, error) { ... }
func (fs *FileStore) Delete(name string) error { ... }

// FileStore tự động implement ProfileStore — KHÔNG cần viết "implements"!

// Sử dụng
func NewManager(store ProfileStore) *Manager {
    return &Manager{store: store}
}

// Có thể truyền bất kỳ implementation nào
mgr := NewManager(&FileStore{basePath: "/etc/bpm"})
```

### Interface phổ biến trong Go standard library

```go
// io.Reader — read bytes
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer — write bytes
type Writer interface {
    Write(p []byte) (n int, err error)
}

// error — mọi error đều là interface này
type error interface {
    Error() string
}

// fmt.Stringer — giống toString()
type Stringer interface {
    String() string
}
```

---

## 7. Goroutine & Concurrency

Goroutine là lightweight thread do Go runtime quản lý.

```go
// Chạy goroutine — chỉ cần thêm keyword "go"
go doSomething()

// Ví dụ thực tế
func scanAllBrowsers() []Browser {
    browsers := []string{"chrome", "edge", "brave"}
    results := make(chan Browser, len(browsers))

    for _, name := range browsers {
        go func(browser string) {
            b, err := detect(browser)
            if err == nil {
                results <- b    // Gửi kết quả vào channel
            }
        }(name)
    }

    // Thu thập kết quả
    var found []Browser
    for i := 0; i < len(browsers); i++ {
        found = append(found, <-results)    // Nhận từ channel
    }
    return found
}
```

### Channel

```go
// Channel = ống truyền data giữa các goroutine
ch := make(chan string)         // Unbuffered channel
ch := make(chan string, 10)     // Buffered channel (queue size 10)

// Gửi & nhận
ch <- "hello"                  // Gửi vào channel
msg := <-ch                    // Nhận từ channel (blocking)

// Select — multiplex channels
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
case <-time.After(5 * time.Second):
    fmt.Println("timeout!")
}
```

---

## 8. Go Module System

### `go.mod` — giống `package.json` cho Node.js

```
module github.com/tquangkhai98/browser-profiles-manager

go 1.25.5

require (
    github.com/spf13/cobra v1.10.2            // CLI framework
    github.com/mark3labs/mcp-go v0.46.0        // MCP protocol
    github.com/wailsapp/wails/v2 v2.12.0       // Desktop GUI
    modernc.org/sqlite v1.48.0                 // SQLite driver
)
```

### `go.sum` — giống `package-lock.json`

Lock file chứa checksum của tất cả dependencies. **KHÔNG cần đọc hay edit**.

### So sánh với Node.js

| Go | Node.js | Mô tả |
|----|---------|--------|
| `go.mod` | `package.json` | Khai báo dependencies |
| `go.sum` | `package-lock.json` | Lock file |
| `go mod tidy` | `npm install` | Sync dependencies |
| `go get pkg` | `npm install pkg` | Add dependency |
| `go build` | `npm run build` | Build project |
| `go run .` | `node index.js` | Run directly |
| `go test` | `npm test` | Run tests |
| `$GOPATH/bin` | `node_modules/.bin` | Installed binaries |

---

## 9. Pattern trong project bpm

### Cobra CLI Pattern

Project dùng [Cobra](https://github.com/spf13/cobra) cho CLI. Mỗi command là 1 file trong `cmd/`:

```go
// cmd/create.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var browser string    // Flag variable

var createCmd = &cobra.Command{
    Use:   "create <name>",                       // Cách dùng
    Short: "Create a new isolated browser profile", // Mô tả ngắn
    Args:  cobra.ExactArgs(1),                    // Yêu cầu đúng 1 argument
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        mgr := profile.NewManager()

        if err := mgr.Create(name, browser); err != nil {
            return fmt.Errorf("failed to create profile: %w", err)
        }
        fmt.Printf("✅ Profile %q created\n", name)
        return nil
    },
}

func init() {
    // init() chạy TỰ ĐỘNG khi package được import
    createCmd.Flags().StringVarP(&browser, "browser", "b", "chrome",
        "Target browser (chrome, edge, brave)")
    rootCmd.AddCommand(createCmd)
}
```

### Wails Desktop Pattern

Project dùng [Wails v2](https://wails.io) cho desktop app. Go code expose method cho frontend qua `Bind`:

```go
// desktop/app.go
type App struct {
    ctx     context.Context
    manager *profile.Manager
}

// Method exported ra frontend JavaScript
func (a *App) GetProfiles() ([]profile.Profile, error) {
    return a.manager.ListAll()
}

// Frontend gọi qua auto-generated bridge:
//   const profiles = await GetProfiles()
```

### MCP Server Pattern

```go
// internal/mcp/server.go
func NewServer() *mcpServer {
    s := server.NewMCPServer("bpm", version)

    s.AddTool(mcp.NewTool("profile_create", ...),
        func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            // Handle tool call
        })

    return &mcpServer{server: s}
}
```

---

## 10. Cheat Sheet

### Lệnh Go thường dùng

```bash
go run .                    # Chạy nhanh (không tạo binary)
go build -o bpm .           # Build binary
go test ./...               # Test tất cả
go mod tidy                 # Sync dependencies
go fmt ./...                # Format code (giống prettier)
go vet ./...                # Lint check
go doc fmt.Println          # Xem documentation
```

### Printf Format Verbs

```go
fmt.Printf("%s", str)       // String
fmt.Printf("%d", num)       // Integer
fmt.Printf("%f", float)     // Float
fmt.Printf("%v", anything)  // Default format (mọi type)
fmt.Printf("%+v", struct)   // Struct với field names
fmt.Printf("%#v", struct)   // Go syntax representation
fmt.Printf("%T", anything)  // Type name
fmt.Printf("%q", str)       // Quoted string — "hello"
fmt.Printf("%p", ptr)       // Pointer address
fmt.Printf("%t", bool)      // Boolean — true/false
```

### Zero Values (default khi khai báo biến)

```go
int     → 0
float64 → 0.0
string  → ""
bool    → false
pointer → nil
slice   → nil
map     → nil
```

### nil trong Go

```go
// nil = "nothing" (giống null/undefined trong JS)
// Chỉ dùng cho: pointers, interfaces, maps, slices, channels, functions

var p *Profile = nil     // Pointer chưa trỏ đến đâu
var s []string = nil     // Slice chưa khởi tạo

// LUÔN check nil trước khi truy cập
if p != nil {
    fmt.Println(p.Name)
}
```

### Pointer Cơ Bản

```go
name := "hello"
ptr := &name        // & = lấy address → ptr là *string
value := *ptr       // * = lấy value tại address → "hello"

// Tại sao dùng pointer?
// 1. Truyền reference thay vì copy (hiệu quả hơn cho struct lớn)
// 2. Cho phép function modify giá trị gốc
// 3. Cho phép nil (biểu diễn "không có giá trị")

func updateName(p *Profile, newName string) {
    p.Name = newName   // Thay đổi trực tiếp struct gốc
}
```

---

## Tài liệu tham khảo

| Resource | Link |
|----------|------|
| Go Tour (interactive) | https://go.dev/tour |
| Go by Example | https://gobyexample.com |
| Effective Go | https://go.dev/doc/effective_go |
| Go Playground | https://go.dev/play |
| Go Standard Library | https://pkg.go.dev/std |
| Cobra CLI Docs | https://cobra.dev |
| Wails v2 Docs | https://wails.io/docs |

---

> 💡 **Tip**: Cách nhanh nhất để học Go là đọc code trong thư mục `cmd/` và `internal/` của project này, rồi thử modify + chạy `go run . <command>` để xem kết quả.
