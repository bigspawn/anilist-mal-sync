package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadUnmappedState_FileNotFound(t *testing.T) {
	t.Parallel()
	state, err := LoadUnmappedState("/tmp/nonexistent-unmapped-test.json")
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Empty(t, state.Entries)
}

func TestUnmappedState_SaveAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "state", "unmapped.json")

	now := time.Now().Truncate(time.Second)
	state := &UnmappedState{
		Entries: []UnmappedEntry{
			{AniListID: 100, Title: "Test Manga", MediaType: "manga", Direction: DirectionForwardStr},
			{AniListID: 200, MALID: 300, Title: "Test Anime", MediaType: "anime", Direction: DirectionReverseStr},
		},
		UpdatedAt: now,
	}

	if err := state.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadUnmappedState(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, loaded.Entries, 2)
	assert.Equal(t, "Test Manga", loaded.Entries[0].Title)
	assert.Equal(t, 100, loaded.Entries[0].AniListID)
	assert.Equal(t, "manga", loaded.Entries[0].MediaType)
	assert.Equal(t, DirectionForwardStr, loaded.Entries[0].Direction)
	assert.Equal(t, 200, loaded.Entries[1].AniListID)
	assert.Equal(t, 300, loaded.Entries[1].MALID)
	assert.Equal(t, DirectionReverseStr, loaded.Entries[1].Direction)
}

func TestUnmappedEntry_DirectionField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unmapped.json")

	state := &UnmappedState{
		Entries: []UnmappedEntry{
			{
				AniListID: 100,
				Title:     "Forward Entry",
				MediaType: "anime",
				Direction: DirectionForwardStr,
			},
			{
				MALID:     500,
				Title:     "Reverse Entry",
				MediaType: "anime",
				Direction: DirectionReverseStr,
			},
		},
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	if err := state.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadUnmappedState(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, DirectionForwardStr, loaded.Entries[0].Direction)
	assert.Equal(t, DirectionReverseStr, loaded.Entries[1].Direction)
}

func TestUnmappedState_MixedDirections(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unmapped.json")

	state := &UnmappedState{
		Entries: []UnmappedEntry{
			{AniListID: 100, Title: "AniList Anime", MediaType: "anime", Direction: DirectionForwardStr},
			{MALID: 200, Title: "MAL Anime", MediaType: "anime", Direction: DirectionReverseStr},
			{AniListID: 300, Title: "AniList Manga", MediaType: "manga", Direction: DirectionForwardStr},
			{MALID: 400, Title: "MAL Manga", MediaType: "manga", Direction: DirectionReverseStr},
		},
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	if err := state.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadUnmappedState(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, loaded.Entries, 4)

	// Verify forward entries
	assert.Equal(t, DirectionForwardStr, loaded.Entries[0].Direction)
	assert.Equal(t, 100, loaded.Entries[0].AniListID)
	assert.Equal(t, DirectionForwardStr, loaded.Entries[2].Direction)
	assert.Equal(t, 300, loaded.Entries[2].AniListID)

	// Verify reverse entries
	assert.Equal(t, DirectionReverseStr, loaded.Entries[1].Direction)
	assert.Equal(t, 200, loaded.Entries[1].MALID)
	assert.Equal(t, DirectionReverseStr, loaded.Entries[3].Direction)
	assert.Equal(t, 400, loaded.Entries[3].MALID)
}

func TestLoadUnmappedState_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unmapped.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadUnmappedState(path)
	assert.Error(t, err)
}
