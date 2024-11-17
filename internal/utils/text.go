package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
	numberRegex     = regexp.MustCompile(`\d+`)
)

// RemoveWhitespace reduces multiple whitespace characters to a single space
func RemoveWhitespace(s string) string {
	//return whitespaceRegex.ReplaceAllString(strings.TrimSpace(s), " ")
	var buf strings.Builder
	space := false

	for _, r := range s {
		if isWhitespace(r) {
			if !space {
				buf.WriteRune(' ')
				space = true
			}
		} else {
			buf.WriteRune(r)
			space = false
		}
	}

	return strings.TrimSpace(buf.String())
}

// EscapeTripleBackticks escapes triple backticks in text
func EscapeTripleBackticks(s string) string {
	return strings.ReplaceAll(s, "```", "\\`\\`\\`")
}

// NaturalLess implements natural string comparison
func NaturalLess(s1, s2 string) bool {
	s1Parts := numberRegex.Split(s1, -1)
	s2Parts := numberRegex.Split(s2, -1)

	s1Numbers := numberRegex.FindAllString(s1, -1)
	s2Numbers := numberRegex.FindAllString(s2, -1)

	minLen := len(s1Parts)
	if len(s2Parts) < minLen {
		minLen = len(s2Parts)
	}

	for i := 0; i < minLen; i++ {
		if s1Parts[i] != s2Parts[i] {
			return s1Parts[i] < s2Parts[i]
		}
		if i < len(s1Numbers) && i < len(s2Numbers) {
			n1, n2 := s1Numbers[i], s2Numbers[i]
			if n1 != n2 {
				// Compare numbers by length first
				if len(n1) != len(n2) {
					return len(n1) < len(n2)
				}
				return n1 < n2
			}
		}
	}

	return len(s1) < len(s2)
}

// EstimateTokenCount provides a rough estimation of tokens in text
func EstimateTokenCount(text string) int {
	const avgCharsPerToken = 4

	// Count only printable characters
	charCount := 0
	for _, r := range text {
		if unicode.IsPrint(r) {
			charCount++
		}
	}

	return charCount / avgCharsPerToken
}

func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}
