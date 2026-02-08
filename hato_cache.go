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
	hatoCacheFile = "mappings.json"
	hatoCacheDir  = "hato-cache"
)

// HatoCacheEntry represents a single cached mapping with full API response.
type HatoCacheEntry struct {
	Data     HatoResponseData `json:"data"`
	CachedAt time.Time        `json:"cached_at"`
}

// HatoCache provides persistent JSON-based caching for Hato API responses.
type HatoCache struct {
	entries  map[string]HatoCacheEntry
	mu       sync.RWMutex
	filePath string
	dirty    bool // Track if cache needs saving
}

// NewHatoCache creates a new cache instance and loads existing data.
//
//nolint:unparam // Error return kept for API compatibility
func NewHatoCache(cacheDir string) (*HatoCache, error) {
	if cacheDir == "" {
		cacheDir = getDefaultHatoCacheDir()
	}

	filePath := filepath.Join(cacheDir, hatoCacheFile)

	cache := &HatoCache{
		entries:  make(map[string]HatoCacheEntry),
		filePath: filePath,
	}

	// Load existing cache if it exists
	if fileExists(filePath) {
		if err := cache.load(); err != nil {
			// Non-fatal: continue with empty cache
			LogWarn(context.Background(), "Failed to load Hato cache: %v (starting fresh)", err)
		}
	}

	return cache, nil
}

// Get retrieves a cached mapping by key.
// Returns (responseData, found).
func (c *HatoCache) Get(service, mediaType string, id int) (*HatoResponseData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := buildCacheKey(service, mediaType, id)
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	return &entry.Data, true
}

// Set stores a complete API response in the cache.
func (c *HatoCache) Set(service, mediaType string, id int, data HatoResponseData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := buildCacheKey(service, mediaType, id)
	c.entries[key] = HatoCacheEntry{
		Data:     data,
		CachedAt: time.Now(),
	}
	c.dirty = true
}

// Save persists the cache to disk if dirty.
func (c *HatoCache) Save(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil // No changes to save
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(c.filePath)
	// #nosec G301 - Cache directory for non-sensitive data
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	// Write to file
	// #nosec G306 - Cache file is non-sensitive
	if err := os.WriteFile(c.filePath, data, 0o600); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	c.dirty = false
	LogDebug(ctx, "[Hato Cache] Saved %d entries to %s", len(c.entries), c.filePath)
	return nil
}

// load reads the cache from disk.
func (c *HatoCache) load() error {
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
func (c *HatoCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// buildCacheKey creates a unique cache key.
// Format: "{service}_{media_type}_{id}"
// Examples: "mal_anime_1", "anilist_manga_87471"
func buildCacheKey(service, mediaType string, id int) string {
	return fmt.Sprintf("%s_%s_%d", service, mediaType, id)
}

func getDefaultHatoCacheDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", hatoCacheDir)
}
