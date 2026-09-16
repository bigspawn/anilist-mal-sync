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

// cacheEntry pairs a cached value with the time it was stored.
type cacheEntry[T any] struct {
	Data     T         `json:"data"`
	CachedAt time.Time `json:"cached_at"`
}

// MappingCache is a generic persistent JSON-based cache for ID-mapping API
// responses, shared by Hato, Jikan and MangaBaka. A non-positive maxAge
// keeps entries forever.
type MappingCache[T any] struct {
	entries  map[string]cacheEntry[T]
	mu       sync.RWMutex
	filePath string
	maxAge   time.Duration
	dirty    bool // Track if cache needs saving
}

// NewMappingCache creates a cache backed by filePath and loads existing data.
// A corrupt cache file is non-fatal: it starts fresh and logs a warning.
func NewMappingCache[T any](filePath string, maxAge time.Duration) *MappingCache[T] {
	cache := &MappingCache[T]{
		entries:  make(map[string]cacheEntry[T]),
		filePath: filePath,
		maxAge:   maxAge,
	}

	if !fileExists(filePath) {
		return cache
	}

	err := cache.load()
	if err != nil {
		LogWarn(context.Background(), "Failed to load cache %s: %v (starting fresh)", filePath, err)
	}

	return cache
}

// Get retrieves a cached value by key.
// Returns (data, found). Expired entries are treated as a cache miss.
func (c *MappingCache[T]) Get(key string) (*T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Negative results are cached too, so without expiry a title that gained
	// a mapping upstream would stay unmapped forever.
	if c.maxAge > 0 && time.Since(entry.CachedAt) > c.maxAge {
		return nil, false
	}

	return &entry.Data, true
}

// Set stores a value in the cache under key.
func (c *MappingCache[T]) Set(key string, data T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry[T]{
		Data:     data,
		CachedAt: time.Now(),
	}
	c.dirty = true
}

// Save persists the cache to disk if dirty.
func (c *MappingCache[T]) Save(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil // No changes to save
	}

	cacheDir := filepath.Dir(c.filePath)
	// #nosec G301 - Cache directory for non-sensitive data
	err := os.MkdirAll(cacheDir, 0o750)
	if err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	// #nosec G306 - Cache file is non-sensitive
	err = os.WriteFile(c.filePath, data, 0o600)
	if err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	c.dirty = false
	LogDebug(ctx, "[Mapping Cache] Saved %d entries to %s", len(c.entries), c.filePath)
	return nil
}

// load reads the cache from disk.
func (c *MappingCache[T]) load() error {
	// #nosec G304 - File path comes from controlled cache directory
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("read cache file: %w", err)
	}

	err = json.Unmarshal(data, &c.entries)
	if err != nil {
		return fmt.Errorf("unmarshal cache: %w", err)
	}

	return nil
}

// Size returns the number of cached entries.
func (c *MappingCache[T]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
