package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tquangkhai98/browser-profiles-manager/internal/browser"
	"github.com/tquangkhai98/browser-profiles-manager/internal/profile"
)

var useBrowser string

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Launch browser with an isolated profile",
	Long:  `Launch a Chromium browser using the specified profile's data directory. Acquires a lock to prevent concurrent usage and waits for the browser to exit.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get the profile
		p, err := profile.Get(name)
		if err != nil {
			return err
		}

		// Resolve browser: flag → profile default → auto-detect
		browserID := useBrowser
		if browserID == "" {
			browserID = p.Browser
		}
		if browserID == "" {
			b, err := browser.DefaultBrowser()
			if err != nil {
				return err
			}
			browserID = b.ID
		}

		b, err := browser.FindBrowser(browserID)
		if err != nil {
			return err
		}

		// Acquire lock
		releaseLock, err := profile.AcquireLock(p.DataDir, "bpm-cli")
		if err != nil {
			return fmt.Errorf("cannot lock profile %q: %w", name, err)
		}

		// Set up signal handler for clean shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		fmt.Printf("Launching %s with profile %q...\n", b.Name, name)

		// Launch browser
		proc, err := browser.Launch(b.ExePath, p.DataDir)
		if err != nil {
			releaseLock()
			return err
		}

		// Update last used timestamp
		_ = profile.UpdateLastUsed(name)

		// Wait for browser exit or signal
		doneCh := make(chan error, 1)
		go func() {
			doneCh <- proc.Wait()
		}()

		select {
		case err := <-doneCh:
			releaseLock()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Browser exited with error: %v\n", err)
			} else {
				fmt.Println("Browser closed.")
			}
		case sig := <-sigCh:
			fmt.Printf("\nReceived %s, cleaning up...\n", sig)
			if proc.Process != nil {
				proc.Process.Signal(syscall.SIGTERM)
			}
			releaseLock()
		}

		return nil
	},
}

func init() {
	useCmd.Flags().StringVar(&useBrowser, "browser", "", "Override browser (chrome, brave, edge, arc)")
	rootCmd.AddCommand(useCmd)
}
