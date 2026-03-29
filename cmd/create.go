package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var createBrowser string

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new isolated browser profile",
	Long:  `Create a new isolated browser profile with its own data directory. Each profile gets a separate --user-data-dir for complete browser isolation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Resolve browser: flag → config default → auto-detect
		browserID := createBrowser
		if browserID == "" {
			b, err := browser.DefaultBrowser()
			if err != nil {
				return err
			}
			browserID = b.ID
		} else {
			// Validate the browser exists
			if _, err := browser.FindBrowser(browserID); err != nil {
				return err
			}
		}

		p, err := profile.Create(name, browserID)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Profile %q created\n", p.Name)
		fmt.Printf("  Browser:  %s\n", p.Browser)
		fmt.Printf("  Data dir: %s\n", p.DataDir)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createBrowser, "browser", "", "Browser to use (chrome, brave, edge, arc). Auto-detected if not specified.")
	rootCmd.AddCommand(createCmd)
}
