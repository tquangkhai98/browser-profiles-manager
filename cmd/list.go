package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
	browserruntime "github.com/tquangkhai98/browser-profiles-manager/internal/runtime"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all browser profiles",
	Long:  `List all browser profiles with their status (free/locked), browser type, and last used timestamp.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profile.List()
		if err != nil {
			return err
		}

		if listJSON {
			return printJSON(profiles)
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles found. Create one with: bpm create <name>")
			return nil
		}

		return printTable(profiles)
	},
}

func printJSON(profiles []profile.ProfileStatus) error {
	items := make([]map[string]any, 0, len(profiles))
	for _, p := range profiles {
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		var item map[string]any
		if err = json.Unmarshal(data, &item); err != nil {
			return err
		}
		state, err := browserruntime.Read(p.Name)
		if err != nil {
			return err
		}
		item["running"] = state != nil
		if state != nil {
			item["pid"] = state.PID
			item["port"] = state.Port
			item["cdp_http"] = state.CDPHTTP
			item["webSocketDebuggerUrl"] = state.WebSocket
			item["headless"] = state.Headless
		}
		items = append(items, item)
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printTable(profiles []profile.ProfileStatus) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tBROWSER\tSTATUS\tCREATED\tLAST USED")
	fmt.Fprintln(w, "----\t-------\t------\t-------\t---------")

	for _, p := range profiles {
		status := "free"
		if p.Locked {
			status = fmt.Sprintf("locked (PID %d)", p.LockInfo.PID)
		}

		lastUsed := "-"
		if p.LastUsed != nil {
			lastUsed = p.LastUsed.Format("2006-01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			p.Name,
			p.Browser,
			status,
			p.CreatedAt.Format("2006-01-02 15:04"),
			lastUsed,
		)
	}
	return w.Flush()
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
}
