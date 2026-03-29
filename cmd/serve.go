package cmd

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	mcpserver "github.com/tquangkhai98/browser-profiles-manager/internal/mcp"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server (stdio transport)",
	Long: `Start the bpm MCP server using stdio transport.
AI IDEs (Claude Code, Cursor, Antigravity) can connect by adding bpm to their MCP config:

  {
    "mcpServers": {
      "bpm": {
        "command": "bpm",
        "args": ["serve"]
      }
    }
  }`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s := mcpserver.NewServer()

		fmt.Fprintln(cmd.ErrOrStderr(), "bpm MCP server started (stdio)")

		if err := server.ServeStdio(s); err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
