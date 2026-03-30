package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "bpm",
	Short:   "Browser Profiles Manager — manage isolated Chromium profiles for AI IDEs",
	Long:    `bpm is a CLI tool for managing isolated Chromium browser profiles across AI-powered development environments like Claude Code, Cursor, and Antigravity.`,
	Version: version,
}

// Execute runs the root command.
func Execute() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bpm %s (commit: %s, built: %s)\n", version, commit, date))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
