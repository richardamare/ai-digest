package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/richardamare/ai-digest/internal/utils"
)

const (
	maxConcurrency = 10
	maxFileSize    = 10 * 1024 * 1024 // 10MB
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ProcessorConfig holds all configuration options
type ProcessorConfig struct {
	InputDir          string
	OutputFile        string
	UseDefaultIgnores bool
	RemoveWhitespace  bool
	ShowOutputFiles   bool
	IgnoreFile        string
	Split             bool
	MaxFileSizeMB     int    // Used when Split is true
	OutputFilePattern string // Used when Split is true
	ChunkSize         int    // Buffer size for writing
}

// ProcessorStats tracks all processing statistics
type ProcessorStats struct {
	mu               sync.RWMutex
	TotalFiles       int
	IncludedCount    int
	IgnoredCount     int
	BinaryCount      int
	TotalSize        int64
	IncludedFiles    []string
	NumberOfFiles    int    // Number of output files created
	AverageFileSize  int64  // Average size per output file
	SmallestFile     string // Name of smallest output file
	SmallestFileSize int64  // Size of smallest output file
	LargestFile      string // Name of largest output file
	LargestFileSize  int64  // Size of largest output file
}

// fileWriter is an interface for writing content
type fileWriter interface {
	Write(content string) error
	Close() error
}

// singleFileWriter writes to a single output file
type singleFileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// multiFileWriter writes to multiple files with size limits
type multiFileWriter struct {
	config      ProcessorConfig
	stats       *ProcessorStats
	currentFile *os.File
	writer      *bufio.Writer
	buffer      *bytes.Buffer
	fileIndex   int
	outputSize  int64
	logger      *utils.Logger
	mu          sync.Mutex
}

// Processor handles file processing and output writing
type Processor struct {
	config  ProcessorConfig
	stats   *ProcessorStats
	writer  fileWriter
	logger  *utils.Logger
	matcher *utils.IgnoreMatcher
}

// NewProcessor creates a new processor instance
func NewProcessor(cfg ProcessorConfig) (*Processor, error) {
	if cfg.ChunkSize == 0 {
		cfg.ChunkSize = 1 * 1024 * 1024 // Default 1MB chunk size
	}

	if cfg.MaxFileSizeMB == 0 {
		cfg.MaxFileSizeMB = 10 // Default 10MB max file size
	}

	stats := &ProcessorStats{}
	logger := utils.NewLogger(false)

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(cfg.OutputFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var writer fileWriter
	var err error

	if cfg.Split {
		writer, err = newMultiFileWriter(cfg, stats, logger)
	} else {
		writer, err = newSingleFileWriter(cfg)
	}

	if err != nil {
		return nil, err
	}

	return &Processor{
		config:  cfg,
		stats:   stats,
		writer:  writer,
		logger:  logger,
		matcher: utils.NewIgnoreMatcher(nil, cfg.UseDefaultIgnores),
	}, nil
}

// Process handles the entire processing workflow
func (p *Processor) Process() error {
	defer p.writer.Close()

	// Collect and process files
	files, err := p.collectFiles()
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	results := p.processFiles(files)

	// Write results
	for result := range results {
		if result.Error != nil {
			p.logger.LogError("Error processing %s: %v", result.RelativePath, result.Error)
			continue
		}

		if err := p.writer.Write(result.Content); err != nil {
			return fmt.Errorf("failed to write content: %w", err)
		}

		p.updateStats(result)
	}

	p.printStats()
	return nil
}

func newSingleFileWriter(cfg ProcessorConfig) (*singleFileWriter, error) {
	file, err := os.Create(cfg.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return &singleFileWriter{
		file:   file,
		writer: bufio.NewWriterSize(file, cfg.ChunkSize),
	}, nil
}

func (w *singleFileWriter) Write(content string) error {
	_, err := w.writer.WriteString(content)
	return err
}

func (w *singleFileWriter) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}

func newMultiFileWriter(cfg ProcessorConfig, stats *ProcessorStats, logger *utils.Logger) (*multiFileWriter, error) {
	w := &multiFileWriter{
		config: cfg,
		stats:  stats,
		logger: logger,
		buffer: bytes.NewBuffer(make([]byte, 0, cfg.ChunkSize)),
	}

	// Create first file
	if err := w.createNewFile(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *multiFileWriter) Write(content string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !utf8.ValidString(content) {
		return fmt.Errorf("invalid UTF-8 content detected")
	}

	contentSize := int64(len(content))

	// If this is the first write or current file would exceed size limit
	if w.writer == nil || w.outputSize+contentSize > int64(w.config.MaxFileSizeMB)*1024*1024 {
		if err := w.createNewFile(); err != nil {
			return fmt.Errorf("failed to create new file: %w", err)
		}
		w.outputSize = 0
	}

	if _, err := w.writer.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	w.outputSize += contentSize
	w.updateFileStats(w.getCurrentPath(), w.outputSize)

	// If we're approaching the size limit, flush the writer
	if w.outputSize >= int64(w.config.MaxFileSizeMB)*1024*1024 {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	}

	return nil
}

func (w *multiFileWriter) Close() error {
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	}
	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}

	// Calculate final stats
	if err := w.calculateFinalStats(); err != nil {
		return fmt.Errorf("failed to calculate final stats: %w", err)
	}

	return nil
}

func (w *multiFileWriter) createNewFile() error {
	// Flush and close current file if it exists
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	}
	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}

	// Create new file
	w.fileIndex++
	path := w.getCurrentPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	if _, err := file.Write(utf8BOM); err != nil {
		file.Close()
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	w.currentFile = file
	w.writer = bufio.NewWriterSize(file, w.config.ChunkSize)
	w.stats.NumberOfFiles++

	w.logger.Log("Created new file: %s", "📄", path)
	return nil
}

func (w *multiFileWriter) getCurrentPath() string {
	dir := filepath.Dir(w.config.OutputFile)
	base := filepath.Base(w.config.OutputFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	if w.config.OutputFilePattern != "" {
		return filepath.Join(dir, fmt.Sprintf(w.config.OutputFilePattern, w.fileIndex))
	}

	return filepath.Join(dir, fmt.Sprintf("%s_part%d%s", nameWithoutExt, w.fileIndex, ext))
}

func (w *multiFileWriter) updateFileStats(path string, size int64) {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()

	if w.stats.SmallestFileSize == 0 || size < w.stats.SmallestFileSize {
		w.stats.SmallestFile = path
		w.stats.SmallestFileSize = size
	}

	if size > w.stats.LargestFileSize {
		w.stats.LargestFile = path
		w.stats.LargestFileSize = size
	}
}

func (p *Processor) collectFiles() ([]string, error) {
	var files []string

	p.logger.Log("Collecting files from %s", "🔍", p.config.InputDir)

	err := filepath.Walk(p.config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(p.config.InputDir, path)
		if err != nil {
			return err
		}

		if p.matcher.ShouldIgnore(relPath) {
			p.stats.mu.Lock()
			p.stats.IgnoredCount++
			p.stats.mu.Unlock()
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	p.stats.TotalFiles = len(files)
	p.logger.Log("Found %d files to process", "📚", len(files))
	return files, nil
}

func (w *multiFileWriter) calculateFinalStats() error {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()

	total := int64(0)
	smallest := int64(math.MaxInt64)
	largest := int64(0)
	var smallestFile, largestFile string

	// Scan all generated files
	for i := 1; i <= w.fileIndex; i++ {
		path := w.getCurrentPathForIndex(i)
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", path, err)
		}

		size := info.Size()
		total += size

		if size < smallest {
			smallest = size
			smallestFile = path
		}
		if size > largest {
			largest = size
			largestFile = path
		}
	}

	w.stats.SmallestFileSize = smallest
	w.stats.SmallestFile = smallestFile
	w.stats.LargestFileSize = largest
	w.stats.LargestFile = largestFile

	if w.stats.NumberOfFiles > 0 {
		w.stats.AverageFileSize = total / int64(w.stats.NumberOfFiles)
	}

	return nil
}

func (w *multiFileWriter) getCurrentPathForIndex(index int) string {
	dir := filepath.Dir(w.config.OutputFile)
	base := filepath.Base(w.config.OutputFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	if w.config.OutputFilePattern != "" {
		return filepath.Join(dir, fmt.Sprintf(w.config.OutputFilePattern, index))
	}

	return filepath.Join(dir, fmt.Sprintf("%s_part%d%s", nameWithoutExt, index, ext))
}

func (p *Processor) processFiles(files []string) chan FileResult {
	resultChan := make(chan FileResult, len(files))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for _, file := range files {
		wg.Add(1)
		go func(relPath string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := p.processFile(relPath)
			resultChan <- result
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

func (p *Processor) processFile(relPath string) FileResult {
	result := FileResult{RelativePath: relPath}
	fullPath := filepath.Join(p.config.InputDir, relPath)

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		result.Error = err
		return result
	}
	result.Size = info.Size()

	// Check if file is text
	isText, err := utils.IsTextFile(fullPath)
	if err != nil {
		result.Error = err
		return result
	}

	if isText && !utils.ShouldTreatAsBinary(fullPath) {
		result.FileType = "text"
		content, err := p.processTextFile(fullPath)
		if err != nil {
			result.Error = err
			return result
		}
		result.Content = content
	} else {
		result.FileType = utils.GetFileType(fullPath)
		result.Content = p.formatBinaryFileContent(relPath, result.FileType)
	}

	return result
}

func (p *Processor) processTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content = bytes.TrimPrefix(content, utf8BOM)
	if !utf8.Valid(content) {
		return "", fmt.Errorf("file %s contains invalid UTF-8 characters", path)
	}

	ext := filepath.Ext(path)
	contentStr := string(content)

	if p.config.RemoveWhitespace && !utils.IsWhitespaceSensitive(ext) {
		contentStr = utils.RemoveWhitespace(contentStr)
	}

	relPath, err := filepath.Rel(p.config.InputDir, path)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "# %s\n\n", relPath)

	// For markdown files, use four backticks to wrap content
	if ext == ".md" || ext == ".markdown" {
		buf.WriteString("````md\n")
		buf.WriteString(contentStr)
		buf.WriteString("\n````\n\n")
	} else {
		fmt.Fprintf(&buf, "```%s\n%s\n```\n\n",
			strings.TrimPrefix(ext, "."),
			contentStr)
	}

	return buf.String(), nil
}

func (p *Processor) formatBinaryFileContent(path, fileType string) string {
	var description string
	if strings.HasSuffix(strings.ToLower(path), ".svg") {
		description = fmt.Sprintf("This is a file of type: %s", fileType)
	} else {
		description = fmt.Sprintf("This is a binary file of type: %s", fileType)
	}

	return fmt.Sprintf("# %s\n\n%s\n\n", path, description)
}

func (p *Processor) updateStats(result FileResult) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	p.stats.IncludedCount++
	p.stats.IncludedFiles = append(p.stats.IncludedFiles, result.RelativePath)
	if result.FileType != "text" {
		p.stats.BinaryCount++
	}
	p.stats.TotalSize += result.Size
}

func (p *Processor) printStats() {
	if p.config.Split {
		p.printSplitStats()
	} else {
		p.printSingleStats()
	}
}

func (p *Processor) printSingleStats() {
	fmt.Println("\n📊 Processing Summary")
	fmt.Println("═══════════════════")

	// File counts section
	fmt.Println("\n📁 File Statistics")
	fmt.Printf("   • Total Files Scanned:     %5d\n", p.stats.TotalFiles)
	fmt.Printf("   • Files in Output:         %5d\n", p.stats.IncludedCount)
	fmt.Printf("   • Files Ignored:           %5d\n", p.stats.IgnoredCount)
	fmt.Printf("   • Binary/SVG Files:        %5d\n", p.stats.BinaryCount)

	// Size metrics
	fmt.Println("\n💾 Size Analysis")
	sizeInMB := float64(p.stats.TotalSize) / (1024 * 1024)
	fmt.Printf("   • Total Size:              %.2f MB\n", sizeInMB)

	// Process effectiveness
	fmt.Println("\n🎯 Processing Effectiveness")
	if p.stats.TotalFiles > 0 {
		inclusionRate := float64(p.stats.IncludedCount) / float64(p.stats.TotalFiles) * 100
		fmt.Printf("   • Inclusion Rate:          %5.1f%%\n", inclusionRate)
	}

	// Token estimation
	fmt.Println("\n🔤 Token Estimation")
	if p.stats.TotalSize > maxFileSize {
		fmt.Println("   ⚠️  Output exceeds recommended size (10 MB)")
		fmt.Println("   ⚠️  Token estimation skipped")
		fmt.Printf("   💡 Tip: Add more patterns to %s to reduce size\n", p.config.IgnoreFile)
	} else {
		tokenCount := utils.EstimateTokenCount(fmt.Sprintf("%d", p.stats.TotalSize))
		fmt.Printf("   • Estimated Tokens:        %5d\n", tokenCount)
		fmt.Println("   📝 Note: Token count may vary ±20% across AI models")
	}

	// File listing (if enabled)
	if p.config.ShowOutputFiles && len(p.stats.IncludedFiles) > 0 {
		fmt.Println("\n📋 Included Files")
		fmt.Println("   Files processed and included in output:")
		for i, file := range p.stats.IncludedFiles {
			if i < 10 { // Show first 10 files only
				fmt.Printf("   %2d. %s\n", i+1, file)
			} else {
				remaining := len(p.stats.IncludedFiles) - 10
				fmt.Printf("   ... and %d more files\n", remaining)
				break
			}
		}
	}

	// Final status
	fmt.Println("\n✨ Process Complete")
	if p.stats.TotalSize > maxFileSize {
		fmt.Println("   ⚠️  Warning: Large output file size")
	} else {
		fmt.Println("   ✅ Output generated successfully")
	}
}

func (p *Processor) printSplitStats() {
	fmt.Println("\n📊 Split Processing Summary")
	fmt.Println("═══════════════════════════")

	// Output files information
	fmt.Println("\n📁 Output Files")
	fmt.Printf("   • Number of Files:         %d\n", p.stats.NumberOfFiles)
	fmt.Printf("   • Average File Size:       %.2f MB\n", float64(p.stats.AverageFileSize)/(1024*1024))

	// Size distribution
	fmt.Println("\n📏 Size Distribution")
	fmt.Printf("   • Smallest File:           %s (%.2f MB)\n",
		filepath.Base(p.stats.SmallestFile),
		float64(p.stats.SmallestFileSize)/(1024*1024))
	fmt.Printf("   • Largest File:            %s (%.2f MB)\n",
		filepath.Base(p.stats.LargestFile),
		float64(p.stats.LargestFileSize)/(1024*1024))

	// Processing statistics
	fmt.Println("\n🔍 Processing Details")
	fmt.Printf("   • Total Files Processed:   %d\n", p.stats.TotalFiles)
	fmt.Printf("   • Files Included:          %d\n", p.stats.IncludedCount)
	fmt.Printf("   • Files Ignored:           %d\n", p.stats.IgnoredCount)
	fmt.Printf("   • Binary/SVG Files:        %d\n", p.stats.BinaryCount)

	// Total size
	fmt.Println("\n💾 Total Size")
	fmt.Printf("   • Combined Size:           %.2f MB\n", float64(p.stats.TotalSize)/(1024*1024))

	// Process effectiveness
	fmt.Println("\n🎯 Processing Effectiveness")
	if p.stats.TotalFiles > 0 {
		inclusionRate := float64(p.stats.IncludedCount) / float64(p.stats.TotalFiles) * 100
		fmt.Printf("   • Inclusion Rate:          %5.1f%%\n", inclusionRate)
	}

	// File listing (if enabled)
	if p.config.ShowOutputFiles && len(p.stats.IncludedFiles) > 0 {
		fmt.Println("\n📋 Included Files")
		fmt.Println("   Files processed and included in output:")
		for i, file := range p.stats.IncludedFiles {
			if i < 10 {
				fmt.Printf("   %2d. %s\n", i+1, file)
			} else {
				remaining := len(p.stats.IncludedFiles) - 10
				fmt.Printf("   ... and %d more files\n", remaining)
				break
			}
		}
	}

	// Final status
	fmt.Println("\n✨ Process Complete")
	fmt.Println("   ✅ Output files generated successfully")
}

func hasUTF8BOM(data []byte) bool {
	return bytes.HasPrefix(data, utf8BOM)
}
