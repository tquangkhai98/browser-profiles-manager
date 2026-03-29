package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var exportCmd = &cobra.Command{
	Use:   "export <name> <path>",
	Short: "Export a bpm profile for backup",
	Long:  `Export a bpm profile's data directory to a specified path for backup or transfer.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, dstPath := args[0], args[1]

		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		fmt.Printf("Exporting profile %q → %s...\n", name, dstPath)
		if err := copyDir(p.DataDir, dstPath); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		fmt.Printf("✓ Exported profile %q to %s\n", name, dstPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
