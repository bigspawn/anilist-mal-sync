package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMappingCache_SetGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMappingCache[string](filepath.Join(tmpDir, "mappings.json"), 720*time.Hour)

	cache.Set("key", "value")

	retrieved, found := cache.Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", *retrieved)
}

func TestMappingCache_NotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMappingCache[string](filepath.Join(tmpDir, "mappings.json"), 720*time.Hour)

	_, found := cache.Get("missing")
	assert.False(t, found)
}

func TestMappingCache_ExpiredEntryIsAMiss(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMappingCache[string](filepath.Join(tmpDir, "mappings.json"), time.Nanosecond)

	cache.Set("key", "value")

	_, found := cache.Get("key")
	assert.False(t, found)
}

func TestMappingCache_ZeroMaxAgeKeepsEntriesForever(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewMappingCache[string](filepath.Join(tmpDir, "mappings.json"), 0)

	cache.Set("key", "value")

	retrieved, found := cache.Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", *retrieved)
}

func TestMappingCache_DirtyFlagSkipsSave(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mappings.json")
	cache := NewMappingCache[string](filePath, 720*time.Hour)

	ctx := t.Context()

	// Nothing was set, cache is clean: Save must not create the file.
	assert.NoError(t, cache.Save(ctx))
	_, err := os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))

	cache.Set("key", "value")
	assert.NoError(t, cache.Save(ctx))
	info, err := os.Stat(filePath)
	assert.NoError(t, err)
	initialModTime := info.ModTime()

	time.Sleep(10 * time.Millisecond)

	// Second save without changes should not touch the file.
	assert.NoError(t, cache.Save(ctx))
	info2, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Equal(t, initialModTime, info2.ModTime())
}

func TestMappingCache_SaveLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mappings.json")
	cache := NewMappingCache[string](filePath, 720*time.Hour)

	cache.Set("key", "value")
	assert.NoError(t, cache.Save(t.Context()))

	cache2 := NewMappingCache[string](filePath, 720*time.Hour)
	assert.Equal(t, 1, cache2.Size())

	retrieved, found := cache2.Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", *retrieved)
}

func TestMappingCache_CorruptFileStartsFresh(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mappings.json")
	assert.NoError(t, os.WriteFile(filePath, []byte("{not json"), 0o600))

	cache := NewMappingCache[string](filePath, 720*time.Hour)

	assert.Equal(t, 0, cache.Size(), "an unreadable cache must not break the run")
}
