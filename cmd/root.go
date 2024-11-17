package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "ai-digest",
		Short:   "AI Digest - Code aggregation tool for AI assistants",
		Long:    `AI Digest aggregates your codebase into a single markdown file for easy sharing with AI assistants.`,
		Version: "1.0.0",
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(digestCmd)
}
