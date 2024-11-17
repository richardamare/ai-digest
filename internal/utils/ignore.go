package utils

import (
	"github.com/sabhiram/go-gitignore"
	"path/filepath"
)

// IgnoreMatcher handles file pattern matching for ignored files
type IgnoreMatcher struct {
	customIgnore  *ignore.GitIgnore
	defaultIgnore *ignore.GitIgnore
}

// NewIgnoreMatcher creates a new ignore matcher with the given patterns
func NewIgnoreMatcher(patterns []string, useDefault bool) *IgnoreMatcher {
	matcher := &IgnoreMatcher{}

	if len(patterns) > 0 {
		matcher.customIgnore = ignore.CompileIgnoreLines(patterns...)
	}

	if useDefault {
		matcher.defaultIgnore = ignore.CompileIgnoreLines(DefaultIgnores...)
	}

	return matcher
}

// ShouldIgnore checks if a file should be ignored
func (im *IgnoreMatcher) ShouldIgnore(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	if im.defaultIgnore != nil && im.defaultIgnore.MatchesPath(path) {
		return true
	}

	if im.customIgnore != nil && im.customIgnore.MatchesPath(path) {
		return true
	}

	return false
}
