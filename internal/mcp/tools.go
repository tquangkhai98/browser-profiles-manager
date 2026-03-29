package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/mapping"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
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

	result, _ := json.MarshalIndent(profiles, "", "  ")
	return mcp.NewToolResultText(string(result)), nil
}

func handleProfileUse(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}

	p, err := profile.Get(name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	browserID := request.GetString("browser", p.Browser)
	if browserID == "" {
		b, err := browser.DefaultBrowser()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		browserID = b.ID
	}

	b, err := browser.FindBrowser(browserID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Acquire lock
	releaseLock, err := profile.AcquireLock(p.DataDir, "mcp-server")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("cannot lock profile: %s", err)), nil
	}

	// Launch browser
	_, err = browser.Launch(b.ExePath, p.DataDir)
	if err != nil {
		releaseLock()
		return mcp.NewToolResultError(err.Error()), nil
	}

	_ = profile.UpdateLastUsed(name)

	return mcp.NewToolResultText(fmt.Sprintf("Launched %s with profile %q. Lock acquired (caller: mcp-server).", b.Name, name)), nil
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
