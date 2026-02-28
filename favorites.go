package main

import (
	"context"
	"time"
)

// favouriteToggler abstracts the AniList ToggleFavourite mutation for testability.
type favouriteToggler interface {
	ToggleFavourite(ctx context.Context, animeID, mangaID int) error
}

// FavoritesSync handles synchronization of favorites between AniList and MAL.
// Due to API limitations, only MAL -> AniList direction can write;
// AniList -> MAL direction can only report differences.
type FavoritesSync struct {
	toggler favouriteToggler
	dryRun  bool
}

// NewFavoritesSync creates a new FavoritesSync instance.
func NewFavoritesSync(anilist *AnilistClient, dryRun bool) *FavoritesSync {
	return &FavoritesSync{
		toggler: anilist,
		dryRun:  dryRun,
	}
}

// FavoritesResult contains statistics from a favorites sync operation.
type FavoritesResult struct {
	Added      int                // Number of favorites added
	Skipped    int                // Number skipped (already in sync or no MAL ID)
	Mismatches []FavoriteMismatch // Differences found (for report-only direction)
	Errors     int                // Number of errors encountered
}

// FavoriteMismatch represents a favorite status difference between services.
type FavoriteMismatch struct {
	Title     string // Title of the anime/manga
	AniListID int    // AniList ID
	MALID     int    // MAL ID
	MediaType string // "anime" or "manga"
	OnAniList bool   // True if favorited on AniList
	OnMAL     bool   // True if favorited on MAL
}

// SyncToAniList syncs favorites from MAL to AniList.
// It adds missing favorites on AniList but does not remove favorites
// that exist on AniList but not on MAL (user may have intentionally
// favorited different items on each service).
func (f *FavoritesSync) SyncToAniList(
	ctx context.Context,
	animes []Anime,
	mangas []Manga,
	malAnimeFavs, malMangaFavs map[int]struct{},
) FavoritesResult {
	result := FavoritesResult{}

	for _, anime := range animes {
		f.syncAnimeToAniList(ctx, anime, malAnimeFavs, &result)
	}

	for _, manga := range mangas {
		f.syncMangaToAniList(ctx, manga, malMangaFavs, &result)
	}

	return result
}

func (f *FavoritesSync) syncAnimeToAniList(ctx context.Context, anime Anime, malFavs map[int]struct{}, result *FavoritesResult) {
	if anime.IDMal <= 0 {
		LogDebug(ctx, "★ [Favorites] Skipping anime %q (no MAL ID)", anime.GetTitle())
		result.Skipped++
		return
	}

	malFav := isFavorite(malFavs, anime.IDMal)
	alFav := anime.IsFavourite

	switch {
	case malFav && !alFav:
		f.addAnimeToAniList(ctx, anime, result)
	case alFav && !malFav:
		LogDebug(ctx, "★ [Favorites] Anime %q is favorited on AniList but not MAL (skipping removal)", anime.GetTitle())
		result.Skipped++
	default:
		result.Skipped++
	}
}

func (f *FavoritesSync) addAnimeToAniList(ctx context.Context, anime Anime, result *FavoritesResult) {
	if f.dryRun {
		LogInfoDryRun(ctx, "★ [Favorites] Would add anime %q (MAL ID %d, AniList ID %d) to AniList favorites",
			anime.GetTitle(), anime.IDMal, anime.IDAnilist)
		result.Added++
		return
	}

	if err := f.toggleWithRateLimit(ctx, anime.IDAnilist, 0); err != nil {
		LogWarn(ctx, "★ [Favorites] Failed to add anime %q to favorites: %v", anime.GetTitle(), err)
		result.Errors++
		return
	}

	LogInfo(ctx, "★ [Favorites] Added anime %q to AniList favorites", anime.GetTitle())
	result.Added++
}

func (f *FavoritesSync) syncMangaToAniList(ctx context.Context, manga Manga, malFavs map[int]struct{}, result *FavoritesResult) {
	if manga.IDMal <= 0 {
		LogDebug(ctx, "★ [Favorites] Skipping manga %q (no MAL ID)", manga.GetTitle())
		result.Skipped++
		return
	}

	malFav := isFavorite(malFavs, manga.IDMal)
	alFav := manga.IsFavourite

	switch {
	case malFav && !alFav:
		f.addMangaToAniList(ctx, manga, result)
	case alFav && !malFav:
		LogDebug(ctx, "★ [Favorites] Manga %q is favorited on AniList but not MAL (skipping removal)", manga.GetTitle())
		result.Skipped++
	default:
		result.Skipped++
	}
}

func (f *FavoritesSync) addMangaToAniList(ctx context.Context, manga Manga, result *FavoritesResult) {
	if f.dryRun {
		LogInfoDryRun(ctx, "★ [Favorites] Would add manga %q (MAL ID %d, AniList ID %d) to AniList favorites",
			manga.GetTitle(), manga.IDMal, manga.IDAnilist)
		result.Added++
		return
	}

	if err := f.toggleWithRateLimit(ctx, 0, manga.IDAnilist); err != nil {
		LogWarn(ctx, "★ [Favorites] Failed to add manga %q to favorites: %v", manga.GetTitle(), err)
		result.Errors++
		return
	}

	LogInfo(ctx, "★ [Favorites] Added manga %q to AniList favorites", manga.GetTitle())
	result.Added++
}

// ReportMismatches compares favorites between AniList and MAL and reports differences.
// This is used for the AniList -> MAL direction where we cannot write to MAL.
func (f *FavoritesSync) ReportMismatches(
	ctx context.Context,
	animes []Anime,
	mangas []Manga,
	malAnimeFavs, malMangaFavs map[int]struct{},
) FavoritesResult {
	result := FavoritesResult{
		Mismatches: make([]FavoriteMismatch, 0),
	}

	for _, anime := range animes {
		if anime.IDMal <= 0 {
			continue
		}
		if isFavorite(malAnimeFavs, anime.IDMal) != anime.IsFavourite {
			result.Mismatches = append(result.Mismatches, FavoriteMismatch{
				Title:     anime.GetTitle(),
				AniListID: anime.IDAnilist,
				MALID:     anime.IDMal,
				MediaType: "anime",
				OnAniList: anime.IsFavourite,
				OnMAL:     isFavorite(malAnimeFavs, anime.IDMal),
			})
		}
	}

	for _, manga := range mangas {
		if manga.IDMal <= 0 {
			continue
		}
		if isFavorite(malMangaFavs, manga.IDMal) != manga.IsFavourite {
			result.Mismatches = append(result.Mismatches, FavoriteMismatch{
				Title:     manga.GetTitle(),
				AniListID: manga.IDAnilist,
				MALID:     manga.IDMal,
				MediaType: "manga",
				OnAniList: manga.IsFavourite,
				OnMAL:     isFavorite(malMangaFavs, manga.IDMal),
			})
		}
	}

	for _, mm := range result.Mismatches {
		switch {
		case mm.OnAniList && !mm.OnMAL:
			LogInfo(ctx, "★ [Favorites] %s %q is only on AniList", mm.MediaType, mm.Title)
		case !mm.OnAniList && mm.OnMAL:
			LogInfo(ctx, "★ [Favorites] %s %q is only on MAL", mm.MediaType, mm.Title)
		default:
			LogInfo(ctx, "★ [Favorites] %s %q is in unknown state", mm.MediaType, mm.Title)
		}
	}

	return result
}

// isFavorite checks if a MAL ID is in the favorites set.
func isFavorite(favSet map[int]struct{}, malID int) bool {
	_, ok := favSet[malID]
	return ok
}

// toggleWithRateLimit calls ToggleFavourite with rate limiting to avoid hitting AniList limits.
// AniList allows ~90 requests/minute, so we add a small delay between calls (~700ms).
func (f *FavoritesSync) toggleWithRateLimit(ctx context.Context, animeID, mangaID int) error {
	const rateLimitDelay = 700 * time.Millisecond
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(rateLimitDelay):
	}
	return f.toggler.ToggleFavourite(ctx, animeID, mangaID)
}
