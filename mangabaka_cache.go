package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	mangaBakaCacheFile = "mappings.json"
	mangaBakaCacheDir  = "mangabaka-cache"
)

// MangaBakaLookupResult is a cached MangaBaka lookup outcome.
// Found distinguishes a cached negative result from a cache miss.
type MangaBakaLookupResult struct {
	ID    int  `json:"id"`
	Found bool `json:"found"`
}

// MangaBakaCache provides persistent JSON-based caching for MangaBaka lookups.
type MangaBakaCache struct {
	*MappingCache[MangaBakaLookupResult]
}

// NewMangaBakaCache creates a new cache instance and loads existing data.
// A non-positive maxAge keeps entries forever.
func NewMangaBakaCache(cacheDir string, maxAge time.Duration) *MangaBakaCache {
	if cacheDir == "" {
		cacheDir = getDefaultMangaBakaCacheDir()
	}

	filePath := filepath.Join(cacheDir, mangaBakaCacheFile)

	cache := NewMappingCache[MangaBakaLookupResult](filePath, maxAge)

	return &MangaBakaCache{MappingCache: cache}
}

// Get retrieves a cached lookup result by direction and id.
// Returns (result, found). Expired entries are treated as a cache miss.
func (c *MangaBakaCache) Get(direction string, id int) (*MangaBakaLookupResult, bool) {
	return c.MappingCache.Get(buildMangaBakaCacheKey(direction, id))
}

// Set stores a lookup result in the cache, positive or negative.
func (c *MangaBakaCache) Set(direction string, id int, result MangaBakaLookupResult) {
	c.MappingCache.Set(buildMangaBakaCacheKey(direction, id), result)
}

// buildMangaBakaCacheKey creates a unique cache key for a direction+id lookup.
// Format: "{direction}_{id}". Examples: "anilist_30002", "mal_2".
func buildMangaBakaCacheKey(direction string, id int) string {
	return fmt.Sprintf("%s_%d", direction, id)
}

func getDefaultMangaBakaCacheDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", mangaBakaCacheDir)
}
