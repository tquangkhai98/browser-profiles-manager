package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/config"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
	"github.com/tquangkhai98/browser-profiles-manager/internal/transfer"
)

var importName string
var importJSON bool
var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func deriveName(src string) string {
	n := strings.ToLower(filepath.Base(filepath.Clean(src)))
	for _, ext := range []string{".tar.gz", ".tgz", ".zip"} {
		if strings.HasSuffix(n, ext) {
			n = strings.TrimSuffix(n, ext)
			break
		}
	}
	return strings.Trim(nonAlphanumeric.ReplaceAllString(n, "-"), "-")
}

var importCmd = &cobra.Command{Use: "import <path> [name]", Short: "Import a Chrome profile", Args: cobra.RangeArgs(1, 2), RunE: func(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := deriveName(args[0])
	if len(args) > 1 {
		name = args[1]
	}
	if cmd.Flags().Changed("name") {
		name = importName
	}
	dir, err := config.ProfilesDir()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(dir, ".bpm-import-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	counts, err := transfer.Unpack(args[0], staging)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}
	if err = transfer.Normalize(staging); err != nil {
		return fmt.Errorf("import failed: %w", err)
	}
	p, err := profile.Create(name, "chrome")
	if err != nil {
		return err
	}
	success := false
	defer func() {
		if !success {
			_ = profile.Delete(name, true)
		}
	}()
	if err = os.Remove(p.DataDir); err != nil {
		return err
	}
	if err = os.Rename(staging, p.DataDir); err != nil {
		return err
	}
	success = true
	if importJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"profile": name, "source": args[0], "data_dir": p.DataDir, "files": counts.Files, "bytes": counts.Bytes, "success": true})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Imported profile %q from %s\n", name, args[0])
	return nil
}}

func init() {
	importCmd.Flags().StringVar(&importName, "name", "", "Profile name")
	importCmd.Flags().BoolVar(&importJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(importCmd)
}
