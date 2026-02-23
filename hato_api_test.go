package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHatoClient_GetAniListID_Anime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	anilistID := 1
	malID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/mappings/mal/anime/1", r.URL.Path)
		assert.Equal(t, "Mozilla/5.0", r.Header.Get("User-Agent"))

		response := HatoResponse{}
		response.Data.AniListID = &anilistID
		response.Data.MalID = &malID
		typeStr := mediaTypeAnime
		response.Data.TypeStr = &typeStr
		writeJSON(t, w, response)
	}))
	defer server.Close()

	// Test without cache
	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	id, found, err := client.GetAniListID(ctx, 1, mediaTypeAnime)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 1, id)
}

func TestHatoClient_CacheHit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tmpDir := t.TempDir()
	requestCount := 0

	anilistID := 1
	malID := 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		response := HatoResponse{}
		response.Data.AniListID = &anilistID
		response.Data.MalID = &malID
		writeJSON(t, w, response)
	}))
	defer server.Close()

	// Create client with cache
	client := NewHatoClient(ctx, server.URL, 5*time.Second, tmpDir)

	// First request - cache MISS
	id1, found1, err := client.GetAniListID(ctx, 1, mediaTypeAnime)
	assert.NoError(t, err)
	assert.True(t, found1)
	assert.Equal(t, 1, id1)
	assert.Equal(t, 1, requestCount, "First request should hit API")

	// Second request - cache HIT
	id2, found2, err := client.GetAniListID(ctx, 1, mediaTypeAnime)
	assert.NoError(t, err)
	assert.True(t, found2)
	assert.Equal(t, 1, id2)
	assert.Equal(t, 1, requestCount, "Second request should NOT hit API (cache hit)")
}

func TestHatoClient_GetAniListID_Manga(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	anilistID := 87471
	malID := 92182
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/mappings/mal/manga/92182", r.URL.Path)
		assert.Equal(t, "Mozilla/5.0", r.Header.Get("User-Agent"))

		response := HatoResponse{}
		response.Data.AniListID = &anilistID
		response.Data.MalID = &malID
		typeStr := mediaTypeManga
		response.Data.TypeStr = &typeStr
		writeJSON(t, w, response)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	id, found, err := client.GetAniListID(ctx, 92182, mediaTypeManga)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 87471, id)
}

func TestHatoClient_GetMALID_Anime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	malID := 1
	anilistID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/mappings/anilist/anime/1", r.URL.Path)

		response := HatoResponse{}
		response.Data.MalID = &malID
		response.Data.AniListID = &anilistID
		typeStr := mediaTypeAnime
		response.Data.TypeStr = &typeStr
		writeJSON(t, w, response)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	id, found, err := client.GetMALID(ctx, 1, mediaTypeAnime)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 1, id)
}

func TestHatoClient_GetMALID_Manga(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	malID := 92182
	anilistID := 87471
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/mappings/anilist/manga/87471", r.URL.Path)

		response := HatoResponse{}
		response.Data.MalID = &malID
		response.Data.AniListID = &anilistID
		typeStr := mediaTypeManga
		response.Data.TypeStr = &typeStr
		writeJSON(t, w, response)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	id, found, err := client.GetMALID(ctx, 87471, mediaTypeManga)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, found)
	assert.Equal(t, 92182, id)
}

func TestHatoClient_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return empty data (null IDs)
		writeJSON(t, w, HatoResponse{})
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	_, found, err := client.GetAniListID(ctx, 999999, mediaTypeAnime)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, found)
}

func TestHatoClient_404(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	_, found, err := client.GetAniListID(ctx, 999999, mediaTypeAnime)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, found)
}

func TestHatoClient_ServerError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")

	_, _, err := client.GetAniListID(ctx, 1, mediaTypeAnime)
	assert.Error(t, err)
}

func TestHatoClient_Unreachable(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	client := NewHatoClient(ctx, "http://127.0.0.1:1", 1*time.Second, "")

	_, _, err := client.GetAniListID(ctx, 1, mediaTypeAnime)
	assert.Error(t, err)
}

func TestHatoClient_NegativeCache(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tmpDir := t.TempDir()
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		// Return empty data (not found)
		writeJSON(t, w, HatoResponse{})
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, tmpDir)

	// First request - not found
	_, found1, err := client.GetAniListID(ctx, 999, mediaTypeAnime)
	assert.NoError(t, err)
	assert.False(t, found1)
	assert.Equal(t, 1, requestCount)

	// Second request - should hit cache (negative result)
	_, found2, err := client.GetAniListID(ctx, 999, mediaTypeAnime)
	assert.NoError(t, err)
	assert.False(t, found2)
	assert.Equal(t, 1, requestCount, "Second request should hit negative cache")
}

func TestHatoAPIStrategy_FindTarget_Anime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	anilistID := 1
	malID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/mappings/mal/anime/1" {
			response := HatoResponse{}
			response.Data.AniListID = &anilistID
			response.Data.MalID = &malID
			typeStr := mediaTypeAnime
			response.Data.TypeStr = &typeStr
			writeJSON(t, w, response)
			return
		}
		writeJSON(t, w, HatoResponse{})
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")
	strategy := HatoAPIStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(ctx)

	t.Run("found in existing targets", func(t *testing.T) {
		src := Anime{
			IDMal:     1,
			IDAnilist: 0,
			TitleEN:   "Cowboy Bebop",
		}

		targetAnime := Anime{
			IDAnilist: 1,
			IDMal:     1,
			TitleEN:   "Cowboy Bebop",
		}

		existingTargets := map[TargetID]Target{
			TargetID(1): targetAnime,
		}

		target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Cowboy Bebop", target.GetTitle())
	})
}

func TestHatoAPIStrategy_FindTarget_Manga(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	anilistID := 87471
	malID := 92182
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/mappings/mal/manga/92182" {
			response := HatoResponse{}
			response.Data.AniListID = &anilistID
			response.Data.MalID = &malID
			typeStr := mediaTypeManga
			response.Data.TypeStr = &typeStr
			writeJSON(t, w, response)
			return
		}
		writeJSON(t, w, HatoResponse{})
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")
	strategy := HatoAPIStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(ctx)

	t.Run("found in existing targets", func(t *testing.T) {
		src := Manga{
			IDMal:     malID,
			IDAnilist: 0,
			TitleEN:   "Seishun Buta Yarou",
		}

		targetManga := Manga{
			IDAnilist: anilistID,
			IDMal:     malID,
			TitleEN:   "Seishun Buta Yarou",
		}

		existingTargets := map[TargetID]Target{
			TargetID(anilistID): targetManga,
		}

		target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "Seishun Buta Yarou", target.GetTitle())
	})
}

func TestHatoAPIStrategy_NilClient(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	nilStrategy := HatoAPIStrategy{Client: nil}
	logCtx := NewLogger(false).WithContext(ctx)

	src := Anime{IDMal: 1}
	existingTargets := map[TargetID]Target{}

	target, found, err := nilStrategy.FindTarget(logCtx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestHatoAPIStrategy_NotInUserList(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	anilistID := 1
	malID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := HatoResponse{}
		response.Data.AniListID = &anilistID
		response.Data.MalID = &malID
		writeJSON(t, w, response)
	}))
	defer server.Close()

	client := NewHatoClient(ctx, server.URL, 5*time.Second, "")
	strategy := HatoAPIStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(ctx)

	src := Anime{IDMal: 1}
	existingTargets := map[TargetID]Target{} // Empty list

	target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestNewHatoClient_DefaultURL(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	client := NewHatoClient(ctx, "", 5*time.Second, "")
	assert.Equal(t, defaultHatoBaseURL, client.baseURL)
}

func TestNewHatoClient_CustomURL(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	client := NewHatoClient(ctx, "http://localhost:3000", 5*time.Second, "")
	assert.Equal(t, "http://localhost:3000", client.baseURL)
}

func TestNewHatoClient_WithCache(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tmpDir := t.TempDir()
	client := NewHatoClient(ctx, "", 5*time.Second, tmpDir)
	assert.NotNil(t, client.cache)
	assert.Equal(t, 0, client.cache.Size())
}
