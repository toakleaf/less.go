// Package main demonstrates an HTTP server that compiles LESS files on-the-fly.
//
// This example creates a web server that serves LESS files as CSS, compiling them
// dynamically on each request. Useful for development environments.
//
// Usage:
//
//	go run main.go
//	# Then visit http://localhost:8080/styles.less
//
// The server will:
//   - Serve .less files as compiled CSS
//   - Serve static files from the current directory
//   - Optionally cache compiled CSS for better performance
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	less "github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

// CacheEntry stores compiled CSS with metadata
type CacheEntry struct {
	CSS      string
	ModTime  time.Time
	ETag     string
	Compiled time.Time
}

// LessHandler handles LESS file requests
type LessHandler struct {
	rootDir  string
	compress bool
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
	useCache bool
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Server port")
	rootDir := flag.String("root", ".", "Root directory for files")
	compress := flag.Bool("compress", false, "Minify CSS output")
	useCache := flag.Bool("cache", true, "Enable caching (checks file modification time)")
	flag.Parse()

	// Create the handler
	handler := &LessHandler{
		rootDir:  *rootDir,
		compress: *compress,
		cache:    make(map[string]*CacheEntry),
		useCache: *useCache,
	}

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ServeHTTP)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("LESS Development Server\n")
	fmt.Printf("  Serving:  %s\n", *rootDir)
	fmt.Printf("  Address:  http://localhost%s\n", addr)
	fmt.Printf("  Compress: %v\n", *compress)
	fmt.Printf("  Cache:    %v\n", *useCache)
	fmt.Println()
	fmt.Println("Request any .less file to get compiled CSS:")
	fmt.Printf("  http://localhost%s/styles.less\n", addr)
	fmt.Println()

	log.Fatal(http.ListenAndServe(addr, mux))
}

func (h *LessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	urlPath := filepath.Clean(r.URL.Path)
	if urlPath == "/" {
		h.serveIndex(w)
		return
	}

	// Get the file path
	filePath := filepath.Join(h.rootDir, urlPath)

	// Security: prevent directory traversal
	absRoot, _ := filepath.Abs(h.rootDir)
	absPath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absPath, absRoot) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if it's a LESS file
	if strings.HasSuffix(urlPath, ".less") {
		h.serveLess(w, r, filePath)
		return
	}

	// Serve static files
	http.ServeFile(w, r, filePath)
}

func (h *LessHandler) serveLess(w http.ResponseWriter, r *http.Request, filePath string) {
	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check cache
	if h.useCache {
		h.mu.RLock()
		cached, exists := h.cache[filePath]
		h.mu.RUnlock()

		if exists && !info.ModTime().After(cached.ModTime) {
			// Check If-None-Match header
			if r.Header.Get("If-None-Match") == cached.ETag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			// Serve from cache
			h.serveCachedCSS(w, cached)
			log.Printf("[CACHE] %s", filePath)
			return
		}
	}

	// Read and compile the LESS file
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the directory for @import resolution
	lessDir := filepath.Dir(filePath)

	// Compile
	startTime := time.Now()
	result, err := less.Compile(string(content), &less.CompileOptions{
		Filename: filePath,
		Paths:    []string{lessDir},
		Compress: h.compress,
	})
	compileDuration := time.Since(startTime)

	if err != nil {
		// Return error as CSS comment for debugging
		errorCSS := fmt.Sprintf("/* LESS Compilation Error:\n%v\n*/\nbody::before { content: 'LESS Error'; color: red; }", err)
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK) // Still return 200 so browser applies the error message
		w.Write([]byte(errorCSS))
		log.Printf("[ERROR] %s: %v", filePath, err)
		return
	}

	// Generate ETag
	hash := md5.New()
	io.WriteString(hash, result.CSS)
	etag := fmt.Sprintf(`"%x"`, hash.Sum(nil))

	// Update cache
	if h.useCache {
		h.mu.Lock()
		h.cache[filePath] = &CacheEntry{
			CSS:      result.CSS,
			ModTime:  info.ModTime(),
			ETag:     etag,
			Compiled: time.Now(),
		}
		h.mu.Unlock()
	}

	// Serve the CSS
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("X-Compile-Time", compileDuration.String())
	w.Write([]byte(result.CSS))

	log.Printf("[COMPILED] %s (%v)", filePath, compileDuration)
}

func (h *LessHandler) serveCachedCSS(w http.ResponseWriter, entry *CacheEntry) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("ETag", entry.ETag)
	w.Header().Set("X-From-Cache", "true")
	w.Write([]byte(entry.CSS))
}

func (h *LessHandler) serveIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>LESS Development Server</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; }
        h1 { color: #1d365d; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        .example { margin: 20px 0; padding: 20px; background: #f9f9f9; border-radius: 8px; }
    </style>
</head>
<body>
    <h1>LESS Development Server</h1>
    <p>This server compiles LESS files to CSS on-the-fly.</p>

    <div class="example">
        <h3>Usage</h3>
        <p>Request any <code>.less</code> file to get compiled CSS:</p>
        <pre>&lt;link rel="stylesheet" href="/styles.less"&gt;</pre>
    </div>

    <div class="example">
        <h3>Example</h3>
        <p>Create a file called <code>styles.less</code> in the root directory:</p>
        <pre>
@primary: #4a90d9;

body {
    color: @primary;
    font-family: sans-serif;
}
        </pre>
        <p>Then visit: <a href="/styles.less">/styles.less</a></p>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}
