package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewJikanCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cache := NewJikanCache(tmpDir, 168*time.Hour)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestJikanCache_SetGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	data, _ := json.Marshal(map[string]interface{}{"mal_id": 123, "title": "One Piece"})
	cache.Set(123, data)

	retrieved, found := cache.Get(123)
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.JSONEq(t, `{"mal_id":123,"title":"One Piece"}`, string(retrieved))
}

func TestJikanCache_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	_, found := cache.Get(999)
	assert.False(t, found)
}

func TestJikanCache_SaveLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	data, _ := json.Marshal(map[string]interface{}{"mal_id": 123, "title": "One Piece"})
	cache.Set(123, data)

	ctx := t.Context()
	err := cache.Save(ctx)
	assert.NoError(t, err)

	// Create new cache instance to load from disk
	cache2 := NewJikanCache(tmpDir, 168*time.Hour)
	assert.Equal(t, 1, cache2.Size())

	retrieved, found := cache2.Get(123)
	assert.True(t, found)
	assert.JSONEq(t, `{"mal_id":123,"title":"One Piece"}`, string(retrieved))
}

func TestJikanCache_Expiration(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 1*time.Millisecond)

	data, _ := json.Marshal(map[string]interface{}{"mal_id": 123})
	cache.Set(123, data)

	// Entry should be found immediately
	_, found := cache.Get(123)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Entry should be expired
	_, found = cache.Get(123)
	assert.False(t, found)
}

func TestJikanCache_SearchSetGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	data, _ := json.Marshal([]map[string]interface{}{
		{"mal_id": 123, "title": "One Piece"},
		{"mal_id": 456, "title": "One Piece: Film Z"},
	})
	cache.SetSearch("One Piece", data)

	retrieved, found := cache.GetSearch("One Piece")
	assert.True(t, found)
	assert.NotNil(t, retrieved)

	var results []map[string]interface{}
	err := json.Unmarshal(retrieved, &results)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestJikanCache_SearchExpiration(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 1*time.Millisecond)

	data, _ := json.Marshal([]map[string]interface{}{{"mal_id": 123}})
	cache.SetSearch("test", data)

	_, found := cache.GetSearch("test")
	assert.True(t, found)

	time.Sleep(5 * time.Millisecond)

	_, found = cache.GetSearch("test")
	assert.False(t, found)
}

func TestJikanCache_DirtyFlag(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	ctx := t.Context()

	// First save should create file
	data, _ := json.Marshal(map[string]interface{}{"mal_id": 123})
	cache.Set(123, data)
	err := cache.Save(ctx)
	assert.NoError(t, err)

	filePath := filepath.Join(tmpDir, jikanCacheFile)
	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	initialModTime := info.ModTime()

	time.Sleep(10 * time.Millisecond)

	// Second save without changes should not write file
	err = cache.Save(ctx)
	assert.NoError(t, err)

	info2, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Equal(t, initialModTime, info2.ModTime(), "File should not be modified if cache is not dirty")

	// Add new entry
	data2, _ := json.Marshal(map[string]interface{}{"mal_id": 456})
	cache.Set(456, data2)

	time.Sleep(10 * time.Millisecond)

	// Third save should write file
	err = cache.Save(ctx)
	assert.NoError(t, err)

	info3, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.True(t, info3.ModTime().After(initialModTime), "File should be modified after new entry")
}

func TestJikanCache_NegativeCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	// Cache a null/empty result (negative cache)
	cache.Set(999, json.RawMessage("null"))

	retrieved, found := cache.Get(999)
	assert.True(t, found, "Negative result should be cached")
	assert.Equal(t, "null", string(retrieved))
}

func TestJikanCache_SearchNormalization(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	data, _ := json.Marshal([]map[string]interface{}{{"mal_id": 123}})
	cache.SetSearch("One Piece", data)

	// Should find with different casing due to normalization
	_, found := cache.GetSearch("one piece")
	assert.True(t, found)
}

func TestGetDefaultJikanCacheDir(t *testing.T) {
	t.Parallel()
	dir := getDefaultJikanCacheDir()
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "anilist-mal-sync")
	assert.Contains(t, dir, "jikan-cache")
}
