package format

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// MarkdownFormatter formats output as markdown
type MarkdownFormatter struct {
	firstFile bool
}

// NewMarkdownFormatter creates a new MarkdownFormatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{
		firstFile: true,
	}
}

// Begin implements Formatter
func (f *MarkdownFormatter) Begin(w io.Writer) error {
	return nil
}

// FormatFile implements Formatter
func (f *MarkdownFormatter) FormatFile(w io.Writer, path string, content string) error {
	if !f.firstFile {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	f.firstFile = false

	// Write file header
	if _, err := fmt.Fprintf(w, "# %s\n\n", path); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))

	// Special handling for markdown files
	if ext == ".md" || ext == ".markdown" {
		if _, err := fmt.Fprintf(w, "```md\n%s\n```\n", content); err != nil {
			return err
		}
		return nil
	}

	// Handle other file types
	if strings.HasPrefix(content, "Binary file") || strings.HasPrefix(content, "Binary file") {
		// Binary files don't need code blocks
		if _, err := fmt.Fprintf(w, "%s\n", content); err != nil {
			return err
		}
		return nil
	}

	// Remove the dot from extension for the code block
	lang := strings.TrimPrefix(ext, ".")
	if lang == "" {
		lang = "text"
	}

	if _, err := fmt.Fprintf(w, "```%s\n%s\n```\n", lang, content); err != nil {
		return err
	}

	return nil
}

// End implements Formatter
func (f *MarkdownFormatter) End(w io.Writer) error {
	return nil
}
