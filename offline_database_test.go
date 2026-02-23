package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractIDFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		url    string
		prefix string
		wantID int
		wantOK bool
	}{
		{
			name:   "MAL anime URL",
			url:    "https://myanimelist.net/anime/1535",
			prefix: "https://myanimelist.net/anime/",
			wantID: 1535,
			wantOK: true,
		},
		{
			name:   "AniList anime URL",
			url:    "https://anilist.co/anime/10378",
			prefix: "https://anilist.co/anime/",
			wantID: 10378,
			wantOK: true,
		},
		{
			name:   "URL with trailing path",
			url:    "https://myanimelist.net/anime/1535/some-title",
			prefix: "https://myanimelist.net/anime/",
			wantID: 1535,
			wantOK: true,
		},
		{
			name:   "wrong prefix",
			url:    "https://kitsu.io/anime/1535",
			prefix: "https://myanimelist.net/anime/",
			wantID: 0,
			wantOK: false,
		},
		{
			name:   "non-numeric ID",
			url:    "https://myanimelist.net/anime/abc",
			prefix: "https://myanimelist.net/anime/",
			wantID: 0,
			wantOK: false,
		},
		{
			name:   "zero ID",
			url:    "https://myanimelist.net/anime/0",
			prefix: "https://myanimelist.net/anime/",
			wantID: 0,
			wantOK: false,
		},
		{
			name:   "empty URL",
			url:    "",
			prefix: "https://myanimelist.net/anime/",
			wantID: 0,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := extractIDFromURL(tt.url, tt.prefix)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestBuildFromEntries(t *testing.T) {
	t.Parallel()
	entries := []AODEntry{
		{
			Sources: []string{
				"https://myanimelist.net/anime/10378",
				"https://anilist.co/anime/10378",
			},
			Title: "Shinryaku!? Ika Musume",
			Type:  "TV",
		},
		{
			Sources: []string{
				"https://myanimelist.net/anime/8557",
				"https://anilist.co/anime/8557",
			},
			Title: "Shinryaku! Ika Musume",
			Type:  "TV",
		},
		{
			Sources: []string{
				"https://myanimelist.net/anime/1535",
				"https://anilist.co/anime/1535",
				"https://kitsu.io/anime/1234",
			},
			Title: "Death Note",
			Type:  "TV",
		},
		{
			// Entry without AniList source — should not create mapping
			Sources: []string{
				"https://myanimelist.net/anime/99999",
			},
			Title: "Unknown Anime",
			Type:  "TV",
		},
	}

	db := BuildFromEntries(entries)

	assert.Equal(t, 3, db.entries)

	// Test MAL -> AniList
	id, ok := db.GetAniListID(10378)
	assert.True(t, ok)
	assert.Equal(t, 10378, id)

	id, ok = db.GetAniListID(8557)
	assert.True(t, ok)
	assert.Equal(t, 8557, id)

	id, ok = db.GetAniListID(1535)
	assert.True(t, ok)
	assert.Equal(t, 1535, id)

	// Test AniList -> MAL
	id, ok = db.GetMALID(10378)
	assert.True(t, ok)
	assert.Equal(t, 10378, id)

	// Test non-existent ID
	_, ok = db.GetAniListID(99999)
	assert.False(t, ok)

	_, ok = db.GetMALID(99999)
	assert.False(t, ok)
}

func TestOfflineDatabaseGetters_NilValues(t *testing.T) {
	t.Parallel()
	db := &OfflineDatabase{
		malToAniList: make(map[int]int),
		anilistToMAL: make(map[int]int),
	}

	_, ok := db.GetAniListID(123)
	assert.False(t, ok)

	_, ok = db.GetMALID(456)
	assert.False(t, ok)
}

func TestParseAODFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-db.json")

	testData := map[string]interface{}{
		"lastUpdate": "2026-01-15",
		"data": []map[string]interface{}{
			{
				"sources": []string{
					"https://myanimelist.net/anime/10378",
					"https://anilist.co/anime/10378",
				},
				"title": "Shinryaku!? Ika Musume",
				"type":  "TV",
			},
			{
				"sources": []string{
					"https://myanimelist.net/anime/1535",
					"https://anilist.co/anime/1535",
				},
				"title": "Death Note",
				"type":  "TV",
			},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(dbPath, data, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	db, err := parseAODFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, db.entries)
	assert.Equal(t, "2026-01-15", db.lastUpdate)

	id, ok := db.GetAniListID(10378)
	assert.True(t, ok)
	assert.Equal(t, 10378, id)

	id, ok = db.GetMALID(1535)
	assert.True(t, ok)
	assert.Equal(t, 1535, id)
}

func TestOfflineDatabaseStrategy_FindTarget(t *testing.T) {
	db := BuildFromEntries([]AODEntry{
		{
			Sources: []string{
				"https://myanimelist.net/anime/10378",
				"https://anilist.co/anime/10378",
			},
			Title: "Shinryaku!? Ika Musume",
			Type:  "TV",
		},
	})

	strategy := OfflineDatabaseStrategy{Database: db}
	ctx := NewLogger(false).WithContext(context.Background())

	t.Run("found in existing targets", func(t *testing.T) {
		// Set up reverse direction (MAL -> AniList)
		defer setReverseDirectionForTest(true)()

		src := Anime{
			IDMal:     10378,
			IDAnilist: 0,
			TitleEN:   "Shinryaku Ika Musume 2",
		}

		targetAnime := Anime{
			IDAnilist: 10378,
			IDMal:     10378,
			TitleEN:   "Squid Girl Season 2",
		}

		existingTargets := map[TargetID]Target{
			TargetID(10378): targetAnime,
		}

		target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Squid Girl Season 2", target.GetTitle())
	})

	t.Run("mapped but not in user's list", func(t *testing.T) {
		defer setReverseDirectionForTest(true)()

		src := Anime{
			IDMal:     10378,
			IDAnilist: 0,
			TitleEN:   "Shinryaku Ika Musume 2",
		}

		existingTargets := map[TargetID]Target{}

		target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, target)
	})

	t.Run("nil database", func(t *testing.T) {
		nilStrategy := OfflineDatabaseStrategy{Database: nil}

		src := Anime{IDMal: 10378}
		existingTargets := map[TargetID]Target{}

		target, found, err := nilStrategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, target)
	})

	t.Run("skips manga", func(t *testing.T) {
		src := Manga{IDMal: 123}
		existingTargets := map[TargetID]Target{}

		target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, target)
	})
}

func TestOfflineDatabaseStrategy_ReverseSync_Issue38(t *testing.T) {
	// Issue #38: MAL 10378 "Shinryaku Ika Musume 2" should map to AniList 10378 (season 2)
	// NOT to AniList 8557 (season 1 "Squid Girl") via fuzzy title matching
	db := BuildFromEntries([]AODEntry{
		{
			Sources: []string{
				"https://myanimelist.net/anime/10378",
				"https://anilist.co/anime/10378",
			},
			Title: "Shinryaku!? Ika Musume",
			Type:  "TV",
		},
		{
			Sources: []string{
				"https://myanimelist.net/anime/8557",
				"https://anilist.co/anime/8557",
			},
			Title: "Shinryaku! Ika Musume",
			Type:  "TV",
		},
	})

	strategy := OfflineDatabaseStrategy{Database: db}
	ctx := NewLogger(false).WithContext(context.Background())

	defer setReverseDirectionForTest(true)()

	// Source from MAL: Shinryaku Ika Musume 2 (MAL ID: 10378)
	src := Anime{
		IDMal:     10378,
		IDAnilist: 0,
		TitleEN:   "Shinryaku!? Ika Musume",
	}

	// User's AniList list has both seasons
	existingTargets := map[TargetID]Target{
		TargetID(8557): Anime{
			IDAnilist: 8557,
			IDMal:     8557,
			TitleEN:   "Squid Girl",
		},
		TargetID(10378): Anime{
			IDAnilist: 10378,
			IDMal:     10378,
			TitleEN:   "Squid Girl Season 2",
		},
	}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	// Must match season 2, not season 1
	assert.Equal(t, TargetID(10378), target.GetTargetID())
	assert.Equal(t, "Squid Girl Season 2", target.GetTitle())
}

func TestGetCachedVersion(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("valid version file", func(t *testing.T) {
		metaPath := filepath.Join(tmpDir, "version.txt")
		err := os.WriteFile(metaPath, []byte("2026-05\n"), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		version, err := getCachedVersion(metaPath)
		assert.NoError(t, err)
		assert.Equal(t, "2026-05", version)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := getCachedVersion(filepath.Join(tmpDir, "nonexistent.txt"))
		assert.Error(t, err)
	})
}

func TestFileExists(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "exists.txt")
		err := os.WriteFile(path, []byte("test"), 0o600)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, fileExists(path))
	})

	t.Run("non-existing file", func(t *testing.T) {
		assert.False(t, fileExists(filepath.Join(tmpDir, "nope.txt")))
	})
}

func TestGetDefaultCacheDir(t *testing.T) {
	t.Parallel()
	dir := getDefaultCacheDir()
	assert.NotEmpty(t, dir)
	assert.True(t, filepath.IsAbs(dir))
	assert.Contains(t, dir, "anilist-mal-sync")
	assert.Contains(t, dir, "aod-cache")
}
