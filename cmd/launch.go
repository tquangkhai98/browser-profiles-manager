package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	browserruntime "github.com/tquangkhai98/browser-profiles-manager/internal/runtime"
)

var launchHeadless, launchJSON bool
var launchPort int
var launchBrowser string
var launchWait time.Duration

var launchCmd = &cobra.Command{
	Use: "launch <profile>", Short: "Launch a detached browser with a CDP endpoint", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		headless := browser.ResolveHeadless(cmd.Flags().Changed("headless"), launchHeadless, os.Getenv("BPM_HEADLESS"), browser.HasDisplay())
		s, err := browserruntime.Launch(args[0], browserruntime.Options{Browser: launchBrowser, Headless: headless, Port: launchPort, Wait: launchWait})
		if err != nil {
			return err
		}
		if launchJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(launchResult(s))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Launched %s with profile %q (pid %d).\n", s.Browser, s.Profile, s.PID)
		if s.Headless {
			fmt.Fprintf(cmd.OutOrStdout(), "Headless mode: CDP endpoint %s\n", s.CDPHTTP)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "CDP endpoint %s\n", s.CDPHTTP)
		}
		return nil
	},
}

func launchResult(s *browserruntime.State) map[string]any {
	return map[string]any{"profile": s.Profile, "browser": s.Browser, "pid": s.PID, "port": s.Port, "cdp_http": s.CDPHTTP, "webSocketDebuggerUrl": s.WebSocket, "headless": s.Headless}
}
func init() {
	launchCmd.Flags().BoolVar(&launchHeadless, "headless", false, "Run headless (auto-detected if omitted)")
	launchCmd.Flags().BoolVar(&launchJSON, "json", false, "Print JSON")
	launchCmd.Flags().IntVar(&launchPort, "cdp-port", 0, "Remote debugging port (0 = auto)")
	launchCmd.Flags().StringVar(&launchBrowser, "browser", "", "Override browser")
	launchCmd.Flags().DurationVar(&launchWait, "wait", 0, "Wait for the profile lock")
	rootCmd.AddCommand(launchCmd)
}
