package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	jikanCacheFile = "mappings.json"
	jikanCacheDir  = "jikan-cache"
)

// JikanCacheEntry represents a single cached entry with timestamp.
type JikanCacheEntry struct {
	Data     json.RawMessage `json:"data"`
	CachedAt time.Time       `json:"cached_at"`
}

// JikanCache provides persistent JSON-based caching for Jikan API responses.
type JikanCache struct {
	entries  map[string]JikanCacheEntry
	mu       sync.RWMutex
	filePath string
	dirty    bool
	maxAge   time.Duration
}

// NewJikanCache creates a new cache instance and loads existing data.
func NewJikanCache(cacheDir string, maxAge time.Duration) *JikanCache {
	if cacheDir == "" {
		cacheDir = getDefaultJikanCacheDir()
	}

	filePath := filepath.Join(cacheDir, jikanCacheFile)

	cache := &JikanCache{
		entries:  make(map[string]JikanCacheEntry),
		filePath: filePath,
		maxAge:   maxAge,
	}

	if fileExists(filePath) {
		if err := cache.load(); err != nil {
			LogWarn(context.Background(), "Failed to load Jikan cache: %v (starting fresh)", err)
		}
	}

	return cache
}

// Get retrieves a cached manga entry by MAL ID.
// Returns (data, found). Expired entries are treated as cache miss.
func (c *JikanCache) Get(malID int) (json.RawMessage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := fmt.Sprintf("manga_%d", malID)
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if c.maxAge > 0 && time.Since(entry.CachedAt) > c.maxAge {
		return nil, false
	}

	return entry.Data, true
}

// Set stores a manga entry in the cache.
func (c *JikanCache) Set(malID int, data json.RawMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("manga_%d", malID)
	c.entries[key] = JikanCacheEntry{
		Data:     data,
		CachedAt: time.Now(),
	}
	c.dirty = true
}

// GetSearch retrieves cached search results by normalized query.
// Returns (data, found). Expired entries are treated as cache miss.
func (c *JikanCache) GetSearch(query string) (json.RawMessage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := fmt.Sprintf("search_%s", normalizeTitle(query))
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if c.maxAge > 0 && time.Since(entry.CachedAt) > c.maxAge {
		return nil, false
	}

	return entry.Data, true
}

// SetSearch stores search results in the cache.
func (c *JikanCache) SetSearch(query string, data json.RawMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("search_%s", normalizeTitle(query))
	c.entries[key] = JikanCacheEntry{
		Data:     data,
		CachedAt: time.Now(),
	}
	c.dirty = true
}

// Save persists the cache to disk if dirty.
func (c *JikanCache) Save(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil
	}

	cacheDir := filepath.Dir(c.filePath)
	// #nosec G301 - Cache directory for non-sensitive data
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	// #nosec G306 - Cache file is non-sensitive
	if err := os.WriteFile(c.filePath, data, 0o600); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	c.dirty = false
	LogDebug(ctx, "[Jikan Cache] Saved %d entries to %s", len(c.entries), c.filePath)
	return nil
}

// load reads the cache from disk.
func (c *JikanCache) load() error {
	// #nosec G304 - File path comes from controlled cache directory
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("read cache file: %w", err)
	}

	if err := json.Unmarshal(data, &c.entries); err != nil {
		return fmt.Errorf("unmarshal cache: %w", err)
	}

	return nil
}

// Size returns the number of cached entries.
func (c *JikanCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func getDefaultJikanCacheDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", jikanCacheDir)
}
