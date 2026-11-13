package core

import (
	"context"
	"io"
)

// File represents a file in the codebase
type File struct {
	Path         string // Relative path from root
	Size         int64
	IsText       bool
	IsExecutable bool
	Content      []byte
}

// DigestOptions configures the digest operation
type DigestOptions struct {
	RootDir     string   // Root directory to start from
	IgnoreFiles []string // List of .gitignore files to use (default: [".gitignore"])
	MaxFileSize int64    // Maximum file size to process (default: 10MB)
}

// FileProcessor processes individual files
type FileProcessor interface {
	// Process handles a single file and returns its formatted content
	Process(ctx context.Context, f *File) (string, error)
}

// FileWalker traverses the filesystem and yields files
type FileWalker interface {
	// Walk traverses the filesystem and sends files to the channel
	Walk(ctx context.Context, ch chan<- *File) error
}

// Formatter formats the processed files into the final output
type Formatter interface {
	// Begin is called before processing starts
	Begin(w io.Writer) error
	// FormatFile formats a single file
	FormatFile(w io.Writer, path string, content string) error
	// End is called after all files are processed
	End(w io.Writer) error
}

// Result contains statistics about the digest operation
type Result struct {
	FilesProcessed  int
	FilesSkipped    int
	TotalSize       int64
	ProcessedFiles  []string
	ProcessingTime  float64
	LargestFile     string
	LargestFileSize int64
}
