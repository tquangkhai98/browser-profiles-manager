package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
)

var detectJSON bool

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "List installed Chromium-based browsers",
	Long:  `Scan the system for installed Chromium-based browsers (Chrome, Brave, Edge, Arc) and display their paths.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		browsers := browser.DetectBrowsers()

		if detectJSON {
			data, _ := json.MarshalIndent(browsers, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(browsers) == 0 {
			fmt.Println("No supported Chromium browsers found.")
			fmt.Println("Supported: Chrome, Brave, Edge, Arc")
			return nil
		}

		fmt.Printf("Found %d browser(s):\n\n", len(browsers))
		for _, b := range browsers {
			fmt.Printf("  %-20s %s\n", b.Name, b.ExePath)
			fmt.Printf("  %-20s ID: %s\n\n", "", b.ID)
		}

		return nil
	},
}

func init() {
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(detectCmd)
}
