package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	browserruntime "github.com/tquangkhai98/browser-profiles-manager/internal/runtime"
)

var stopJSON bool
var stopTimeout time.Duration
var stopCmd = &cobra.Command{Use: "stop <profile>", Short: "Stop a detached browser", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	stopped, err := browserruntime.Stop(args[0], stopTimeout)
	if err != nil {
		return err
	}
	if stopJSON {
		result := map[string]any{"profile": args[0], "stopped": stopped}
		if !stopped {
			result["reason"] = "not running"
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
	}
	if stopped {
		fmt.Fprintf(cmd.OutOrStdout(), "Stopped profile %q.\n", args[0])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Profile %q is not running.\n", args[0])
	}
	return nil
}}

func init() {
	stopCmd.Flags().BoolVar(&stopJSON, "json", false, "Print JSON")
	stopCmd.Flags().DurationVar(&stopTimeout, "timeout", 5*time.Second, "Time to wait before forcing termination")
	rootCmd.AddCommand(stopCmd)
}
