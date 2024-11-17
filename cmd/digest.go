package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richardamare/ai-digest/internal/processor"
	"github.com/spf13/cobra"
)

var (
	// Command flags
	inputDir          string
	outputFile        string
	useDefaultIgnores bool
	removeWhitespace  bool
	showOutputFiles   bool
	ignoreFile        string
	splitOutput       bool
	maxFileSizeMB     int
	outputPattern     string
	chunkSize         int
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Create a digest of your codebase",
	Long: `Digest creates a digest of your codebase in single or multiple markdown files.
It handles text and binary files, respects ignore patterns, and provides various formatting options.

Examples:
  ai-digest digest -i /path/to/project -o output.md
  ai-digest digest -i /path/to/project -o output.md --split --max-size 5
  ai-digest digest -i /path/to/project -o output.md --split --output-pattern "part_%d.md"`,
	RunE:    runDigest,
	PreRunE: validateFlags,
}

func init() {
	// Required flags
	digestCmd.Flags().StringVarP(&inputDir, "input", "i", ".",
		"Input directory containing the codebase")
	digestCmd.Flags().StringVarP(&outputFile, "output", "o", "codebase.md",
		"Output markdown file path")

	// Optional flags
	digestCmd.Flags().BoolVar(&useDefaultIgnores, "no-default-ignores", true,
		"Disable default ignore patterns")
	digestCmd.Flags().BoolVar(&removeWhitespace, "whitespace-removal", false,
		"Enable whitespace removal for non-sensitive files")
	digestCmd.Flags().BoolVar(&showOutputFiles, "show-output-files", false,
		"Display a list of files included in the output")
	digestCmd.Flags().StringVar(&ignoreFile, "ignore-file", ".aidigestignore",
		"Custom ignore file name")

	// Split-specific flags
	digestCmd.Flags().BoolVar(&splitOutput, "split", false,
		"Split output into multiple files")
	digestCmd.Flags().IntVar(&maxFileSizeMB, "max-size", 10,
		"Maximum size of each output file in MB (only used with --split)")
	digestCmd.Flags().StringVar(&outputPattern, "output-pattern", "",
		"Pattern for split output files (e.g., 'part_%d.md')")
	digestCmd.Flags().IntVar(&chunkSize, "chunk-size", 1,
		"Size of processing chunks in MB")

	rootCmd.AddCommand(digestCmd)
}

func validateFlags(cmd *cobra.Command, args []string) error {
	// Validate input directory
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}

	// Validate and create output directory
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Validate max file size
	if maxFileSizeMB <= 0 {
		return fmt.Errorf("max-size must be greater than 0")
	}

	// Validate chunk size
	if chunkSize <= 0 {
		return fmt.Errorf("chunk-size must be greater than 0")
	}

	// Validate output pattern if provided
	if splitOutput && outputPattern != "" {
		_ = fmt.Sprintf(outputPattern, 1)
	}

	return nil
}

func runDigest(cmd *cobra.Command, args []string) error {
	// Create processor configuration
	config := processor.ProcessorConfig{
		InputDir:          inputDir,
		OutputFile:        outputFile,
		UseDefaultIgnores: useDefaultIgnores,
		RemoveWhitespace:  removeWhitespace,
		ShowOutputFiles:   showOutputFiles,
		IgnoreFile:        ignoreFile,
		Split:             splitOutput,
		MaxFileSizeMB:     maxFileSizeMB,
		OutputFilePattern: outputPattern,
		ChunkSize:         chunkSize * 1024 * 1024, // Convert to bytes
	}

	// Create processor instance
	proc, err := processor.NewProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	// Process the codebase
	if err := proc.Process(); err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	return nil
}
