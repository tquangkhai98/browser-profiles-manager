package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
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

		// Refuse to adopt a directory that already exists on disk: Create would
		// take it over and delete it if saving the config failed. A name that
		// is already a profile is left to Create's own "already exists" error.
		if _, err := profile.Get(name); err != nil {
			profilesDir, err := config.ProfilesDir()
			if err != nil {
				return err
			}
			dataDir := filepath.Join(profilesDir, name)
			if _, err := os.Lstat(dataDir); err == nil {
				return fmt.Errorf("directory %s already exists; move or remove it before creating profile %q", dataDir, name)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("cannot check %s: %w", dataDir, err)
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
