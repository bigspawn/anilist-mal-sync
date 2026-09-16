package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Titles shared across strategy tests.
const (
	testTitleAnime    = "Test Anime"
	testTitleBerserk  = "BERSERK"
	testTitleManga    = "Test Manga"
	testTitleOnePiece = "One Piece"
	// testTitleUnrelated is the decoy a wrong-direction lookup would select.
	testTitleUnrelated        = "Unrelated Series"
	testTitleCowboyBebop      = "Cowboy Bebop"
	testTitleSeishunButaYarou = "Seishun Buta Yarou"
)

// newUnretryingMangaBakaClient points a client at a server without the retry
// wrapper, so failure tests do not sit through the backoff.
func newUnretryingMangaBakaClient(t *testing.T, serverURL string) *MangaBakaClient {
	t.Helper()
	return &MangaBakaClient{
		apiSource: apiSource{
			baseURL:     serverURL,
			httpClient:  &http.Client{Timeout: 5 * time.Second},
			maxFailures: mangaBakaMaxFailureStreak,
		},
	}
}

func TestNewMangaBakaClient_WithCache(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	tmpDir := t.TempDir()
	client := NewMangaBakaClient(ctx, "", 5*time.Second, tmpDir, "720h")
	assert.NotNil(t, client.cache)
	assert.Equal(t, 0, client.cache.Size())
}

func TestMangaBakaClient_DoRequest_EmptyDataIsNotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "anilist-mal-sync", r.Header.Get("User-Agent"))
		writeJSON(t, w, MangaBakaSearchResponse{Data: []MangaBakaSeries{}})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	resp, err := client.doRequest(ctx, server.URL+"/v1/series/search?q=anilist:99999999&limit=5")
	assert.NoError(t, err)
	assert.Nil(t, resp)
	assert.Zero(t, client.failureStreak)
}

func TestMangaBakaClient_DoRequest_404IsNotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	resp, err := client.doRequest(ctx, server.URL+"/v1/series/search?q=mal:158901&limit=5")
	assert.NoError(t, err)
	assert.Nil(t, resp)
	assert.Zero(t, client.failureStreak)
}

func TestMangaBakaClient_DoRequest_Success(t *testing.T) {
	t.Parallel()
	anilistID := 30002
	malID := 2
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{
				{
					ID:    1692,
					Title: testTitleBerserk,
					Type:  mediaTypeManga,
					State: "active",
					Source: MangaBakaSource{
						AniList:     MangaBakaSourceID{ID: &anilistID},
						MyAnimeList: MangaBakaSourceID{ID: &malID},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	resp, err := client.doRequest(ctx, server.URL+"/v1/series/search?q=anilist:30002&limit=5")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "active", resp.Data[0].State)
	assert.Equal(t, anilistID, *resp.Data[0].Source.AniList.ID)
	assert.Equal(t, malID, *resp.Data[0].Source.MyAnimeList.ID)
}

func TestMangaBakaClient_DoRequest_ServerErrorIsAFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	resp, err := client.doRequest(ctx, server.URL+"/v1/series/search?q=anilist:1&limit=5")
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "unexpected status")
	assert.Equal(t, 1, client.failureStreak)
}

func TestMangaBakaClient_DoRequest_MalformedPayloadIsAFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	resp, err := client.doRequest(ctx, server.URL+"/v1/series/search?q=anilist:1&limit=5")
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "decode response")
	assert.Equal(t, 1, client.failureStreak)
}

func TestMangaBakaClient_StopsCallingAfterRepeatedFailures(t *testing.T) {
	t.Parallel()
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	for i := range mangaBakaMaxFailureStreak {
		url := fmt.Sprintf("%s/v1/series/search?q=anilist:%d&limit=5", server.URL, 100+i)
		_, err := client.doRequest(ctx, url)
		assert.ErrorContains(t, err, "unexpected status")
	}
	assert.Equal(t, mangaBakaMaxFailureStreak, requests)
	assert.True(t, client.gaveUp())
}

func TestMangaBakaClient_SaveCache(t *testing.T) {
	t.Parallel()
	ctx := NewLogger(false).WithContext(t.Context())

	assert.NoError(t, newUnretryingMangaBakaClient(t, "http://example.test").SaveCache(ctx))
}

// mangaBakaSeries builds a search hit with the given state and opposite-side ids.
func mangaBakaSeries(state string, anilistID, malID *int) MangaBakaSeries {
	return MangaBakaSeries{
		ID:    1692,
		Title: testTitleBerserk,
		Type:  mediaTypeManga,
		State: state,
		Source: MangaBakaSource{
			AniList:     MangaBakaSourceID{ID: anilistID},
			MyAnimeList: MangaBakaSourceID{ID: malID},
		},
	}
}

func TestMangaBakaClient_GetMALID_Success(t *testing.T) {
	t.Parallel()
	anilistID, malID := 30002, 2
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
		})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	id, found, err := client.GetMALID(ctx, anilistID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, malID, id)
	assert.Equal(t, "/series/search", gotPath)
	assert.Equal(t, "q=anilist:30002&limit=5", gotQuery)
}

func TestMangaBakaClient_GetAniListID_Success(t *testing.T) {
	t.Parallel()
	anilistID, malID := 30002, 2
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
		})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	ctx := NewLogger(false).WithContext(t.Context())

	id, found, err := client.GetAniListID(ctx, malID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, anilistID, id)
	assert.Equal(t, "/series/search", gotPath)
	assert.Equal(t, "q=mal:2&limit=5", gotQuery)
}

func TestSelectMangaBakaHit(t *testing.T) {
	t.Parallel()
	anilistID, malID := 30013, 13
	otherMalID := 999

	tests := []struct {
		name      string
		hits      []MangaBakaSeries
		wantID    int
		wantFound bool
	}{
		{
			name:      "single active hit",
			hits:      []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
			wantID:    malID,
			wantFound: true,
		},
		{
			name: "active plus merged duplicate picks the active one",
			hits: []MangaBakaSeries{
				mangaBakaSeries("active", &anilistID, &malID),
				mangaBakaSeries("merged", &anilistID, &otherMalID),
			},
			wantID:    malID,
			wantFound: true,
		},
		{
			name: "two active hits with differing ids is ambiguous",
			hits: []MangaBakaSeries{
				mangaBakaSeries("active", &anilistID, &malID),
				mangaBakaSeries("active", &anilistID, &otherMalID),
			},
			wantFound: false,
		},
		{
			name:      "active hit with null opposite id is not found",
			hits:      []MangaBakaSeries{mangaBakaSeries("active", &anilistID, nil)},
			wantFound: false,
		},
		{
			name:      "no hits",
			hits:      nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, found := selectMangaBakaHit(tt.hits, mangaBakaDirectionAniList)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestMangaBakaClient_Lookup_CachesPositiveResult(t *testing.T) {
	t.Parallel()
	anilistID, malID := 30002, 2
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
		})
	}))
	defer server.Close()

	ctx := NewLogger(false).WithContext(t.Context())
	client := NewMangaBakaClient(ctx, server.URL, 5*time.Second, t.TempDir(), "720h")
	client.httpClient = &http.Client{Timeout: 5 * time.Second}

	id1, found1, err := client.GetMALID(ctx, anilistID)
	assert.NoError(t, err)
	assert.True(t, found1)
	assert.Equal(t, malID, id1)
	assert.Equal(t, 1, requests, "first lookup should hit the API")

	id2, found2, err := client.GetMALID(ctx, anilistID)
	assert.NoError(t, err)
	assert.True(t, found2)
	assert.Equal(t, malID, id2)
	assert.Equal(t, 1, requests, "second lookup should be served from cache")
}

func TestMangaBakaClient_Lookup_CachesNegativeResult(t *testing.T) {
	t.Parallel()
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		writeJSON(t, w, MangaBakaSearchResponse{Data: []MangaBakaSeries{}})
	}))
	defer server.Close()

	ctx := NewLogger(false).WithContext(t.Context())
	client := NewMangaBakaClient(ctx, server.URL, 5*time.Second, t.TempDir(), "720h")
	client.httpClient = &http.Client{Timeout: 5 * time.Second}

	id1, found1, err := client.GetAniListID(ctx, 158901)
	assert.NoError(t, err)
	assert.False(t, found1)
	assert.Zero(t, id1)
	assert.Equal(t, 1, requests, "first lookup should hit the API")

	id2, found2, err := client.GetAniListID(ctx, 158901)
	assert.NoError(t, err)
	assert.False(t, found2)
	assert.Zero(t, id2)
	assert.Equal(t, 1, requests, "negative result should also be cached")
}

func TestMangaBakaStrategy_FindTarget_Manga(t *testing.T) {
	t.Parallel()
	anilistID, malID := 87471, 92182
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
		})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	strategy := MangaBakaStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(t.Context())

	// MAL is the source and AniList the target, so this is a reverse sync.
	src := Manga{IDMal: malID, TitleEN: testTitleBerserk, isReverse: true}
	targetManga := Manga{IDAnilist: anilistID, IDMal: malID, TitleEN: testTitleBerserk}
	existingTargets := map[TargetID]Target{TargetID(anilistID): targetManga}

	target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testTitleBerserk, target.GetTitle())
}

func TestMangaBakaStrategy_NotInUserList(t *testing.T) {
	t.Parallel()
	anilistID, malID := 87471, 92182
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, MangaBakaSearchResponse{
			Data: []MangaBakaSeries{mangaBakaSeries("active", &anilistID, &malID)},
		})
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	strategy := MangaBakaStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(t.Context())

	src := Manga{IDMal: malID, isReverse: true}
	target, found, err := strategy.FindTarget(logCtx, src, map[TargetID]Target{}, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestMangaBakaStrategy_NilClient(t *testing.T) {
	t.Parallel()
	strategy := MangaBakaStrategy{Client: nil}
	logCtx := NewLogger(false).WithContext(t.Context())

	target, found, err := strategy.FindTarget(logCtx, Manga{IDMal: 1}, map[TargetID]Target{}, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestMangaBakaStrategy_AnimeSourceIsNoOp(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("MangaBaka must not be queried for an Anime source")
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	strategy := MangaBakaStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(t.Context())

	target, found, err := strategy.FindTarget(logCtx, Anime{IDMal: 1}, map[TargetID]Target{}, "test", nil)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, target)
}

func TestMangaBakaStrategy_ClientErrorIsANoOp(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newUnretryingMangaBakaClient(t, server.URL)
	strategy := MangaBakaStrategy{Client: client}
	logCtx := NewLogger(false).WithContext(t.Context())

	target, found, err := strategy.FindTarget(logCtx, Manga{IDMal: 1}, map[TargetID]Target{}, "test", nil)
	assert.NoError(t, err, "strategy swallows client errors so the chain continues")
	assert.False(t, found)
	assert.Nil(t, target)
}

// mangaBakaDirectionServer answers each direction with a different series, so a
// test can tell which query the strategy actually issued.
func mangaBakaDirectionServer(t *testing.T, byQuery map[string]MangaBakaSeries) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		series, ok := byQuery[r.URL.Query().Get("q")]
		if !ok {
			writeJSON(t, w, MangaBakaSearchResponse{Data: []MangaBakaSeries{}})
			return
		}
		writeJSON(t, w, MangaBakaSearchResponse{Data: []MangaBakaSeries{series}})
	}))
}

// A forward sync needs a MAL id. Asking MangaBaka for the AniList id instead
// returns a number from the wrong id space, which can collide with an unrelated
// MAL entry in the user's list and mismatch the entry.
func TestMangaBakaStrategy_ForwardSync_LooksUpMALID(t *testing.T) {
	t.Parallel()
	srcAniList, srcMAL := 145164, 92182
	collidingAniListID, correctMALID := 104502, 777

	server := mangaBakaDirectionServer(t, map[string]MangaBakaSeries{
		"mal:92182":      mangaBakaSeries("active", &collidingAniListID, &srcMAL),
		"anilist:145164": mangaBakaSeries("active", &srcAniList, &correctMALID),
	})
	defer server.Close()

	strategy := MangaBakaStrategy{Client: newUnretryingMangaBakaClient(t, server.URL)}
	logCtx := NewLogger(false).WithContext(t.Context())

	// Forward sync: the source is an AniList entry that also knows its MAL id.
	src := Manga{IDAnilist: srcAniList, IDMal: srcMAL, TitleEN: testTitleBerserk}
	// Targets are MAL entries keyed by MAL id; the colliding number is a
	// different series that must never be selected.
	existingTargets := map[TargetID]Target{
		TargetID(collidingAniListID): Manga{IDMal: collidingAniListID, TitleEN: testTitleUnrelated},
		TargetID(correctMALID):       Manga{IDMal: correctMALID, TitleEN: testTitleBerserk},
	}

	target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testTitleBerserk, target.GetTitle())
}

// The mirror case: a reverse sync needs the AniList id.
func TestMangaBakaStrategy_ReverseSync_LooksUpAniListID(t *testing.T) {
	t.Parallel()
	srcAniList, srcMAL := 145164, 92182
	correctAniListID := 104502

	server := mangaBakaDirectionServer(t, map[string]MangaBakaSeries{
		"mal:92182":      mangaBakaSeries("active", &correctAniListID, &srcMAL),
		"anilist:145164": mangaBakaSeries("active", &srcAniList, &srcMAL),
	})
	defer server.Close()

	strategy := MangaBakaStrategy{Client: newUnretryingMangaBakaClient(t, server.URL)}
	logCtx := NewLogger(false).WithContext(t.Context())

	src := Manga{IDAnilist: srcAniList, IDMal: srcMAL, TitleEN: testTitleBerserk, isReverse: true}
	existingTargets := map[TargetID]Target{
		TargetID(correctAniListID): Manga{IDAnilist: correctAniListID, TitleEN: testTitleBerserk},
	}

	target, found, err := strategy.FindTarget(logCtx, src, existingTargets, "test", nil)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testTitleBerserk, target.GetTitle())
}
