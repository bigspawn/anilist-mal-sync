package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func writeJSON(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("failed to encode JSON response: %v", err)
	}
}

func TestARMClient_GetAniListID(t *testing.T) {
	anilistID := 10378
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/ids", r.URL.Path)
		assert.Equal(t, "myanimelist", r.URL.Query().Get("source"))
		assert.Equal(t, "10378", r.URL.Query().Get("id"))
		assert.Equal(t, "anilist", r.URL.Query().Get("include"))

		writeJSON(t, w, ARMResponse{AniList: &anilistID})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	id, found, err := client.GetAniListID(t.Context(), 10378)
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
		assert.Equal(t, "myanimelist", r.URL.Query().Get("include"))

		writeJSON(t, w, ARMResponse{MyAnimeList: &malID})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)

	id, found, err := client.GetMALID(t.Context(), 10378)
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

	_, found, err := client.GetAniListID(t.Context(), 999999)
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

	_, found, err := client.GetAniListID(t.Context(), 999999)
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

	_, _, err := client.GetAniListID(t.Context(), 10378)
	assert.Error(t, err)
}

func TestARMClient_Unreachable(t *testing.T) {
	client := NewARMClient("http://127.0.0.1:1", 1*time.Second)

	_, _, err := client.GetAniListID(t.Context(), 10378)
	assert.Error(t, err)
}

func TestARMAPIStrategy_FindTarget(t *testing.T) {
	// Cannot use t.Parallel() - tests below use HTTP servers that may conflict

	anilistID := 10378
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		idStr := r.URL.Query().Get("id")

		if source == "myanimelist" && idStr == "10378" {
			writeJSON(t, w, ARMResponse{AniList: &anilistID})
			return
		}

		writeJSON(t, w, ARMResponse{})
	}))
	defer server.Close()

	client := NewARMClient(server.URL, 5*time.Second)
	strategy := ARMAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(t.Context())

	t.Run("found in existing targets", func(t *testing.T) {
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
	ctx := NewLogger(false).WithContext(t.Context())

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
