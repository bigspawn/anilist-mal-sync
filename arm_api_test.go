package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		t.Fatalf("failed to encode JSON response: %v", err)
	}
}

func TestARMClient_GetAniListID(t *testing.T) {
	anilistID := 10378
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/ids", r.URL.Path)
		assert.Equal(t, string(ServiceMyAnimeList), r.URL.Query().Get("source"))
		assert.Equal(t, "10378", r.URL.Query().Get("id"))
		assert.Equal(t, "anilist", r.URL.Query().Get("include"))

		writeJSON(t, w, ARMResponse{AniList: &anilistID})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	id, found, err := client.GetAniListID(context.Background(), 10378)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 10378, id)
}

func TestARMClient_GetMALID(t *testing.T) {
	malID := 10378
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/ids", r.URL.Path)
		assert.Equal(t, "anilist", r.URL.Query().Get("source"))
		assert.Equal(t, "10378", r.URL.Query().Get("id"))
		assert.Equal(t, string(ServiceMyAnimeList), r.URL.Query().Get("include"))

		writeJSON(t, w, ARMResponse{MyAnimeList: &malID})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	id, found, err := client.GetMALID(context.Background(), 10378)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 10378, id)
}

func TestARMClient_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, ARMResponse{})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	_, found, err := client.GetAniListID(context.Background(), 999999)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, found)
}

func TestARMClient_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	_, found, err := client.GetAniListID(context.Background(), 999999)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, found)
}

func TestARMClient_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	_, _, err := client.GetAniListID(context.Background(), 10378)
	assert.Error(t, err)
}

func TestARMClient_Unreachable(t *testing.T) {
	client := NewARMClient("http://127.0.0.1:1", 1*time.Second)

	_, _, err := client.GetAniListID(context.Background(), 10378)
	assert.Error(t, err)
}

func TestARMAPIStrategy_FindTarget(t *testing.T) {
	anilistID := 10378
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		idStr := r.URL.Query().Get("id")

		if source == string(ServiceMyAnimeList) && idStr == "10378" {
			writeJSON(t, w, ARMResponse{AniList: &anilistID})
			return
		}

		writeJSON(t, w, ARMResponse{})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)
	strategy := ARMAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(context.Background())

	t.Run("found in existing targets", func(t *testing.T) {
		// ARM strategy accesses raw fields (IDMal/IDAnilist), not direction-aware methods.
		// existingTargets is keyed by the ARM-resolved ID (10378 = AniList ID from MAL ID).
		src := Anime{
			IDMal:     10378,
			IDAnilist: 0,
			TitleEN:   "Shinryaku Ika Musume 2",
			isReverse: true,
		}

		targetAnime := Anime{
			IDAnilist: 10378,
			IDMal:     10378,
			TitleEN:   "Squid Girl Season 2",
			isReverse: true,
		}

		existingTargets := map[TargetID]Target{
			TargetID(10378): targetAnime,
		}

		target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Squid Girl Season 2", target.GetTitle())
	})

	t.Run("not found", func(t *testing.T) {
		src := Anime{
			IDMal:     99999,
			IDAnilist: 0,
			TitleEN:   "Unknown Anime",
		}

		existingTargets := map[TargetID]Target{}

		target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, target)
	})

	t.Run("nil client", func(t *testing.T) {
		nilStrategy := ARMAPIStrategy{Client: nil}

		src := Anime{IDMal: 10378}
		existingTargets := map[TargetID]Target{}

		target, found, err := nilStrategy.FindTarget(ctx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, target)
	})
}

func TestARMAPIStrategy_SkipsManga(t *testing.T) {
	client := NewARMClient("http://unused", 5*time.Second)
	strategy := ARMAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(context.Background())

	src := Manga{IDMal: 123}
	existingTargets := map[TargetID]Target{}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestNewARMClient_DefaultURL(t *testing.T) {
	client := NewARMClient("", 5*time.Second)
	assert.Equal(t, defaultARMBaseURL, client.baseURL)
}

func TestNewARMClient_CustomURL(t *testing.T) {
	client := NewARMClient("http://localhost:3000", 5*time.Second)
	assert.Equal(t, "http://localhost:3000", client.baseURL)
}

// A forward sync needs a MAL id. Asking ARM for the AniList id instead returns a
// number from the wrong id space, which can collide with an unrelated MAL entry
// in the user's list and mismatch the entry.
func TestARMAPIStrategy_ForwardSync_LooksUpMALID(t *testing.T) {
	srcAniList, srcMAL := 21, 30013
	collidingAniListID, correctMALID := 104502, 777

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		id := r.URL.Query().Get("id")

		switch {
		case source == string(ServiceMyAnimeList) && id == "30013":
			writeJSON(t, w, ARMResponse{AniList: &collidingAniListID})
		case source == "anilist" && id == "21":
			writeJSON(t, w, ARMResponse{MyAnimeList: &correctMALID})
		default:
			writeJSON(t, w, ARMResponse{})
		}
	}))
	defer server.Close()

	strategy := ARMAPIStrategy{Client: NewARMClient(server.URL, 5*time.Second)}
	ctx := NewLogger(false).WithContext(context.Background())

	// Forward sync: the source is an AniList entry that also knows its MAL id.
	src := Anime{IDAnilist: srcAniList, IDMal: srcMAL, TitleEN: testTitleOnePiece}
	// Targets are MAL entries keyed by MAL id; the colliding number is a
	// different series that must never be selected.
	existingTargets := map[TargetID]Target{
		TargetID(collidingAniListID): Anime{IDMal: collidingAniListID, TitleEN: testTitleUnrelated},
		TargetID(correctMALID):       Anime{IDMal: correctMALID, TitleEN: testTitleOnePiece},
	}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testTitleOnePiece, target.GetTitle())
}
