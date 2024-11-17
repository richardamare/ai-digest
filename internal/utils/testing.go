package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHelper provides utility functions for testing
type TestHelper struct {
	t      *testing.T
	tmpDir string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	tmpDir, err := os.MkdirTemp("", "aidig-test-*")
	if err != nil {
		t.Fatal(err)
	}

	return &TestHelper{
		t:      t,
		tmpDir: tmpDir,
	}
}

// Cleanup removes temporary test files
func (h *TestHelper) Cleanup() {
	os.RemoveAll(h.tmpDir)
}

// CreateTempFile creates a temporary file with content
func (h *TestHelper) CreateTempFile(name, content string) string {
	path := filepath.Join(h.tmpDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		h.t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		h.t.Fatal(err)
	}

	return path
}

// CreateTempDir creates a temporary directory
func (h *TestHelper) CreateTempDir(name string) string {
	path := filepath.Join(h.tmpDir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		h.t.Fatal(err)
	}
	return path
}
