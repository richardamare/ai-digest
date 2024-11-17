package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Logger provides structured logging with emojis
type Logger struct {
	showTimestamp bool
}

// NewLogger creates a new logger instance
func NewLogger(showTimestamp bool) *Logger {
	return &Logger{
		showTimestamp: showTimestamp,
	}
}

// Log prints a formatted log message with optional emoji
func (l *Logger) Log(format string, emoji string, args ...interface{}) {
	var builder strings.Builder

	if l.showTimestamp {
		builder.WriteString(time.Now().Format("15:04:05 "))
	}

	if emoji != "" {
		builder.WriteString(emoji)
		builder.WriteString(" ")
	}

	builder.WriteString(format)

	fmt.Printf(builder.String()+"\n", args...)
}

// LogDebug prints a debug message
func (l *Logger) LogDebug(format string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		l.Log(format, "🔍", args...)
	}
}

// LogError prints an error message
func (l *Logger) LogError(format string, args ...interface{}) {
	l.Log(format, "❌", args...)
}

// LogWarning prints a warning message
func (l *Logger) LogWarning(format string, args ...interface{}) {
	l.Log(format, "⚠️", args...)
}

// LogSuccess prints a success message
func (l *Logger) LogSuccess(format string, args ...interface{}) {
	l.Log(format, "✅ ", args...)
}
