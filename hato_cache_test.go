package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHatoCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cache, err := NewHatoCache(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestHatoCache_SetGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	anilistID := 123
	malID := 456
	typeStr := "anime"
	data := HatoResponseData{
		AniListID: &anilistID,
		MalID:     &malID,
		TypeStr:   &typeStr,
	}

	cache.Set("mal", "anime", 456, data)

	retrieved, found := cache.Get("mal", "anime", 456)
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, 123, *retrieved.AniListID)
	assert.Equal(t, 456, *retrieved.MalID)
}

func TestHatoCache_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	_, found := cache.Get("mal", "anime", 999)
	assert.False(t, found)
}

func TestHatoCache_SaveLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	anilistID := 123
	malID := 456
	typeStr := "anime"
	data := HatoResponseData{
		AniListID: &anilistID,
		MalID:     &malID,
		TypeStr:   &typeStr,
	}

	cache.Set("mal", "anime", 456, data)

	ctx := context.Background()
	err := cache.Save(ctx)
	assert.NoError(t, err)

	// Create new cache instance to load from disk
	cache2, err := NewHatoCache(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, cache2.Size())

	retrieved, found := cache2.Get("mal", "anime", 456)
	assert.True(t, found)
	assert.Equal(t, 123, *retrieved.AniListID)
}

func TestHatoCache_DirtyFlag(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	ctx := context.Background()

	// First save should create file
	anilistID := 123
	data := HatoResponseData{AniListID: &anilistID}
	cache.Set("mal", "anime", 456, data)
	err := cache.Save(ctx)
	assert.NoError(t, err)

	filePath := filepath.Join(tmpDir, hatoCacheFile)
	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	initialModTime := info.ModTime()

	// Wait a bit to ensure mod time would be different
	time.Sleep(10 * time.Millisecond)

	// Second save without changes should not write file
	err = cache.Save(ctx)
	assert.NoError(t, err)

	info2, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Equal(t, initialModTime, info2.ModTime(), "File should not be modified if cache is not dirty")

	// Add new entry
	anilistID2 := 789
	data2 := HatoResponseData{AniListID: &anilistID2}
	cache.Set("mal", "anime", 789, data2)

	time.Sleep(10 * time.Millisecond)

	// Third save should write file
	err = cache.Save(ctx)
	assert.NoError(t, err)

	info3, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.True(t, info3.ModTime().After(initialModTime), "File should be modified after new entry")
}

func TestHatoCache_BuildCacheKey(t *testing.T) {
	t.Parallel()
	key := buildCacheKey("mal", "anime", 123)
	assert.Equal(t, "mal_anime_123", key)

	key2 := buildCacheKey("anilist", "manga", 456)
	assert.Equal(t, "anilist_manga_456", key2)
}

func TestHatoCache_MultipleEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	// Add multiple entries
	for i := 1; i <= 10; i++ {
		anilistID := i * 100
		data := HatoResponseData{AniListID: &anilistID}
		cache.Set("mal", "anime", i, data)
	}

	assert.Equal(t, 10, cache.Size())

	// Verify all entries
	for i := 1; i <= 10; i++ {
		retrieved, found := cache.Get("mal", "anime", i)
		assert.True(t, found, "Entry %d should be found", i)
		assert.Equal(t, i*100, *retrieved.AniListID)
	}
}

func TestHatoCache_NegativeCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache, _ := NewHatoCache(tmpDir)

	// Cache a negative result (empty data)
	data := HatoResponseData{}
	cache.Set("mal", "anime", 999, data)

	retrieved, found := cache.Get("mal", "anime", 999)
	assert.True(t, found, "Negative result should be cached")
	assert.Nil(t, retrieved.AniListID, "AniListID should be nil for negative cache")
	assert.Nil(t, retrieved.MalID, "MalID should be nil for negative cache")
}

func TestGetDefaultHatoCacheDir(t *testing.T) {
	t.Parallel()
	dir := getDefaultHatoCacheDir()
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "anilist-mal-sync")
	assert.Contains(t, dir, "hato-cache")
}
