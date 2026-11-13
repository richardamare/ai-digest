package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Common file type extensions
var (
	binaryExtensions = map[string]string{
		".ico":   "Icon",
		".jpg":   "Image",
		".jpeg":  "Image",
		".png":   "Image",
		".gif":   "Image",
		".bmp":   "Image",
		".webp":  "Image",
		".wasm":  "WebAssembly",
		".pdf":   "PDF",
		".zip":   "Archive",
		".gz":    "Archive",
		".tar":   "Archive",
		".exe":   "Executable",
		".dll":   "Library",
		".so":    "Library",
		".dylib": "Library",
	}

	// Extensions where whitespace is significant
	whitespaceSignificant = map[string]bool{
		".py":   true,
		".yaml": true,
		".yml":  true,
		".nim":  true,
		".haml": true,
		".pug":  true,
		".jade": true,
	}
)

// DefaultFileProcessor implements FileProcessor for standard file handling
type DefaultFileProcessor struct {
	preserveWhitespace bool
}

// NewDefaultFileProcessor creates a new DefaultFileProcessor
func NewDefaultFileProcessor(preserveWhitespace bool) *DefaultFileProcessor {
	return &DefaultFileProcessor{
		preserveWhitespace: preserveWhitespace,
	}
}

// Process implements FileProcessor
func (p *DefaultFileProcessor) Process(ctx context.Context, f *File) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	ext := strings.ToLower(filepath.Ext(f.Path))

	// Handle known binary files
	if fileType, ok := binaryExtensions[ext]; ok {
		return fmt.Sprintf("Binary file of type: %s", fileType), nil
	}

	// Special handling for SVG files
	if ext == ".svg" {
		return string(f.Content), nil
	}

	// Check if file appears to be binary
	if !f.IsText {
		// Do a more thorough check of content
		contentType := http.DetectContentType(f.Content)
		if strings.Contains(contentType, "text") || strings.Contains(contentType, "application/json") {
			f.IsText = true
		} else {
			return fmt.Sprintf("Binary file (detected type: %s)", contentType), nil
		}
	}

	// Process text content
	content := string(f.Content)

	// Remove UTF-8 BOM if present
	if len(f.Content) > 2 && f.Content[0] == 0xEF && f.Content[1] == 0xBB && f.Content[2] == 0xBF {
		content = string(f.Content[3:])
	}

	// Handle whitespace
	if !p.preserveWhitespace && !whitespaceSignificant[ext] {
		content = p.normalizeWhitespace(content)
	}

	// Escape backticks in content if needed
	if strings.Contains(content, "```") {
		content = strings.ReplaceAll(content, "```", "\\`\\`\\`")
	}

	return content, nil
}

// normalizeWhitespace reduces multiple whitespace characters to a single space
func (p *DefaultFileProcessor) normalizeWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	space := false
	for _, r := range s {
		if isWhitespace(r) {
			if !space {
				b.WriteByte(' ')
				space = true
			}
		} else {
			b.WriteRune(r)
			space = false
		}
	}

	return strings.TrimSpace(b.String())
}

// isWhitespace returns true for whitespace characters
func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}

// ReadFile reads a file and creates a File struct
func ReadFile(path string, relativeTo string) (*File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(relativeTo, path)
	if err != nil {
		return nil, err
	}

	// Detect if file is text
	isText := true
	if len(content) > 0 {
		contentType := http.DetectContentType(content)
		isText = strings.Contains(contentType, "text") ||
			strings.Contains(contentType, "application/json") ||
			filepath.Ext(path) == ".svg"
	}

	return &File{
		Path:         relPath,
		Size:         info.Size(),
		IsText:       isText,
		IsExecutable: info.Mode()&0111 != 0,
		Content:      content,
	}, nil
}
