// internal/walker/walker.go
package walker

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardamare/ai-digest/internal/core"
)

// DefaultWalker implements FileWalker with gitignore support
type DefaultWalker struct {
	rootDir     string
	ignoreFiles []string
	matcher     *IgnoreMatcher
}

// NewDefaultWalker creates a new DefaultWalker
func NewDefaultWalker(rootDir string, ignoreFiles []string) (*DefaultWalker, error) {
	matcher, err := NewIgnoreMatcher(rootDir, ignoreFiles)
	if err != nil {
		return nil, err
	}

	return &DefaultWalker{
		rootDir:     rootDir,
		ignoreFiles: ignoreFiles,
		matcher:     matcher,
	}, nil
}

// Walk implements FileWalker
func (w *DefaultWalker) Walk(ctx context.Context, ch chan<- *core.File) error {
	return filepath.WalkDir(w.rootDir, func(path string, d fs.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			// Check if directory should be ignored
			if w.shouldIgnore(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be ignored
		if w.shouldIgnore(path) {
			return nil
		}

		// Read and process file
		file, err := core.ReadFile(path, w.rootDir)
		if err != nil {
			// Log error but continue processing
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return nil
		}

		select {
		case ch <- file:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})
}

// shouldIgnore checks if a path should be ignored
func (w *DefaultWalker) shouldIgnore(path string) bool {
	// Get path relative to root
	relPath, err := filepath.Rel(w.rootDir, path)
	if err != nil {
		return false
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	return w.matcher.Matches(relPath)
}

// IgnoreMatcher handles gitignore pattern matching
type IgnoreMatcher struct {
	patterns []pattern
}

// pattern represents a single gitignore pattern
type pattern struct {
	pattern    string
	negate     bool
	matchFiles bool
	matchDirs  bool
}

// NewIgnoreMatcher creates a new IgnoreMatcher from gitignore files
func NewIgnoreMatcher(rootDir string, ignoreFiles []string) (*IgnoreMatcher, error) {
	var patterns []pattern

	// Add default patterns
	defaultPatterns := []string{
		".git/",
		"node_modules/",
		".idea/",
		".vscode/",
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",
		"*.test",
		"*.out",
		".DS_Store",
		"Thumbs.db",
	}

	for _, p := range defaultPatterns {
		pat, err := parsePattern(p)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pat)
	}

	// Read and parse gitignore files
	for _, ignoreFile := range ignoreFiles {
		path := filepath.Join(rootDir, ignoreFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			pat, err := parsePattern(line)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pat)
		}
	}

	return &IgnoreMatcher{patterns: patterns}, nil
}

// Matches checks if a path matches any ignore pattern
func (m *IgnoreMatcher) Matches(path string) bool {
	path = filepath.ToSlash(path)
	isDir := strings.HasSuffix(path, "/")

	var lastMatch *pattern
	for i := len(m.patterns) - 1; i >= 0; i-- {
		p := m.patterns[i]
		if p.matches(path, isDir) {
			if p.negate {
				return false
			}
			lastMatch = &p
		}
	}

	return lastMatch != nil
}

// parsePattern parses a gitignore pattern
func parsePattern(patternStr string) (pattern, error) {
	p := pattern{
		matchFiles: true,
		matchDirs:  true,
	}

	// Handle negation
	if strings.HasPrefix(patternStr, "!") {
		p.negate = true
		patternStr = patternStr[1:]
	}

	// Handle directory-only patterns
	if strings.HasSuffix(patternStr, "/") {
		p.matchFiles = false
		patternStr = patternStr[:len(patternStr)-1]
	}

	// Convert glob pattern to regex
	patternStr = strings.ReplaceAll(patternStr, ".", "\\.")
	patternStr = strings.ReplaceAll(patternStr, "**", ".*")
	patternStr = strings.ReplaceAll(patternStr, "*", "[^/]*")
	patternStr = strings.ReplaceAll(patternStr, "?", "[^/]")
	patternStr = "^" + patternStr + "$"

	p.pattern = patternStr
	return p, nil
}

// matches checks if a path matches this pattern
func (p pattern) matches(path string, isDir bool) bool {
	if !p.matchFiles && !isDir {
		return false
	}
	if !p.matchDirs && isDir {
		return false
	}

	matched, err := filepath.Match(p.pattern, path)
	if err != nil {
		return false
	}
	return matched
}
