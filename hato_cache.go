package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	hatoCacheFile = "mappings.json"
	hatoCacheDir  = "hato-cache"
)

// HatoCache provides persistent JSON-based caching for Hato API responses.
type HatoCache struct {
	*MappingCache[HatoResponseData]
}

// NewHatoCache creates a new cache instance and loads existing data.
// A non-positive maxAge keeps entries forever.
//
//nolint:unparam // Error return kept for API compatibility
func NewHatoCache(cacheDir string, maxAge time.Duration) (*HatoCache, error) {
	if cacheDir == "" {
		cacheDir = getDefaultHatoCacheDir()
	}

	filePath := filepath.Join(cacheDir, hatoCacheFile)

	cache := NewMappingCache[HatoResponseData](filePath, maxAge)

	return &HatoCache{MappingCache: cache}, nil
}

// Get retrieves a cached mapping by key.
// Returns (responseData, found). Expired entries are treated as a cache miss.
func (c *HatoCache) Get(service, mediaType string, id int) (*HatoResponseData, bool) {
	return c.MappingCache.Get(buildCacheKey(service, mediaType, id))
}

// Set stores a complete API response in the cache.
func (c *HatoCache) Set(service, mediaType string, id int, data HatoResponseData) {
	c.MappingCache.Set(buildCacheKey(service, mediaType, id), data)
}

// buildCacheKey creates a unique cache key.
// Format: "{service}_{media_type}_{id}"
// Examples: "mal_anime_1", "anilist_manga_87471".
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
