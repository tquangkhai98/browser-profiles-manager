package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var statusCmd = &cobra.Command{
	Use:   "status <name>",
	Short: "Show profile lock status",
	Long:  `Show detailed lock status for a profile, including the PID and caller that holds the lock.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		locked, lockInfo := profile.IsLocked(p.DataDir)

		fmt.Printf("Profile: %s\n", p.Name)
		fmt.Printf("Browser: %s\n", p.Browser)
		fmt.Printf("DataDir: %s\n", p.DataDir)
		fmt.Printf("Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))

		if p.LastUsed != nil {
			fmt.Printf("Last Used: %s\n", p.LastUsed.Format("2006-01-02 15:04:05"))
		}

		if locked {
			fmt.Printf("\nStatus: 🔒 Locked\n")
			fmt.Printf("  PID:       %d\n", lockInfo.PID)
			fmt.Printf("  Caller:    %s\n", lockInfo.Caller)
			fmt.Printf("  Locked at: %s\n", lockInfo.LockedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("\nStatus: 🟢 Free\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
