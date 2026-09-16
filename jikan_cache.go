package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	jikanCacheFile = "mappings.json"
	jikanCacheDir  = "jikan-cache"
)

// JikanCache provides persistent JSON-based caching for Jikan API responses.
type JikanCache struct {
	*MappingCache[json.RawMessage]
}

// NewJikanCache creates a new cache instance and loads existing data.
func NewJikanCache(cacheDir string, maxAge time.Duration) *JikanCache {
	if cacheDir == "" {
		cacheDir = getDefaultJikanCacheDir()
	}

	filePath := filepath.Join(cacheDir, jikanCacheFile)

	cache := NewMappingCache[json.RawMessage](filePath, maxAge)

	return &JikanCache{MappingCache: cache}
}

// Get retrieves a cached manga entry by MAL ID.
// Returns (data, found). Expired entries are treated as cache miss.
func (c *JikanCache) Get(malID int) (json.RawMessage, bool) {
	data, found := c.MappingCache.Get(fmt.Sprintf("manga_%d", malID))
	if !found {
		return nil, false
	}
	return *data, true
}

// Set stores a manga entry in the cache.
func (c *JikanCache) Set(malID int, data json.RawMessage) {
	c.MappingCache.Set(fmt.Sprintf("manga_%d", malID), data)
}

// GetSearch retrieves cached search results by normalized query.
// Returns (data, found). Expired entries are treated as cache miss.
func (c *JikanCache) GetSearch(query string) (json.RawMessage, bool) {
	data, found := c.MappingCache.Get("search_" + normalizeTitle(query))
	if !found {
		return nil, false
	}
	return *data, true
}

// SetSearch stores search results in the cache.
func (c *JikanCache) SetSearch(query string, data json.RawMessage) {
	c.MappingCache.Set("search_"+normalizeTitle(query), data)
}

func getDefaultJikanCacheDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", jikanCacheDir)
}
