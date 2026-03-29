package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a browser profile and its data",
	Long:  `Delete a browser profile, removing it from config and deleting its data directory. Use --force to skip confirmation and delete even if locked.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Verify profile exists
		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		// Confirmation prompt unless --force
		if !deleteForce {
			fmt.Printf("Delete profile %q and all its data at %s?\n", name, p.DataDir)
			fmt.Print("Type 'yes' to confirm: ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if strings.TrimSpace(scanner.Text()) != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := profile.Delete(name, deleteForce); err != nil {
			return err
		}

		fmt.Printf("✓ Profile %q deleted\n", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation and delete even if locked")
	rootCmd.AddCommand(deleteCmd)
}
