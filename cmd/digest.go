package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/richardamare/ai-digest/internal/core"
	"github.com/richardamare/ai-digest/internal/format"
	"github.com/richardamare/ai-digest/internal/walker"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Create a digest of your codebase",
	Long: `Digest creates a markdown digest of your codebase.
It handles text and binary files, respects gitignore patterns,
and provides various formatting options.

Examples:
  ai-digest digest -i /path/to/project -o output.md
  ai-digest digest -i /path/to/project --preserve-ws`,
	RunE: runDigest,
}

func runDigest(cmd *cobra.Command, args []string) error {
	// Resolve absolute paths
	absInputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("invalid input directory: %w", err)
	}

	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("invalid output file: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(absOutputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open output file
	out, err := os.Create(absOutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Create components
	w, err := walker.NewDefaultWalker(absInputDir, []string{ignoreFile})
	if err != nil {
		return fmt.Errorf("failed to create walker: %w", err)
	}

	p := core.NewDefaultFileProcessor(preserveWhitespace)
	f := format.NewMarkdownFormatter()

	// Create digester
	d := core.NewDigester(w, p, f, core.DigestOptions{
		RootDir:     absInputDir,
		IgnoreFiles: []string{ignoreFile},
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	})

	// Process with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run digest
	result, err := d.Digest(ctx, out)
	if err != nil {
		return fmt.Errorf("digest failed: %w", err)
	}

	// Print summary
	cmd.Println("\n📊 Processing Summary")
	cmd.Println("═══════════════════\n")
	cmd.Printf("📁 Files processed: %d\n", result.FilesProcessed)
	cmd.Printf("⏭️  Files skipped: %d\n", result.FilesSkipped)
	cmd.Printf("💾 Total size: %.2f MB\n", float64(result.TotalSize)/(1024*1024))
	cmd.Printf("⏱️  Processing time: %.2f seconds\n", result.ProcessingTime)
	cmd.Printf("📄 Largest file: %s (%.2f MB)\n",
		result.LargestFile,
		float64(result.LargestFileSize)/(1024*1024))
	cmd.Printf("\n✨ Output written to: %s\n", absOutputFile)

	return nil
}
