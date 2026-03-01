package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newTestJikanClient creates a JikanClient pointing at a test server.
func newTestJikanClient(t *testing.T, serverURL, cacheDir string) *JikanClient {
	t.Helper()
	cache := NewJikanCache(cacheDir, 168*time.Hour)
	return &JikanClient{
		baseURL:    serverURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		cache:      cache,
	}
}

func jikanMangaResponse(id int, title, titleEN, titleJP string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"mal_id":         id,
			"title":          title,
			"title_english":  titleEN,
			"title_japanese": titleJP,
			"title_synonyms": []string{},
			"type":           "Manga",
			"chapters":       nil,
			"volumes":        nil,
			"status":         "Publishing",
		},
	}
}

func jikanSearchResponseHelper(entries ...map[string]interface{}) map[string]interface{} {
	data := make([]interface{}, 0, len(entries))
	for _, e := range entries {
		data = append(data, e)
	}
	return map[string]interface{}{
		"data": data,
		"pagination": map[string]interface{}{
			"last_visible_page": 1,
			"has_next_page":     false,
		},
	}
}

func jikanSearchEntry(id int, title, titleEN, titleJP string) map[string]interface{} {
	return map[string]interface{}{
		"mal_id":         id,
		"title":          title,
		"title_english":  titleEN,
		"title_japanese": titleJP,
		"title_synonyms": []string{},
		"type":           "Manga",
		"chapters":       nil,
		"volumes":        nil,
		"status":         "Publishing",
	}
}

func TestJikanClient_GetMangaByMALID(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manga/13" {
			writeJSON(t, w, jikanMangaResponse(13, "One Piece", "One Piece", "ONE PIECE"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	data, found := client.GetMangaByMALID(ctx, 13)
	assert.True(t, found)
	assert.NotNil(t, data)
	assert.Equal(t, 13, data.MalID)
	assert.Equal(t, "One Piece", data.Title)
	assert.Equal(t, "One Piece", data.TitleEnglish)
	assert.Equal(t, "ONE PIECE", data.TitleJapanese)
}

func TestJikanClient_GetMangaByMALID_CacheHit(t *testing.T) {
	t.Parallel()
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		writeJSON(t, w, jikanMangaResponse(13, "One Piece", "One Piece", "ONE PIECE"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	// First request — API call
	data1, found1 := client.GetMangaByMALID(ctx, 13)
	assert.True(t, found1)
	assert.Equal(t, 13, data1.MalID)
	assert.Equal(t, 1, requestCount)

	// Second request — cache hit
	data2, found2 := client.GetMangaByMALID(ctx, 13)
	assert.True(t, found2)
	assert.Equal(t, 13, data2.MalID)
	assert.Equal(t, 1, requestCount, "Second request should NOT hit API (cache hit)")
}

func TestJikanClient_GetMangaByMALID_NotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	data, found := client.GetMangaByMALID(ctx, 999999)
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestJikanClient_GetMangaByMALID_NegativeCache(t *testing.T) {
	t.Parallel()
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	// First request — 404
	_, found1 := client.GetMangaByMALID(ctx, 999999)
	assert.False(t, found1)
	assert.Equal(t, 1, requestCount)

	// Second request — should hit negative cache
	_, found2 := client.GetMangaByMALID(ctx, 999999)
	assert.False(t, found2)
	assert.Equal(t, 1, requestCount, "Should hit negative cache")
}

func TestJikanClient_SearchManga(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "One Piece" {
			writeJSON(t, w, jikanSearchResponseHelper(
				jikanSearchEntry(13, "One Piece", "One Piece", "ONE PIECE"),
				jikanSearchEntry(100, "One Piece: Film Z", "One Piece Film Z", ""),
			))
			return
		}
		writeJSON(t, w, jikanSearchResponseHelper())
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	results := client.SearchManga(ctx, "One Piece")
	assert.Len(t, results, 2)
	assert.Equal(t, 13, results[0].MalID)
	assert.Equal(t, "One Piece", results[0].Title)
}

func TestJikanClient_SearchManga_Empty(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, jikanSearchResponseHelper())
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	results := client.SearchManga(ctx, "nonexistent_manga_xyz")
	assert.Empty(t, results)
}

func TestJikanAPIStrategy_FindTarget_MangaFound(t *testing.T) {
	t.Parallel()
	// Seed cache directly to avoid HTTP calls
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	searchResults := []JikanMangaData{
		{MalID: 92182, Title: "Seishun Buta Yarou", TitleEnglish: "Rascal Does Not Dream of Bunny Girl Senpai", TitleJapanese: "青春ブタ野郎"},
	}
	encoded, _ := json.Marshal(searchResults)
	cache.SetSearch("Seishun Buta Yarou", encoded)

	client := &JikanClient{cache: cache}
	strategy := JikanAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(t.Context())

	// AniList→MAL direction: IDMal=0, need to find it
	src := Manga{
		IDAnilist:   87471,
		IDMal:       0,
		TitleEN:     "Rascal Does Not Dream of Bunny Girl Senpai",
		TitleJP:     "青春ブタ野郎",
		TitleRomaji: "Seishun Buta Yarou",
	}

	targetManga := Manga{
		IDAnilist:   87471,
		IDMal:       92182,
		TitleEN:     "Rascal Does Not Dream of Bunny Girl Senpai",
		TitleJP:     "青春ブタ野郎",
		TitleRomaji: "Seishun Buta Yarou",
	}

	existingTargets := map[TargetID]Target{
		TargetID(92182): targetManga,
	}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Rascal Does Not Dream of Bunny Girl Senpai", target.GetTitle())
}

func TestJikanAPIStrategy_FindTarget_SkipsAnime(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)
	client := &JikanClient{cache: cache}
	strategy := JikanAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(t.Context())

	src := Anime{
		IDMal:     1,
		IDAnilist: 0,
		TitleEN:   "Cowboy Bebop",
	}

	existingTargets := map[TargetID]Target{}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestJikanAPIStrategy_NilClient(t *testing.T) {
	t.Parallel()
	strategy := JikanAPIStrategy{Client: nil}
	ctx := NewLogger(false).WithContext(t.Context())

	src := Manga{IDMal: 0, IDAnilist: 123}
	existingTargets := map[TargetID]Target{}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestJikanAPIStrategy_FindTarget_ReverseDirection(t *testing.T) {
	t.Parallel()
	// Seed cache with manga data for MAL ID lookup
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)

	mangaData := JikanMangaData{
		MalID:         92182,
		Title:         "Seishun Buta Yarou",
		TitleEnglish:  "Rascal Does Not Dream of Bunny Girl Senpai",
		TitleJapanese: "青春ブタ野郎",
	}
	encoded, _ := json.Marshal(mangaData)
	cache.Set(92182, encoded)

	client := &JikanClient{cache: cache}
	strategy := JikanAPIStrategy{Client: client}
	ctx := NewLogger(false).WithContext(t.Context())

	// MAL→AniList direction: IDAnilist=0, need to find match by title
	src := Manga{
		IDAnilist:   0,
		IDMal:       92182,
		TitleEN:     "Rascal Does Not Dream of Bunny Girl Senpai",
		TitleJP:     "青春ブタ野郎",
		TitleRomaji: "",
	}

	targetManga := Manga{
		IDAnilist:   87471,
		IDMal:       92182,
		TitleEN:     "Rascal Does Not Dream of Bunny Girl Senpai",
		TitleJP:     "青春ブタ野郎",
		TitleRomaji: "Seishun Buta Yarou",
	}

	existingTargets := map[TargetID]Target{
		TargetID(87471): targetManga,
	}

	target, found, err := strategy.FindTarget(ctx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Rascal Does Not Dream of Bunny Girl Senpai", target.GetTitle())
}

func TestMatchJikanMangaToSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		jikan    JikanMangaData
		en, jp   string
		romaji   string
		expected bool
	}{
		{
			name:     "exact english match",
			jikan:    JikanMangaData{TitleEnglish: "One Piece", TitleJapanese: "ONE PIECE", Title: "One Piece"},
			en:       "One Piece",
			jp:       "",
			romaji:   "",
			expected: true,
		},
		{
			name:     "exact japanese match",
			jikan:    JikanMangaData{TitleEnglish: "", TitleJapanese: "ONE PIECE", Title: "One Piece"},
			en:       "",
			jp:       "ONE PIECE",
			romaji:   "",
			expected: true,
		},
		{
			name:     "romaji to title match",
			jikan:    JikanMangaData{TitleEnglish: "", TitleJapanese: "", Title: "Seishun Buta Yarou"},
			en:       "",
			jp:       "",
			romaji:   "Seishun Buta Yarou",
			expected: true,
		},
		{
			name:     "no match",
			jikan:    JikanMangaData{TitleEnglish: "Naruto", TitleJapanese: "ナルト", Title: "Naruto"},
			en:       "One Piece",
			jp:       "ONE PIECE",
			romaji:   "One Piece",
			expected: false,
		},
		{
			name: "synonym match",
			jikan: JikanMangaData{
				TitleEnglish:  "Bunny Girl Senpai",
				TitleJapanese: "青春ブタ野郎",
				Title:         "Seishun Buta Yarou",
				TitleSynonyms: []string{"Rascal Does Not Dream of Bunny Girl Senpai"},
			},
			en:       "Rascal Does Not Dream of Bunny Girl Senpai",
			jp:       "",
			romaji:   "",
			expected: true,
		},
		{
			name:     "cross match EN vs romaji",
			jikan:    JikanMangaData{TitleEnglish: "", TitleJapanese: "", Title: "Seishun Buta Yarou"},
			en:       "Seishun Buta Yarou",
			jp:       "",
			romaji:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchJikanMangaToSource(t.Context(), &tt.jikan, tt.en, tt.jp, tt.romaji)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindBestJikanMatch(t *testing.T) {
	t.Parallel()
	results := []JikanMangaData{
		{MalID: 100, Title: "One Piece Film Z", TitleEnglish: "One Piece Film Z", TitleJapanese: ""},
		{MalID: 13, Title: "One Piece", TitleEnglish: "One Piece", TitleJapanese: "ONE PIECE"},
	}

	malID := findBestJikanMatch(t.Context(), results, "One Piece", "ONE PIECE", "")
	assert.Equal(t, 13, malID)
}

func TestFindBestJikanMatch_NoMatch(t *testing.T) {
	t.Parallel()
	results := []JikanMangaData{
		{MalID: 100, Title: "Naruto", TitleEnglish: "Naruto", TitleJapanese: "ナルト"},
	}

	malID := findBestJikanMatch(t.Context(), results, "One Piece", "ONE PIECE", "")
	assert.Equal(t, 0, malID)
}

func TestSearchTitlesForJikan(t *testing.T) {
	t.Parallel()
	titles := searchTitlesForJikan("One Piece", "", "One Piece")
	// Should deduplicate "One Piece" (romaji == EN after normalization)
	assert.Len(t, titles, 1)

	titles2 := searchTitlesForJikan("Rascal Does Not Dream", "", "Seishun Buta Yarou")
	assert.Len(t, titles2, 2)
	assert.Equal(t, "Seishun Buta Yarou", titles2[0])    // romaji first
	assert.Equal(t, "Rascal Does Not Dream", titles2[1]) // then EN
}

func TestSearchTitlesForJikan_Empty(t *testing.T) {
	t.Parallel()
	titles := searchTitlesForJikan("", "", "")
	assert.Empty(t, titles)
}

func TestJikanClient_GetMangaByMALID_InvalidID(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)
	client := &JikanClient{cache: cache}
	ctx := NewLogger(false).WithContext(t.Context())

	data, found := client.GetMangaByMALID(ctx, 0)
	assert.False(t, found)
	assert.Nil(t, data)

	data, found = client.GetMangaByMALID(ctx, -1)
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestJikanClient_SearchManga_EmptyQuery(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cache := NewJikanCache(tmpDir, 168*time.Hour)
	client := &JikanClient{cache: cache}
	ctx := NewLogger(false).WithContext(t.Context())

	results := client.SearchManga(ctx, "")
	assert.Nil(t, results)
}

func TestJikanClient_ServerError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	client := newTestJikanClient(t, server.URL, tmpDir)
	ctx := NewLogger(false).WithContext(t.Context())

	// Server error is non-fatal
	data, found := client.GetMangaByMALID(ctx, 1)
	assert.False(t, found)
	assert.Nil(t, data)
}
