package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "bpm",
	Short:   "Browser Profiles Manager — manage isolated Chromium profiles for AI IDEs",
	Long:    `bpm is a CLI tool for managing isolated Chromium browser profiles across AI-powered development environments like Claude Code, Cursor, and Antigravity.`,
	Version: version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
