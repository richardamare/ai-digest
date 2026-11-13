package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Common flags
	inputDir           string
	outputFile         string
	ignoreFile         string
	preserveWhitespace bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "ai-digest",
		Short: "AI Digest - Code aggregation tool for AI assistants",
		Long: `AI Digest aggregates your codebase into a single markdown file, 
making it easier to share with AI assistants. It intelligently handles 
text and binary files, respects .gitignore patterns, and provides various 
formatting options.`,
	}
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add persistent flags shared across all commands
	rootCmd.PersistentFlags().StringVarP(&inputDir, "input", "i", ".",
		"Input directory containing the codebase")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "codebase.md",
		"Output markdown file path")
	rootCmd.PersistentFlags().StringVar(&ignoreFile, "ignore", ".gitignore",
		"Path to gitignore file")
	rootCmd.PersistentFlags().BoolVar(&preserveWhitespace, "preserve-ws", false,
		"Preserve whitespace in all files")

	// Add commands
	rootCmd.AddCommand(digestCmd)
	rootCmd.AddCommand(versionCmd)
}
