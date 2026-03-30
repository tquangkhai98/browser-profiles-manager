package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var exportJSON bool

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

		if !exportJSON {
			fmt.Printf("Exporting profile %q → %s...\n", name, dstPath)
		}

		if err := copyDir(p.DataDir, dstPath); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		if exportJSON {
			data, _ := json.MarshalIndent(map[string]any{
				"profile":     name,
				"destination": dstPath,
				"success":     true,
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Exported profile %q to %s\n", name, dstPath)
		return nil
	},
}

func init() {
	exportCmd.Flags().BoolVar(&exportJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(exportCmd)
}
