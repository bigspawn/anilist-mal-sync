package main

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockToggler is a test double for favouriteToggler.
type mockToggler struct {
	calls []mockToggleCall
	err   error // error to return on all calls
	// errOn maps call index to error (overrides err for specific calls)
	errOn map[int]error
	idx   int
}

type mockToggleCall struct {
	AnimeID int
	MangaID int
}

func (m *mockToggler) ToggleFavourite(_ context.Context, animeID, mangaID int) error {
	m.calls = append(m.calls, mockToggleCall{AnimeID: animeID, MangaID: mangaID})
	defer func() { m.idx++ }()
	if m.errOn != nil {
		if err, ok := m.errOn[m.idx]; ok {
			return err
		}
	}
	return m.err
}

// testCtx returns a context with a silent logger to avoid nil pointer panics in log calls.
func testCtx() context.Context {
	logger := NewLogger(false)
	logger.SetOutput(io.Discard)
	return logger.WithContext(context.Background())
}

// --- Helper builders ---

func makeAnime(anilistID, malID int, title string, isFav bool) Anime {
	return Anime{
		IDAnilist:   anilistID,
		IDMal:       malID,
		TitleEN:     title,
		IsFavourite: isFav,
	}
}

func makeManga(anilistID, malID int, title string, isFav bool) Manga {
	return Manga{
		IDAnilist:   anilistID,
		IDMal:       malID,
		TitleEN:     title,
		IsFavourite: isFav,
	}
}

func favSet(ids ...int) map[int]struct{} {
	m := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

// =============================================================================
// isFavorite tests
// =============================================================================

func TestIsFavorite(t *testing.T) {
	set := favSet(1, 2, 3)

	assert.True(t, isFavorite(set, 1))
	assert.True(t, isFavorite(set, 2))
	assert.True(t, isFavorite(set, 3))
	assert.False(t, isFavorite(set, 4))
	assert.False(t, isFavorite(set, 0))
}

func TestIsFavorite_EmptySet(t *testing.T) {
	assert.False(t, isFavorite(map[int]struct{}{}, 1))
}

func TestIsFavorite_NilSet(t *testing.T) {
	assert.False(t, isFavorite(nil, 1))
}

// =============================================================================
// Constructor tests
// =============================================================================

func TestNewFavoritesSync(t *testing.T) {
	favSync := NewFavoritesSync(nil, false)

	assert.NotNil(t, favSync)
	assert.Nil(t, favSync.toggler)
	assert.False(t, favSync.dryRun)
}

func TestNewFavoritesSync_WithDryRun(t *testing.T) {
	favSync := NewFavoritesSync(nil, true)

	assert.NotNil(t, favSync)
	assert.True(t, favSync.dryRun)
}

// =============================================================================
// SyncToAniList tests
// =============================================================================

func TestSyncToAniList_DryRun_CountsAdded(t *testing.T) {
	// Anime: MAL fav, not AniList fav -> should count as Added in dry run
	// Manga: MAL fav, not AniList fav -> should count as Added in dry run
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: true}

	animes := []Anime{makeAnime(100, 1, "Anime A", false)}
	mangas := []Manga{makeManga(200, 2, "Manga B", false)}

	result := fs.SyncToAniList(testCtx(), animes, mangas, favSet(1), favSet(2))

	assert.Equal(t, 2, result.Added)
	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 0, result.Skipped)
}

func TestSyncToAniList_DryRun_DoesNotCallToggle(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: true}

	animes := []Anime{makeAnime(100, 1, "Anime A", false)}
	fs.SyncToAniList(testCtx(), animes, nil, favSet(1), nil)

	assert.Empty(t, mock.calls, "should not call ToggleFavourite in dry run")
}

func TestSyncToAniList_AddsAnimeAndManga(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{makeAnime(100, 1, "Anime A", false)}
	mangas := []Manga{makeManga(200, 2, "Manga B", false)}

	result := fs.SyncToAniList(testCtx(), animes, mangas, favSet(1), favSet(2))

	assert.Equal(t, 2, result.Added)
	assert.Equal(t, 0, result.Errors)
	assert.Len(t, mock.calls, 2)
	// Anime toggle: animeID=100, mangaID=0
	assert.Equal(t, mockToggleCall{AnimeID: 100, MangaID: 0}, mock.calls[0])
	// Manga toggle: animeID=0, mangaID=200
	assert.Equal(t, mockToggleCall{AnimeID: 0, MangaID: 200}, mock.calls[1])
}

func TestSyncToAniList_SkipsAlreadySynced(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{
		makeAnime(100, 1, "Both fav", true),     // both favorited
		makeAnime(101, 2, "Neither fav", false), // neither favorited
	}

	result := fs.SyncToAniList(testCtx(), animes, nil, favSet(1), nil)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 2, result.Skipped)
	assert.Empty(t, mock.calls)
}

func TestSyncToAniList_SkipsAniListOnlyFavorite(t *testing.T) {
	// AniList favorite but not MAL -> should skip (don't remove)
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{makeAnime(100, 1, "AL only fav", true)}

	result := fs.SyncToAniList(testCtx(), animes, nil, favSet(), nil) // MAL has no favs

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 1, result.Skipped)
	assert.Empty(t, mock.calls)
}

func TestSyncToAniList_SkipsEntriesWithoutMALID(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{makeAnime(100, 0, "No MAL ID", false)}
	mangas := []Manga{makeManga(200, -1, "Negative MAL ID", false)}

	result := fs.SyncToAniList(testCtx(), animes, mangas, favSet(), favSet())

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 2, result.Skipped)
	assert.Empty(t, mock.calls)
}

func TestSyncToAniList_ErrorDoesNotCountAsAdded(t *testing.T) {
	mock := &mockToggler{err: fmt.Errorf("API error")}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{makeAnime(100, 1, "Anime A", false)}

	result := fs.SyncToAniList(testCtx(), animes, nil, favSet(1), nil)

	assert.Equal(t, 0, result.Added, "failed toggle should not count as added")
	assert.Equal(t, 1, result.Errors)
}

func TestSyncToAniList_PartialError(t *testing.T) {
	// First call succeeds, second fails
	mock := &mockToggler{errOn: map[int]error{1: fmt.Errorf("fail")}}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{
		makeAnime(100, 1, "Success", false),
		makeAnime(101, 2, "Failure", false),
	}

	result := fs.SyncToAniList(testCtx(), animes, nil, favSet(1, 2), nil)

	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Errors)
	assert.Len(t, mock.calls, 2)
}

func TestSyncToAniList_MixedAnimeAndManga(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	animes := []Anime{
		makeAnime(100, 1, "Fav anime", false),     // needs add
		makeAnime(101, 2, "Synced anime", true),   // already synced
		makeAnime(102, 0, "No MAL anime", false),  // skip: no MAL ID
		makeAnime(103, 4, "AL only anime", true),  // skip: AL-only fav
		makeAnime(104, 5, "Neither anime", false), // skip: neither
	}
	mangas := []Manga{
		makeManga(200, 10, "Fav manga", false),   // needs add
		makeManga(201, 11, "Synced manga", true), // already synced
	}

	malAnimeFavs := favSet(1, 2)   // 1=fav, 2=fav
	malMangaFavs := favSet(10, 11) // 10=fav, 11=fav

	result := fs.SyncToAniList(testCtx(), animes, mangas, malAnimeFavs, malMangaFavs)

	assert.Equal(t, 2, result.Added)   // anime 100 + manga 200
	assert.Equal(t, 5, result.Skipped) // anime 101(synced) + 102(no MAL) + 103(AL only) + 104(neither) + manga 201(synced)
	assert.Equal(t, 0, result.Errors)
	assert.Len(t, mock.calls, 2)
}

func TestSyncToAniList_EmptyLists(t *testing.T) {
	mock := &mockToggler{}
	fs := &FavoritesSync{toggler: mock, dryRun: false}

	result := fs.SyncToAniList(testCtx(), nil, nil, nil, nil)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, result.Errors)
	assert.Empty(t, mock.calls)
}

// =============================================================================
// ReportMismatches tests
// =============================================================================

func TestReportMismatches_FindsMismatches(t *testing.T) {
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: false}

	animes := []Anime{
		makeAnime(100, 1, "AL only", true),   // on AniList, not on MAL
		makeAnime(101, 2, "MAL only", false), // on MAL, not on AniList
		makeAnime(102, 3, "Both", true),      // both -> no mismatch
		makeAnime(103, 4, "Neither", false),  // neither -> no mismatch
	}

	malFavs := favSet(2, 3) // 2=fav, 3=fav

	result := fs.ReportMismatches(testCtx(), animes, nil, malFavs, nil)

	assert.Len(t, result.Mismatches, 2)

	// AL only: on AniList=true, on MAL=false
	assert.Equal(t, "AL only", result.Mismatches[0].Title)
	assert.True(t, result.Mismatches[0].OnAniList)
	assert.False(t, result.Mismatches[0].OnMAL)
	assert.Equal(t, "anime", result.Mismatches[0].MediaType)

	// MAL only: on AniList=false, on MAL=true
	assert.Equal(t, "MAL only", result.Mismatches[1].Title)
	assert.False(t, result.Mismatches[1].OnAniList)
	assert.True(t, result.Mismatches[1].OnMAL)
}

func TestReportMismatches_NoMismatches(t *testing.T) {
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: false}

	animes := []Anime{
		makeAnime(100, 1, "Both fav", true),
		makeAnime(101, 2, "Neither", false),
	}

	result := fs.ReportMismatches(testCtx(), animes, nil, favSet(1), nil)

	assert.Empty(t, result.Mismatches)
}

func TestReportMismatches_SkipsEntriesWithoutMALID(t *testing.T) {
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: false}

	animes := []Anime{makeAnime(100, 0, "No MAL", true)}

	result := fs.ReportMismatches(testCtx(), animes, nil, nil, nil)

	assert.Empty(t, result.Mismatches)
}

func TestReportMismatches_MixedAnimeAndManga(t *testing.T) {
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: false}

	animes := []Anime{makeAnime(100, 1, "Anime mismatch", true)}
	mangas := []Manga{makeManga(200, 10, "Manga mismatch", false)}

	result := fs.ReportMismatches(testCtx(), animes, mangas, favSet(), favSet(10))

	assert.Len(t, result.Mismatches, 2)
	assert.Equal(t, "anime", result.Mismatches[0].MediaType)
	assert.Equal(t, "manga", result.Mismatches[1].MediaType)
}

func TestReportMismatches_EmptyLists(t *testing.T) {
	fs := &FavoritesSync{toggler: &mockToggler{}, dryRun: false}

	result := fs.ReportMismatches(testCtx(), nil, nil, nil, nil)

	assert.Empty(t, result.Mismatches)
}

// =============================================================================
// FavoritesResult tests
// =============================================================================

func TestFavoritesResult_DefaultValues(t *testing.T) {
	result := FavoritesResult{}

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, result.Errors)
	assert.Empty(t, result.Mismatches)
}
