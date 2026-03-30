package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/install"
)

var installList bool
var installUninstall bool

var installCmd = &cobra.Command{
	Use:   "install [ide]",
	Short: "Install bpm MCP server into an AI IDE",
	Long: `Install bpm as an MCP server into supported AI IDEs.

Supported IDEs:
  claude-code      Claude Code CLI (~/.claude/mcp_servers.json)
  claude-desktop   Claude Desktop App
  cursor           Cursor IDE (~/.cursor/mcp.json)
  antigravity      Antigravity IDE (~/.gemini/settings.json)

Examples:
  bpm install claude-code     # Add bpm to Claude Code
  bpm install cursor          # Add bpm to Cursor
  bpm install --list          # Show all IDEs and status
  bpm install --uninstall cursor  # Remove bpm from Cursor`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// List mode
		if installList || len(args) == 0 {
			return listIDEs()
		}

		ideID := args[0]

		// Uninstall mode
		if installUninstall {
			result, err := install.Uninstall(ideID)
			if err != nil {
				return err
			}
			fmt.Println(result.Message)
			if result.Action == "uninstalled" {
				fmt.Printf("  Config: %s\n", result.ConfigPath)
			}
			return nil
		}

		// Install mode
		result, err := install.Install(ideID)
		if err != nil {
			return err
		}
		fmt.Println(result.Message)
		if result.Action == "installed" {
			fmt.Printf("  Config: %s\n", result.ConfigPath)
			fmt.Println("  Restart your IDE to activate.")
		}
		return nil
	},
}

func listIDEs() error {
	ides := install.SupportedIDEs()

	fmt.Println("Supported AI IDEs:")
	fmt.Println()
	for _, ide := range ides {
		status := "  ○"
		if ide.BPMEnabled {
			status = "  ●"
		}
		installed := ""
		if !ide.Installed {
			installed = " (not detected)"
		}
		fmt.Printf("%s %-18s %s%s\n", status, ide.Name, ide.ConfigPath, installed)
	}
	fmt.Println()
	fmt.Printf("  ● = bpm configured    ○ = not configured\n")
	fmt.Println()
	fmt.Println("Usage: bpm install <ide-id>")
	return nil
}

func init() {
	installCmd.Flags().BoolVar(&installList, "list", false, "List all supported IDEs and their status")
	installCmd.Flags().BoolVar(&installUninstall, "uninstall", false, "Remove bpm from the specified IDE")
	rootCmd.AddCommand(installCmd)
}
