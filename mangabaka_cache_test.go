package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMangaBakaCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	cache := NewMangaBakaCache(tmpDir, 720*time.Hour)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Size())
}

func TestMangaBakaCache_SetGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMangaBakaCache(tmpDir, 720*time.Hour)

	cache.Set("anilist", 30002, MangaBakaLookupResult{ID: 2, Found: true})

	retrieved, found := cache.Get("anilist", 30002)
	assert.True(t, found)
	assert.Equal(t, MangaBakaLookupResult{ID: 2, Found: true}, *retrieved)
}

func TestMangaBakaCache_NegativeResultCached(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMangaBakaCache(tmpDir, 720*time.Hour)

	cache.Set("mal", 158901, MangaBakaLookupResult{Found: false})

	retrieved, found := cache.Get("mal", 158901)
	assert.True(t, found, "negative result should be cached")
	assert.False(t, retrieved.Found)
}

func TestMangaBakaCache_NotInCache(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMangaBakaCache(tmpDir, 720*time.Hour)

	_, found := cache.Get("anilist", 999)
	assert.False(t, found)
}

func TestMangaBakaCache_Expiration(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMangaBakaCache(tmpDir, 1*time.Millisecond)

	cache.Set("anilist", 30002, MangaBakaLookupResult{ID: 2, Found: true})

	_, found := cache.Get("anilist", 30002)
	assert.True(t, found)

	time.Sleep(5 * time.Millisecond)

	_, found = cache.Get("anilist", 30002)
	assert.False(t, found)
}

func TestMangaBakaCache_SaveLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMangaBakaCache(tmpDir, 720*time.Hour)

	cache.Set("mal", 2, MangaBakaLookupResult{ID: 30002, Found: true})

	err := cache.Save(t.Context())
	assert.NoError(t, err)

	cache2 := NewMangaBakaCache(tmpDir, 720*time.Hour)
	assert.Equal(t, 1, cache2.Size())

	retrieved, found := cache2.Get("mal", 2)
	assert.True(t, found)
	assert.Equal(t, MangaBakaLookupResult{ID: 30002, Found: true}, *retrieved)
}

func TestGetDefaultMangaBakaCacheDir(t *testing.T) {
	t.Parallel()
	dir := getDefaultMangaBakaCacheDir()
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "anilist-mal-sync")
	assert.Contains(t, dir, "mangabaka-cache")
}
