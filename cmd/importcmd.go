package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/fsutil"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var importCmd = &cobra.Command{
	Use:   "import <path> <name>",
	Short: "Import an existing Chrome profile into bpm",
	Long:  `Import an existing Chromium profile directory (e.g., Chrome's Default profile) into bpm as a new managed profile.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, name := args[0], args[1]

		// Verify source exists
		info, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("source path %q not found: %w", srcPath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("source path %q is not a directory", srcPath)
		}

		// Create the profile first (validates name, creates dir)
		p, err := profile.Create(name, "chrome")
		if err != nil {
			return err
		}

		// Copy contents from source to profile's data dir
		fmt.Printf("Importing %s → %s...\n", srcPath, p.DataDir)
		if err := fsutil.CopyDir(srcPath, p.DataDir); err != nil {
			// Cleanup on failure
			profile.Delete(name, true)
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("✓ Imported profile %q from %s\n", name, srcPath)
		return nil
	},
}

func init() {
	// "import" is a Go keyword, so the variable is named importCmd
	rootCmd.AddCommand(importCmd)
}
