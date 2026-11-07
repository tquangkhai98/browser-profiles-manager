package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
	"github.com/tquangkhai98/browser-profiles-manager/internal/transfer"
)

var exportJSON, exportForce bool
var exportCmd = &cobra.Command{Use: "export <profile> <destination>", Short: "Export a profile archive", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	p, err := profile.Get(args[0])
	if err != nil {
		return err
	}
	if !exportForce {
		release, e := profile.AcquireLock(p.DataDir, "export")
		if e != nil {
			return fmt.Errorf("profile %q is locked or running; run bpm stop %s before exporting (or use --force): %w", p.Name, p.Name, e)
		}
		defer release()
	}
	c, err := transfer.Export(p.DataDir, args[1], p.Name, p.Browser)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	if exportJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"profile": p.Name, "destination": args[1], "files": c.Files, "bytes": c.Bytes, "success": true})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Exported profile %q to %s\n", p.Name, args[1])
	return nil
}}

func init() {
	exportCmd.Flags().BoolVar(&exportJSON, "json", false, "Output in JSON format")
	exportCmd.Flags().BoolVar(&exportForce, "force", false, "Export even if locked")
	rootCmd.AddCommand(exportCmd)
}
