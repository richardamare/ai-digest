package core

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// Digester orchestrates the codebase digesting process
type Digester struct {
	walker    FileWalker
	processor FileProcessor
	formatter Formatter
	options   DigestOptions
}

// NewDigester creates a new Digester with the given components
func NewDigester(w FileWalker, p FileProcessor, f Formatter, opts DigestOptions) *Digester {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}
	if len(opts.IgnoreFiles) == 0 {
		opts.IgnoreFiles = []string{".gitignore"}
	}

	return &Digester{
		walker:    w,
		processor: p,
		formatter: f,
		options:   opts,
	}
}

// Digest processes the codebase and writes the output
func (d *Digester) Digest(ctx context.Context, out io.Writer) (*Result, error) {
	start := time.Now()
	result := &Result{}

	// Initialize formatter
	if err := d.formatter.Begin(out); err != nil {
		return nil, fmt.Errorf("formatter initialization failed: %w", err)
	}
	defer d.formatter.End(out)

	// Create channels for file processing pipeline
	filesCh := make(chan *File, 100)
	resultsCh := make(chan struct {
		path    string
		content string
		err     error
	}, 100)

	// Start walker
	var walkErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(filesCh)
		walkErr = d.walker.Walk(ctx, filesCh)
	}()

	// Start processor workers
	const numWorkers = 4
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for file := range filesCh {
				if file.Size > d.options.MaxFileSize {
					result.FilesSkipped++
					continue
				}

				content, err := d.processor.Process(ctx, file)
				resultsCh <- struct {
					path    string
					content string
					err     error
				}{file.Path, content, err}

				// Update stats
				result.FilesProcessed++
				result.TotalSize += file.Size
				result.ProcessedFiles = append(result.ProcessedFiles, file.Path)

				if file.Size > result.LargestFileSize {
					result.LargestFileSize = file.Size
					result.LargestFile = file.Path
				}
			}
		}()
	}

	// Start result collector
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Write results as they come in
	for res := range resultsCh {
		if res.err != nil {
			return nil, fmt.Errorf("processing %s failed: %w", res.path, res.err)
		}
		if err := d.formatter.FormatFile(out, res.path, res.content); err != nil {
			return nil, fmt.Errorf("formatting %s failed: %w", res.path, res.err)
		}
	}

	if walkErr != nil {
		return nil, fmt.Errorf("walk failed: %w", walkErr)
	}

	result.ProcessingTime = time.Since(start).Seconds()
	return result, nil
}
