package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/mapping"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
	browserruntime "github.com/tquangkhai98/browser-profiles-manager/internal/runtime"
)

func handleProfileCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	browserID := request.GetString("browser", "")
	if browserID == "" {
		b, err := browser.DefaultBrowser()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		browserID = b.ID
	} else if _, err := browser.FindBrowser(browserID); err != nil {
		// Same validation as `bpm create --browser`.
		return mcp.NewToolResultError(err.Error()), nil
	}

	p, err := profile.Create(name, browserID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, _ := json.Marshal(map[string]any{
		"name":       p.Name,
		"browser":    p.Browser,
		"data_dir":   p.DataDir,
		"created_at": p.CreatedAt,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func handleProfileList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profiles, err := profile.List()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := make([]map[string]any, 0, len(profiles))
	for _, p := range profiles {
		data, _ := json.Marshal(p)
		var item map[string]any
		_ = json.Unmarshal(data, &item)
		state, err := browserruntime.Read(p.Name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		item["running"] = state != nil
		if state != nil {
			item["pid"] = state.PID
			item["port"] = state.Port
			item["cdp_http"] = state.CDPHTTP
			item["webSocketDebuggerUrl"] = state.WebSocket
			item["headless"] = state.Headless
		}
		result = append(result, item)
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleProfileUse(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	headless := request.GetBool("headless", browser.ResolveHeadless(false, false, os.Getenv("BPM_HEADLESS"), browser.HasDisplay()))
	state, err := browserruntime.Launch(name, browserruntime.Options{Browser: request.GetString("browser", ""), Headless: headless, Port: request.GetInt("cdp_port", 0)})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result := map[string]any{"profile": state.Profile, "browser": state.Browser, "pid": state.PID, "port": state.Port, "cdp_http": state.CDPHTTP, "webSocketDebuggerUrl": state.WebSocket, "headless": state.Headless}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleProfileStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}
	stopped, err := browserruntime.Stop(name, 5*time.Second)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	result := map[string]any{"profile": name, "stopped": stopped}
	if !stopped {
		result["reason"] = "not running"
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func handleProfileStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	p, err := profile.Get(name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	locked, lockInfo := profile.IsLocked(p.DataDir)

	status := map[string]any{
		"name":    p.Name,
		"browser": p.Browser,
		"locked":  locked,
	}
	if lockInfo != nil {
		status["lock_info"] = lockInfo
	}

	result, _ := json.MarshalIndent(status, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func handleProfileDelete(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	force := request.GetBool("force", false)

	if err := profile.Delete(name, force); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Profile %q deleted", name)), nil
}

func handleMappingSet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir, err := request.RequireString("directory")
	if err != nil {
		return mcp.NewToolResultError("directory is required"), nil
	}
	// A relative path would be resolved against the MCP server's working
	// directory, which is unrelated to the client's project.
	if !filepath.IsAbs(dir) {
		return mcp.NewToolResultError(fmt.Sprintf("directory must be an absolute path, got %q", dir)), nil
	}

	profileName, err := request.RequireString("profile")
	if err != nil {
		return mcp.NewToolResultError("profile is required"), nil
	}

	if err := mapping.Set(dir, profileName); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Mapped %s → %s", dir, profileName)), nil
}

func handleMappingGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir, err := request.RequireString("directory")
	if err != nil {
		return mcp.NewToolResultError("directory is required"), nil
	}
	// A relative path would be resolved against the MCP server's working
	// directory, which is unrelated to the client's project.
	if !filepath.IsAbs(dir) {
		return mcp.NewToolResultError(fmt.Sprintf("directory must be an absolute path, got %q", dir)), nil
	}

	profileName, err := mapping.Get(dir)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if profileName == "" {
		return mcp.NewToolResultText(fmt.Sprintf("No profile mapped for %s", dir)), nil
	}

	result, _ := json.Marshal(map[string]string{
		"directory": dir,
		"profile":   profileName,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func handleBrowserDetect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	browsers := browser.DetectBrowsers()
	result, _ := json.MarshalIndent(browsers, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func handleCredsInspect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	p, err := profile.Get(name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	inspectResult, err := credential.Inspect(p.DataDir, p.Name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, _ := json.MarshalIndent(inspectResult, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func handleCredsSync(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, err := request.RequireString("source")
	if err != nil {
		return mcp.NewToolResultError("source is required"), nil
	}

	target, err := request.RequireString("target")
	if err != nil {
		return mcp.NewToolResultError("target is required"), nil
	}

	src, err := profile.Get(source)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("source profile: %s", err)), nil
	}

	dst, err := profile.Get(target)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("target profile: %s", err)), nil
	}

	copied, err := credential.Sync(src.DataDir, dst.DataDir)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Synced %d credential file(s) from %s to %s", copied, source, target)), nil
}
