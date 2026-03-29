package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates and configures the bpm MCP server with all tools.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"bpm",
		"0.1.0",
	)

	registerTools(s)
	return s
}

func registerTools(s *server.MCPServer) {
	// Profile tools
	s.AddTool(
		mcp.NewTool("profile_create",
			mcp.WithDescription("Create a new isolated browser profile"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Profile name (alphanumeric, hyphens, underscores)")),
			mcp.WithString("browser", mcp.Description("Browser to use: chrome, brave, edge, arc. Auto-detected if not specified.")),
		),
		handleProfileCreate,
	)

	s.AddTool(
		mcp.NewTool("profile_list",
			mcp.WithDescription("List all browser profiles with their lock status"),
		),
		handleProfileList,
	)

	s.AddTool(
		mcp.NewTool("profile_use",
			mcp.WithDescription("Launch browser with an isolated profile. Acquires a lock."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to launch")),
			mcp.WithString("browser", mcp.Description("Override browser to use")),
		),
		handleProfileUse,
	)

	s.AddTool(
		mcp.NewTool("profile_status",
			mcp.WithDescription("Check the lock status of a profile"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to check")),
		),
		handleProfileStatus,
	)

	s.AddTool(
		mcp.NewTool("profile_delete",
			mcp.WithDescription("Delete a browser profile and its data"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to delete")),
			mcp.WithBoolean("force", mcp.Description("Force delete even if locked")),
		),
		handleProfileDelete,
	)

	// Mapping tools
	s.AddTool(
		mcp.NewTool("mapping_set",
			mcp.WithDescription("Map a project directory to a browser profile"),
			mcp.WithString("directory", mcp.Required(), mcp.Description("Project directory path")),
			mcp.WithString("profile", mcp.Required(), mcp.Description("Profile name to map to")),
		),
		handleMappingSet,
	)

	s.AddTool(
		mcp.NewTool("mapping_get",
			mcp.WithDescription("Resolve which profile is mapped to a directory"),
			mcp.WithString("directory", mcp.Required(), mcp.Description("Directory to look up")),
		),
		handleMappingGet,
	)

	// Browser detection
	s.AddTool(
		mcp.NewTool("browser_detect",
			mcp.WithDescription("List installed Chromium-based browsers on the system"),
		),
		handleBrowserDetect,
	)

	// Credential tools
	s.AddTool(
		mcp.NewTool("creds_inspect",
			mcp.WithDescription("List which sites have cookies/logins in a profile (domain + count only, never decrypts)"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to inspect")),
		),
		handleCredsInspect,
	)

	s.AddTool(
		mcp.NewTool("creds_sync",
			mcp.WithDescription("Copy credential databases from one profile to another (encrypted data stays encrypted)"),
			mcp.WithString("source", mcp.Required(), mcp.Description("Source profile name")),
			mcp.WithString("target", mcp.Required(), mcp.Description("Target profile name")),
		),
		handleCredsSync,
	)
}
