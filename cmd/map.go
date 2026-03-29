package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/mapping"
)

var (
	mapList   bool
	mapAuto   bool
	mapRemove string
)

var mapCmd = &cobra.Command{
	Use:   "map [<directory> <profile>]",
	Short: "Map project directories to browser profiles",
	Long: `Map project directories to browser profiles for automatic resolution.
  bpm map <dir> <profile>  — Map a directory to a profile
  bpm map --auto           — Auto-resolve profile for current directory
  bpm map --list           — Show all mappings
  bpm map --remove <dir>   — Remove a mapping`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mapList {
			return runMapList()
		}
		if mapAuto {
			return runMapAuto()
		}
		if mapRemove != "" {
			return runMapRemove(mapRemove)
		}
		if len(args) == 2 {
			return runMapSet(args[0], args[1])
		}

		return cmd.Help()
	},
}

func runMapSet(dir, profileName string) error {
	if err := mapping.Set(dir, profileName); err != nil {
		return err
	}
	fmt.Printf("✓ Mapped %s → %s\n", dir, profileName)
	return nil
}

func runMapList() error {
	mappings, err := mapping.List()
	if err != nil {
		return err
	}
	if len(mappings) == 0 {
		fmt.Println("No mappings configured. Use: bpm map <dir> <profile>")
		return nil
	}
	fmt.Println("Directory → Profile mappings:\n")
	for _, m := range mappings {
		fmt.Printf("  %s → %s\n", m.Directory, m.Profile)
	}
	return nil
}

func runMapAuto() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current directory: %w", err)
	}

	profileName, err := mapping.Get(cwd)
	if err != nil {
		return err
	}
	if profileName == "" {
		fmt.Printf("No profile mapped for %s\n", cwd)
		fmt.Println("Use: bpm map <dir> <profile> to create a mapping")
		return nil
	}

	fmt.Printf("Profile for %s: %s\n", cwd, profileName)
	return nil
}

func runMapRemove(dir string) error {
	if err := mapping.Remove(dir); err != nil {
		return err
	}
	fmt.Printf("✓ Removed mapping for %s\n", dir)
	return nil
}

func init() {
	mapCmd.Flags().BoolVar(&mapList, "list", false, "Show all directory → profile mappings")
	mapCmd.Flags().BoolVar(&mapAuto, "auto", false, "Auto-resolve profile for current directory")
	mapCmd.Flags().StringVar(&mapRemove, "remove", "", "Remove mapping for directory")
	rootCmd.AddCommand(mapCmd)
}
