package utils

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// BinaryFileTypes maps extensions to their file types
var BinaryFileTypes = map[string]string{
	".jpg":   "Image",
	".jpeg":  "Image",
	".png":   "Image",
	".gif":   "Image",
	".bmp":   "Image",
	".webp":  "Image",
	".svg":   "SVG Image",
	".wasm":  "WebAssembly",
	".pdf":   "PDF",
	".doc":   "Word Document",
	".docx":  "Word Document",
	".xls":   "Excel Spreadsheet",
	".xlsx":  "Excel Spreadsheet",
	".ppt":   "PowerPoint Presentation",
	".pptx":  "PowerPoint Presentation",
	".zip":   "Compressed Archive",
	".rar":   "Compressed Archive",
	".7z":    "Compressed Archive",
	".exe":   "Executable",
	".dll":   "Dynamic-link Library",
	".so":    "Shared Object",
	".dylib": "Dynamic Library",
}

// IsTextFile checks if a file is a text file
func IsTextFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read first 512 bytes
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Check content type
	contentType := http.DetectContentType(buffer[:n])

	// Consider SVG files as text
	if strings.HasSuffix(strings.ToLower(path), ".svg") {
		return true, nil
	}

	return !strings.Contains(contentType, "binary"), nil
}

// GetFileType returns the type of file based on its extension
func GetFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if fileType, ok := BinaryFileTypes[ext]; ok {
		return fileType
	}
	return "Binary"
}

// ShouldTreatAsBinary determines if a file should be treated as binary
func ShouldTreatAsBinary(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, isBinary := BinaryFileTypes[ext]
	return isBinary
}

// IsWhitespaceSensitive checks if a file extension is whitespace sensitive
func IsWhitespaceSensitive(ext string) bool {
	return WhitespaceDependentExtensions[ext]
}
