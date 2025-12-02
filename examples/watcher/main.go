// Package main demonstrates a file watcher that automatically recompiles LESS files.
//
// This example watches a directory for changes to .less files and automatically
// compiles them to CSS when modified. Useful for development workflows.
//
// Usage:
//
//	go run main.go [watch-dir] [output-dir]
//	go run main.go ./styles ./dist
//
// If no arguments provided, watches current directory and outputs to ./css/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	less "github.com/toakleaf/less.go/less"
)

// FileInfo tracks file modification times
type FileInfo struct {
	ModTime time.Time
}

func main() {
	// Parse command line flags
	compress := flag.Bool("compress", false, "Minify CSS output")
	interval := flag.Duration("interval", 500*time.Millisecond, "Watch interval")
	flag.Parse()

	// Get watch and output directories from args
	args := flag.Args()
	watchDir := "."
	outputDir := "./css"

	if len(args) >= 1 {
		watchDir = args[0]
	}
	if len(args) >= 2 {
		outputDir = args[1]
	}

	// Ensure directories exist
	if _, err := os.Stat(watchDir); os.IsNotExist(err) {
		log.Fatalf("Watch directory does not exist: %s", watchDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Printf("LESS Watcher Started\n")
	fmt.Printf("  Watching: %s\n", watchDir)
	fmt.Printf("  Output:   %s\n", outputDir)
	fmt.Printf("  Compress: %v\n", *compress)
	fmt.Printf("  Interval: %v\n", *interval)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Track file modification times
	fileCache := make(map[string]FileInfo)

	// Initial compilation of all files
	if err := compileAllFiles(watchDir, outputDir, *compress, fileCache); err != nil {
		log.Printf("Initial compilation error: %v", err)
	}

	// Watch loop
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := watchAndCompile(watchDir, outputDir, *compress, fileCache); err != nil {
			log.Printf("Watch error: %v", err)
		}
	}
}

// compileAllFiles compiles all LESS files in the directory
func compileAllFiles(watchDir, outputDir string, compress bool, cache map[string]FileInfo) error {
	return filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-LESS files
		if info.IsDir() || !strings.HasSuffix(path, ".less") {
			return nil
		}

		// Skip partials (files starting with _)
		if strings.HasPrefix(filepath.Base(path), "_") {
			return nil
		}

		// Compile the file
		if err := compileLessFile(path, outputDir, compress); err != nil {
			log.Printf("Error compiling %s: %v", path, err)
		} else {
			cache[path] = FileInfo{ModTime: info.ModTime()}
		}

		return nil
	})
}

// watchAndCompile checks for file changes and recompiles as needed
func watchAndCompile(watchDir, outputDir string, compress bool, cache map[string]FileInfo) error {
	return filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-LESS files
		if info.IsDir() || !strings.HasSuffix(path, ".less") {
			return nil
		}

		// Skip partials (files starting with _)
		if strings.HasPrefix(filepath.Base(path), "_") {
			// But check if partials changed - need to recompile dependent files
			cached, exists := cache[path]
			if !exists || info.ModTime().After(cached.ModTime) {
				cache[path] = FileInfo{ModTime: info.ModTime()}
				fmt.Printf("[%s] Partial changed: %s (dependent files will recompile on next change)\n",
					time.Now().Format("15:04:05"), filepath.Base(path))
			}
			return nil
		}

		// Check if file was modified
		cached, exists := cache[path]
		if !exists || info.ModTime().After(cached.ModTime) {
			fmt.Printf("[%s] Compiling: %s\n", time.Now().Format("15:04:05"), path)

			startTime := time.Now()
			if err := compileLessFile(path, outputDir, compress); err != nil {
				log.Printf("Error: %v", err)
			} else {
				duration := time.Since(startTime)
				outputPath := getOutputPath(path, outputDir)
				fmt.Printf("[%s] Success: %s (%v)\n",
					time.Now().Format("15:04:05"), filepath.Base(outputPath), duration)
				cache[path] = FileInfo{ModTime: info.ModTime()}
			}
		}

		return nil
	})
}

// compileLessFile compiles a single LESS file to CSS
func compileLessFile(lessPath, outputDir string, compress bool) error {
	// Read the LESS file
	content, err := os.ReadFile(lessPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Get the directory of the LESS file for @import resolution
	lessDir := filepath.Dir(lessPath)

	// Compile with options
	result, err := less.Compile(string(content), &less.CompileOptions{
		Filename: lessPath,
		Paths:    []string{lessDir},
		Compress: compress,
	})
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Write the CSS output
	outputPath := getOutputPath(lessPath, outputDir)
	if err := os.WriteFile(outputPath, []byte(result.CSS), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// getOutputPath converts a .less path to a .css path in the output directory
func getOutputPath(lessPath, outputDir string) string {
	base := filepath.Base(lessPath)
	cssName := strings.TrimSuffix(base, ".less") + ".css"
	return filepath.Join(outputDir, cssName)
}
