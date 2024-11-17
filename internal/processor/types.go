package processor

import (
	"io"
	"sync"
)

// Config holds the processor configuration
type Config struct {
	InputDir          string
	OutputFile        string
	UseDefaultIgnores bool
	RemoveWhitespace  bool
	ShowOutputFiles   bool
	IgnoreFile        string
}

// Stats tracks processing statistics
type Stats struct {
	mu                  sync.RWMutex
	TotalFiles          int
	IncludedCount       int
	DefaultIgnoredCount int
	CustomIgnoredCount  int
	BinaryAndSvgCount   int
	TotalSize           int64
	IncludedFiles       []string
}

// FileResult represents the result of processing a single file
type FileResult struct {
	RelativePath string
	Content      string
	FileType     string
	Size         int64
	Error        error
}

// FileProcessor handles a single file processing operation
type FileProcessor func(path string, w io.Writer) error

// FileFilter determines if a file should be processed
type FileFilter func(path string) (bool, error)

// ContentTransformer modifies file content before writing
type ContentTransformer func(content string, ext string) string
