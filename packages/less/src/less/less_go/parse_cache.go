package less_go

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"
)

// ParseCacheEntry holds a cached parse result with metadata
type ParseCacheEntry struct {
	Root        any       // The parsed AST root (typically *Ruleset)
	Hash        string    // SHA256 hash of the source content
	ParsedAt    time.Time // When this entry was parsed
	HitCount    int64     // Number of cache hits
	SourceLen   int       // Length of source for quick validation
}

// GlobalParseCache provides a thread-safe cache for parsed LESS files.
// This enables incremental compilation by reusing parsed ASTs for unchanged imports.
type GlobalParseCache struct {
	mu      sync.RWMutex
	entries map[string]*ParseCacheEntry // keyed by filename

	// Statistics
	hits   int64
	misses int64

	// Configuration
	maxEntries int           // Maximum number of entries (0 = unlimited)
	maxAge     time.Duration // Maximum age of entries (0 = unlimited)
	enabled    bool
}

// globalParseCache is the singleton instance
var globalParseCache = &GlobalParseCache{
	entries:    make(map[string]*ParseCacheEntry),
	maxEntries: 1000,  // Default: cache up to 1000 files
	maxAge:     0,     // Default: no expiration
	enabled:    false, // Disabled by default - enable with LESS_GO_PARSE_CACHE=1
}

// GetGlobalParseCache returns the global parse cache instance
func GetGlobalParseCache() *GlobalParseCache {
	return globalParseCache
}

// hashContent computes a fast hash of the content
func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:16]) // Use first 128 bits for speed
}

// Get retrieves a cached parse result if available and valid
func (c *GlobalParseCache) Get(filename string, content string) (any, bool) {
	// Check environment variable override (enable with LESS_GO_PARSE_CACHE=1)
	if os.Getenv("LESS_GO_PARSE_CACHE") == "1" {
		// Force enable via env var
	} else if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	entry, exists := c.entries[filename]
	c.mu.RUnlock()

	if !exists {
		c.recordMiss()
		return nil, false
	}

	// Quick validation: check length first (cheap)
	if entry.SourceLen != len(content) {
		c.recordMiss()
		return nil, false
	}

	// Full validation: check hash
	hash := hashContent(content)
	if entry.Hash != hash {
		c.recordMiss()
		return nil, false
	}

	// Check age if maxAge is set
	if c.maxAge > 0 && time.Since(entry.ParsedAt) > c.maxAge {
		c.recordMiss()
		return nil, false
	}

	// Cache hit!
	c.mu.Lock()
	entry.HitCount++
	c.mu.Unlock()

	c.recordHit()

	// Return a deep copy of the root to avoid mutation issues
	// For now, we return the original since LESS evaluation creates new nodes
	// If issues arise, implement DeepCopy for AST nodes
	return entry.Root, true
}

// Put stores a parse result in the cache
func (c *GlobalParseCache) Put(filename string, content string, root any) {
	// Check environment variable override (enable with LESS_GO_PARSE_CACHE=1)
	if os.Getenv("LESS_GO_PARSE_CACHE") == "1" {
		// Force enable via env var
	} else if !c.enabled {
		return
	}

	// Evict if at capacity
	c.mu.Lock()
	if c.maxEntries > 0 && len(c.entries) >= c.maxEntries {
		c.evictOldest()
	}

	c.entries[filename] = &ParseCacheEntry{
		Root:      root,
		Hash:      hashContent(content),
		ParsedAt:  time.Now(),
		HitCount:  0,
		SourceLen: len(content),
	}
	c.mu.Unlock()
}

// Invalidate removes a specific entry from the cache
func (c *GlobalParseCache) Invalidate(filename string) {
	c.mu.Lock()
	delete(c.entries, filename)
	c.mu.Unlock()
}

// Clear removes all entries from the cache
func (c *GlobalParseCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*ParseCacheEntry)
	c.hits = 0
	c.misses = 0
	c.mu.Unlock()
}

// SetEnabled enables or disables the cache
func (c *GlobalParseCache) SetEnabled(enabled bool) {
	c.mu.Lock()
	c.enabled = enabled
	c.mu.Unlock()
}

// IsEnabled returns whether the cache is enabled
func (c *GlobalParseCache) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// SetMaxEntries sets the maximum number of cache entries
func (c *GlobalParseCache) SetMaxEntries(max int) {
	c.mu.Lock()
	c.maxEntries = max
	c.mu.Unlock()
}

// SetMaxAge sets the maximum age of cache entries
func (c *GlobalParseCache) SetMaxAge(age time.Duration) {
	c.mu.Lock()
	c.maxAge = age
	c.mu.Unlock()
}

// Stats returns cache statistics
func (c *GlobalParseCache) Stats() (hits, misses int64, entries int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, len(c.entries)
}

// recordHit increments the hit counter
func (c *GlobalParseCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments the miss counter
func (c *GlobalParseCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// evictOldest removes the least recently used entry
// Must be called with lock held
func (c *GlobalParseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.ParsedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ParsedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// Size returns the number of entries in the cache
func (c *GlobalParseCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
