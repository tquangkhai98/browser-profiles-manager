package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/credential"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var syncJSON bool

var syncCmd = &cobra.Command{
	Use:   "sync <source> <target>",
	Short: "Sync credentials between profiles",
	Long:  `Copy cookie and login databases from one profile to another. Data remains encrypted — bpm never decrypts credentials.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcName, dstName := args[0], args[1]

		src, err := profile.Get(srcName)
		if err != nil {
			return fmt.Errorf("source profile: %w", err)
		}

		dst, err := profile.Get(dstName)
		if err != nil {
			return fmt.Errorf("target profile: %w", err)
		}

		if !syncJSON {
			fmt.Printf("Syncing credentials: %s → %s\n", srcName, dstName)
		}

		copied, err := credential.Sync(src.DataDir, dst.DataDir)
		if err != nil {
			return err
		}

		if syncJSON {
			data, _ := json.MarshalIndent(map[string]any{
				"source":       srcName,
				"target":       dstName,
				"files_copied": copied,
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Synced %d credential file(s)\n", copied)
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(syncCmd)
}
