package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rl404/verniy"
)

type App struct {
	config Config

	mal                *MyAnimeListClient
	anilist            *AnilistClient
	anilistScoreFormat verniy.ScoreFormat

	animeUpdater        *Updater
	mangaUpdater        *Updater
	reverseAnimeUpdater *Updater
	reverseMangaUpdater *Updater
}

// NewApp creates a new App instance with configured clients and updaters
//
//nolint:funlen //ok
func NewApp(ctx context.Context, config Config) (*App, error) {
	oauthMAL, err := NewMyAnimeListOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating mal oauth: %w", err)
	}

	log.Println("Got MAL token")

	malClient := NewMyAnimeListClient(ctx, oauthMAL, config.MyAnimeList.Username)

	log.Println("MAL client created")

	oauthAnilist, err := NewAnilistOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating anilist oauth: %w", err)
	}

	log.Println("Got Anilist token")

	anilistClient := NewAnilistClient(ctx, oauthAnilist, config.Anilist.Username)

	log.Println("Anilist client created")

	scoreFormat, err := anilistClient.GetUserScoreFormat(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user score format: %w", err)
	}
	log.Printf("AniList score format: %s", scoreFormat)

	animeUpdater := &Updater{
		Prefix:     "AniList to MAL Anime",
		Statistics: new(Statistics),
		IgnoreTitles: map[string]struct{}{ // in lowercase, TODO: move to config
			"scott pilgrim takes off":       {}, // this anime is not in MAL
			"bocchi the rock! recap part 2": {}, // this anime is not in MAL
		},
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := malClient.GetAnimeByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("error getting anime by id: %w", err)
					}
					ani, err := newAnimeFromMalAnime(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating anime from mal anime: %w", err)
					}
					return ani, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := malClient.GetAnimesByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("error getting anime by name: %w", err)
					}
					return newTargetsFromAnimes(newAnimesFromMalAnimes(resp)), nil
				},
			},
		),

		UpdateTargetBySourceFunc: func(ctx context.Context, id TargetID, src Source) error {
			a, ok := src.(Anime)
			if !ok {
				return fmt.Errorf("source is not an anime")
			}
			if err := malClient.UpdateAnimeByIDAndOptions(ctx, int(id), a.GetUpdateOptions()); err != nil {
				return fmt.Errorf("error updating anime by id and options: %w", err)
			}
			return nil
		},
	}

	mangaUpdater := &Updater{
		Prefix:       "AniList to MAL Manga",
		Statistics:   new(Statistics),
		IgnoreTitles: map[string]struct{}{},
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := malClient.GetMangaByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("error getting manga by id: %w", err)
					}
					manga, err := newMangaFromMalManga(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating manga from mal manga: %w", err)
					}
					return manga, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := malClient.GetMangasByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("error getting manga by name: %w", err)
					}
					return newTargetsFromMangas(newMangasFromMalMangas(resp)), nil
				},
			},
		),

		UpdateTargetBySourceFunc: func(ctx context.Context, id TargetID, src Source) error {
			m, ok := src.(Manga)
			if !ok {
				return fmt.Errorf("source is not an anime")
			}
			if err := malClient.UpdateMangaByIDAndOptions(ctx, int(id), m.GetUpdateOptions()); err != nil {
				return fmt.Errorf("error updating anime by id and options: %w", err)
			}
			return nil
		},
	}

	// Reverse updaters for MAL -> AniList sync
	reverseAnimeUpdater := &Updater{
		Prefix:       "MAL to AniList Anime",
		Statistics:   new(Statistics),
		IgnoreTitles: map[string]struct{}{},
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			MALIDStrategy{
				GetTargetByMALIDFunc: func(ctx context.Context, malID int) (Target, error) {
					resp, err := anilistClient.GetAnimeByMALID(ctx, malID, "MAL to AniList Anime")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist anime by MAL ID: %w", err)
					}
					ani, err := newAnimeFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating anime from anilist media: %w", err)
					}
					return ani, nil
				},
			},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := anilistClient.GetAnimeByID(ctx, int(id), "MAL to AniList Anime")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist anime by id: %w", err)
					}
					ani, err := newAnimeFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating anime from anilist media: %w", err)
					}
					return ani, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := anilistClient.GetAnimesByName(ctx, name, "MAL to AniList Anime")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist anime by name: %w", err)
					}
					return newTargetsFromAnimes(newAnimesFromVerniyMedias(resp)), nil
				},
			},
		),

		UpdateTargetBySourceFunc: func(ctx context.Context, id TargetID, src Source) error {
			a, ok := src.(Anime)
			if !ok {
				return fmt.Errorf("source is not an anime")
			}
			// Denormalize score from 0-10 format back to user's AniList format
			anilistScore := denormalizeScoreForAniList(a.Score, scoreFormat)
			if err := anilistClient.UpdateAnimeEntry(
				ctx,
				int(id),
				a.Status.GetAnilistStatus(),
				a.Progress,
				anilistScore,
				"MAL to AniList Anime"); err != nil {
				return fmt.Errorf("error updating anilist anime: %w", err)
			}
			return nil
		},
	}

	reverseMangaUpdater := &Updater{
		Prefix:       "MAL to AniList Manga",
		Statistics:   new(Statistics),
		IgnoreTitles: map[string]struct{}{},
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			MALIDStrategy{
				GetTargetByMALIDFunc: func(ctx context.Context, malID int) (Target, error) {
					resp, err := anilistClient.GetMangaByMALID(ctx, malID, "MAL to AniList Manga")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist manga by MAL ID: %w", err)
					}
					manga, err := newMangaFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating manga from anilist media: %w", err)
					}
					return manga, nil
				},
			},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := anilistClient.GetMangaByID(ctx, int(id), "MAL to AniList Manga")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist manga by id: %w", err)
					}
					manga, err := newMangaFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating manga from anilist media: %w", err)
					}
					return manga, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := anilistClient.GetMangasByName(ctx, name, "MAL to AniList Manga")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist manga by name: %w", err)
					}
					return newTargetsFromMangas(newMangasFromVerniyMedias(resp)), nil
				},
			},
		),

		UpdateTargetBySourceFunc: func(ctx context.Context, id TargetID, src Source) error {
			m, ok := src.(Manga)
			if !ok {
				return fmt.Errorf("source is not a manga")
			}
			// Denormalize score from 0-10 format back to user's AniList format
			anilistScore := denormalizeMangaScoreForAniList(m.Score, scoreFormat)
			if err := anilistClient.UpdateMangaEntry(
				ctx,
				int(id),
				m.Status.GetAnilistStatus(),
				m.Progress,
				m.ProgressVolumes,
				anilistScore,
				"MAL to AniList Manga"); err != nil {
				return fmt.Errorf("error updating anilist manga: %w", err)
			}
			return nil
		},
	}

	return &App{
		config:              config,
		mal:                 malClient,
		anilist:             anilistClient,
		anilistScoreFormat:  scoreFormat,
		animeUpdater:        animeUpdater,
		mangaUpdater:        mangaUpdater,
		reverseAnimeUpdater: reverseAnimeUpdater,
		reverseMangaUpdater: reverseMangaUpdater,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if *reverseDirection {
		return a.runReverseSync(ctx)
	}
	return a.runNormalSync(ctx)
}

func (a *App) runNormalSync(ctx context.Context) error {
	if *mangaSync || *allSync {
		if err := a.syncManga(ctx); err != nil {
			return fmt.Errorf("error syncing manga: %w", err)
		}
	}

	if !(*mangaSync) || *allSync {
		if err := a.syncAnime(ctx); err != nil {
			return fmt.Errorf("error syncing anime: %w", err)
		}
	}

	return nil
}

func (a *App) runReverseSync(ctx context.Context) error {
	if *mangaSync || *allSync {
		if err := a.reverseSyncManga(ctx); err != nil {
			return fmt.Errorf("error reverse syncing manga: %w", err)
		}
	}

	if !(*mangaSync) || *allSync {
		if err := a.reverseSyncAnime(ctx); err != nil {
			return fmt.Errorf("error reverse syncing anime: %w", err)
		}
	}

	return nil
}

// syncAnimeFromAnilistToMAL syncs anime from AniList to MAL
func (a *App) syncAnime(ctx context.Context) error {
	return a.performSync(ctx, "anime", false, a.animeUpdater)
}

// syncMangaFromAnilistToMAL syncs manga from AniList to MAL
func (a *App) syncManga(ctx context.Context) error {
	return a.performSync(ctx, "manga", false, a.mangaUpdater)
}

// reverseSyncAnimeFromMALToAnilist syncs anime from MAL to AniList
func (a *App) reverseSyncAnime(ctx context.Context) error {
	return a.performSync(ctx, "anime", true, a.reverseAnimeUpdater)
}

// reverseSyncMangaFromMALToAnilist syncs manga from MAL to AniList
func (a *App) reverseSyncManga(ctx context.Context) error {
	return a.performSync(ctx, "manga", true, a.reverseMangaUpdater)
}

// performSync is a generic sync function that handles both anime and manga syncing
func (a *App) performSync(ctx context.Context, mediaType string, reverse bool, updater *Updater) error {
	var srcs []Source
	var tgts []Target
	var err error

	if reverse {
		// Reverse sync: MAL -> AniList
		srcs, tgts, err = a.fetchReverseSyncData(ctx, mediaType, updater.Prefix)
	} else {
		// Normal sync: AniList -> MAL
		srcs, tgts, err = a.fetchNormalSyncData(ctx, mediaType, updater.Prefix)
	}

	if err != nil {
		return err
	}

	updater.Update(ctx, srcs, tgts)
	updater.Statistics.Print(updater.Prefix)

	return nil
}

// fetchData is a generic helper for fetching data from both services
func (a *App) fetchData(ctx context.Context, mediaType string, fromAnilist bool, prefix string) ([]Source, []Target, error) {
	if fromAnilist {
		return a.fetchFromAnilistToMAL(ctx, mediaType, prefix)
	}
	return a.fetchFromMALToAnilist(ctx, mediaType, prefix)
}

// fetchFromAnilistToMAL fetches data for AniList -> MAL sync.
// Note: this function intentionally mirrors the corresponding MAL -> AniList
// fetch function, so some duplication is expected and acceptable. We keep
// these flows separate for clarity and to make each sync direction easier
// to reason about, even though this may trigger dupl linter warnings.
func (a *App) fetchFromAnilistToMAL(ctx context.Context, mediaType string, prefix string) ([]Source, []Target, error) {
	log.Printf("[%s] Fetching AniList...", prefix)

	if mediaType == "anime" {
		srcList, err := a.anilist.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from anilist: %w", err)
		}

		log.Printf("[%s] Fetching MAL...", prefix)
		tgtList, err := a.mal.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from mal: %w", err)
		}

		srcs := newSourcesFromAnimes(newAnimesFromMediaListGroups(srcList, a.anilistScoreFormat))
		tgts := newTargetsFromAnimes(newAnimesFromMalUserAnimes(tgtList))

		log.Printf("[%s] Got %d from AniList", prefix, len(srcs))
		log.Printf("[%s] Got %d from Mal", prefix, len(tgts))

		return srcs, tgts, nil
	}

	// manga
	srcList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	log.Printf("[%s] Fetching MAL...", prefix)
	tgtList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMediaListGroups(srcList, a.anilistScoreFormat))
	tgts := newTargetsFromMangas(newMangasFromMalUserMangas(tgtList))

	log.Printf("[%s] Got %d from AniList", prefix, len(srcs))
	log.Printf("[%s] Got %d from Mal", prefix, len(tgts))

	return srcs, tgts, nil
}

// fetchFromMALToAnilist fetches data for MAL -> AniList sync.
// The structure of this function intentionally mirrors fetchFromAnilistToMAL
// to keep the two sync directions explicit and symmetrical.
func (a *App) fetchFromMALToAnilist(ctx context.Context, mediaType string, prefix string) ([]Source, []Target, error) {
	log.Printf("[%s] Fetching MAL...", prefix)

	if mediaType == "anime" {
		srcList, err := a.mal.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from mal: %w", err)
		}

		log.Printf("[%s] Fetching AniList...", prefix)
		tgtList, err := a.anilist.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from anilist: %w", err)
		}

		srcs := newSourcesFromAnimes(newAnimesFromMalUserAnimes(srcList))
		tgts := newTargetsFromAnimes(newAnimesFromMediaListGroups(tgtList, a.anilistScoreFormat))

		log.Printf("[%s] Got %d from MAL", prefix, len(srcs))
		log.Printf("[%s] Got %d from AniList", prefix, len(tgts))

		return srcs, tgts, nil
	}

	// manga
	srcList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	log.Printf("[%s] Fetching AniList...", prefix)
	tgtList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMalUserMangas(srcList))
	tgts := newTargetsFromMangas(newMangasFromMediaListGroups(tgtList, a.anilistScoreFormat))

	log.Printf("[%s] Got %d from MAL", prefix, len(srcs))
	log.Printf("[%s] Got %d from AniList", prefix, len(tgts))

	return srcs, tgts, nil
}

// fetchNormalSyncData fetches data for AniList -> MAL sync
func (a *App) fetchNormalSyncData(ctx context.Context, mediaType string, prefix string) ([]Source, []Target, error) {
	return a.fetchData(ctx, mediaType, true, prefix)
}

// fetchReverseSyncData fetches data for MAL -> AniList sync
func (a *App) fetchReverseSyncData(ctx context.Context, mediaType string, prefix string) ([]Source, []Target, error) {
	return a.fetchData(ctx, mediaType, false, prefix)
}
