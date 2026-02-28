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
	jikan   *JikanClient
	dryRun  bool
}

// NewFavoritesSync creates a new FavoritesSync instance.
func NewFavoritesSync(anilist *AnilistClient, jikan *JikanClient, dryRun bool) *FavoritesSync {
	return &FavoritesSync{
		toggler: anilist,
		jikan:   jikan,
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
//
// Parameters:
//   - animes: List of anime entries with both AniList and MAL IDs
//   - mangas: List of manga entries with both AniList and MAL IDs
//   - malAnimeFavs: Set of MAL IDs for anime favorites on MAL
//   - malMangaFavs: Set of MAL IDs for manga favorites on MAL
func (f *FavoritesSync) SyncToAniList(
	ctx context.Context,
	animes []Anime,
	mangas []Manga,
	malAnimeFavs, malMangaFavs map[int]struct{},
) FavoritesResult {
	result := FavoritesResult{}

	// Sync anime favorites
	for _, anime := range animes {
		if anime.IDMal <= 0 {
			LogDebug(ctx, "[Favorites] Skipping anime %q (no MAL ID)", anime.GetTitle())
			result.Skipped++
			continue
		}

		malFav := isFavorite(malAnimeFavs, anime.IDMal)
		alFav := anime.IsFavourite

		// MAL favorite but not AniList favorite -> add to AniList
		if malFav && !alFav {
			if f.dryRun {
				LogInfoDryRun(ctx, "[Favorites] Would add anime %q (MAL ID %d, AniList ID %d) to AniList favorites",
					anime.GetTitle(), anime.IDMal, anime.IDAnilist)
				result.Added++
			} else {
				if err := f.toggleWithRateLimit(ctx, anime.IDAnilist, 0); err != nil {
					LogWarn(ctx, "[Favorites] Failed to add anime %q to favorites: %v", anime.GetTitle(), err)
					result.Errors++
				} else {
					LogInfoUpdate(ctx, "[Favorites] Added anime %q to AniList favorites", anime.GetTitle())
					result.Added++
				}
			}
			continue
		}

		// AniList favorite but not MAL favorite -> skip (don't remove)
		if alFav && !malFav {
			LogDebug(ctx, "[Favorites] Anime %q is favorited on AniList but not MAL (skipping removal)", anime.GetTitle())
			result.Skipped++
			continue
		}

		// Both favorited or neither -> already in sync
		result.Skipped++
	}

	// Sync manga favorites
	for _, manga := range mangas {
		if manga.IDMal <= 0 {
			LogDebug(ctx, "[Favorites] Skipping manga %q (no MAL ID)", manga.GetTitle())
			result.Skipped++
			continue
		}

		malFav := isFavorite(malMangaFavs, manga.IDMal)
		alFav := manga.IsFavourite

		// MAL favorite but not AniList favorite -> add to AniList
		if malFav && !alFav {
			if f.dryRun {
				LogInfoDryRun(ctx, "[Favorites] Would add manga %q (MAL ID %d, AniList ID %d) to AniList favorites",
					manga.GetTitle(), manga.IDMal, manga.IDAnilist)
				result.Added++
			} else {
				if err := f.toggleWithRateLimit(ctx, 0, manga.IDAnilist); err != nil {
					LogWarn(ctx, "[Favorites] Failed to add manga %q to favorites: %v", manga.GetTitle(), err)
					result.Errors++
				} else {
					LogInfoUpdate(ctx, "[Favorites] Added manga %q to AniList favorites", manga.GetTitle())
					result.Added++
				}
			}
			continue
		}

		// AniList favorite but not MAL favorite -> skip (don't remove)
		if alFav && !malFav {
			LogDebug(ctx, "[Favorites] Manga %q is favorited on AniList but not MAL (skipping removal)", manga.GetTitle())
			result.Skipped++
			continue
		}

		// Both favorited or neither -> already in sync
		result.Skipped++
	}

	return result
}

// ReportMismatches compares favorites between AniList and MAL and reports differences.
// This is used for the AniList -> MAL direction where we cannot write to MAL.
//
// Parameters:
//   - animes: List of anime entries with both AniList and MAL IDs
//   - mangas: List of manga entries with both AniList and MAL IDs
//   - malAnimeFavs: Set of MAL IDs for anime favorites on MAL
//   - malMangaFavs: Set of MAL IDs for manga favorites on MAL
func (f *FavoritesSync) ReportMismatches(
	ctx context.Context,
	animes []Anime,
	mangas []Manga,
	malAnimeFavs, malMangaFavs map[int]struct{},
) FavoritesResult {
	result := FavoritesResult{
		Mismatches: make([]FavoriteMismatch, 0),
	}

	// Compare anime favorites
	for _, anime := range animes {
		if anime.IDMal <= 0 {
			continue
		}

		malFav := isFavorite(malAnimeFavs, anime.IDMal)
		alFav := anime.IsFavourite

		if malFav != alFav {
			result.Mismatches = append(result.Mismatches, FavoriteMismatch{
				Title:     anime.GetTitle(),
				AniListID: anime.IDAnilist,
				MALID:     anime.IDMal,
				MediaType: "anime",
				OnAniList: alFav,
				OnMAL:     malFav,
			})
		}
	}

	// Compare manga favorites
	for _, manga := range mangas {
		if manga.IDMal <= 0 {
			continue
		}

		malFav := isFavorite(malMangaFavs, manga.IDMal)
		alFav := manga.IsFavourite

		if malFav != alFav {
			result.Mismatches = append(result.Mismatches, FavoriteMismatch{
				Title:     manga.GetTitle(),
				AniListID: manga.IDAnilist,
				MALID:     manga.IDMal,
				MediaType: "manga",
				OnAniList: alFav,
				OnMAL:     malFav,
			})
		}
	}

	// Log mismatches
	for _, mm := range result.Mismatches {
		direction := "only on"
		if mm.OnAniList && !mm.OnMAL {
			direction = "only on AniList"
		} else if !mm.OnAniList && mm.OnMAL {
			direction = "only on MAL"
		}
		LogInfo(ctx, "[Favorites] %s %q is %s", mm.MediaType, mm.Title, direction)
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
	// Rate limit: ~90 req/min = ~700ms between requests
	time.Sleep(700 * time.Millisecond)
	return f.toggler.ToggleFavourite(ctx, animeID, mangaID)
}
