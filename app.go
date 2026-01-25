package main

import (
	"context"
	"fmt"

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
//nolint:funlen // Function creates multiple services and updaters - acceptable complexity
func NewApp(ctx context.Context, config Config) (*App, error) {
	LogStage(ctx, "Initializing...")

	oauthMAL, err := NewMyAnimeListOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating mal oauth: %w", err)
	}
	LogDebug(ctx, "Got MAL token")

	malClient := NewMyAnimeListClient(ctx, oauthMAL, config.MyAnimeList.Username, config.GetHTTPTimeout(), *verbose)
	LogDebug(ctx, "MAL client created")

	oauthAnilist, err := NewAnilistOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating anilist oauth: %w", err)
	}
	LogDebug(ctx, "Got Anilist token")

	anilistClient := NewAnilistClient(ctx, oauthAnilist, config.Anilist.Username, config.GetHTTPTimeout(), *verbose)
	LogDebug(ctx, "Anilist client created")

	scoreFormat, err := anilistClient.GetUserScoreFormat(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user score format: %w", err)
	}
	LogDebug(ctx, "AniList score format: %s", scoreFormat)

	// Create services
	malAnimeService := NewMALAnimeService(malClient)
	malMangaService := NewMALMangaService(malClient)
	anilistAnimeService := NewAniListAnimeService(anilistClient, scoreFormat)
	anilistMangaService := NewAniListMangaService(anilistClient, scoreFormat)

	// Default ignore titles
	defaultIgnoreTitles := map[string]struct{}{
		"scott pilgrim takes off":       {},
		"bocchi the rock! recap part 2": {},
	}

	// Create updaters
	animeUpdater := &Updater{
		Prefix:       "AniList to MAL Anime",
		Service:      malAnimeService,
		Statistics:   NewStatistics(),
		IgnoreTitles: defaultIgnoreTitles,
		ForceSync:    *forceSync,
		DryRun:       *dryRun,
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{Service: malAnimeService},
		),
	}

	mangaUpdater := &Updater{
		Prefix:       "AniList to MAL Manga",
		Service:      malMangaService,
		Statistics:   NewStatistics(),
		IgnoreTitles: map[string]struct{}{},
		ForceSync:    *forceSync,
		DryRun:       *dryRun,
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{Service: malMangaService},
		),
	}

	reverseAnimeUpdater := &Updater{
		Prefix:       "MAL to AniList Anime",
		Service:      anilistAnimeService,
		Statistics:   NewStatistics(),
		IgnoreTitles: map[string]struct{}{},
		ForceSync:    *forceSync,
		DryRun:       *dryRun,
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			MALIDStrategy{Service: anilistAnimeService},
			APISearchStrategy{Service: anilistAnimeService},
		),
	}

	reverseMangaUpdater := &Updater{
		Prefix:       "MAL to AniList Manga",
		Service:      anilistMangaService,
		Statistics:   NewStatistics(),
		IgnoreTitles: map[string]struct{}{},
		ForceSync:    *forceSync,
		DryRun:       *dryRun,
		StrategyChain: NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			MALIDStrategy{Service: anilistMangaService},
			APISearchStrategy{Service: anilistMangaService},
		),
	}

	LogInfoSuccess(ctx, "Initialization complete")

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
	updater.Statistics.Print(ctx, updater.Prefix)
	updater.Statistics.Reset()

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
	LogDebug(ctx, "[%s] Fetching AniList...", prefix)

	if mediaType == "anime" {
		srcList, err := a.anilist.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from anilist: %w", err)
		}

		LogDebug(ctx, "[%s] Fetching MAL...", prefix)
		tgtList, err := a.mal.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from mal: %w", err)
		}

		srcs := newSourcesFromAnimes(newAnimesFromMediaListGroups(srcList, a.anilistScoreFormat))
		tgts := newTargetsFromAnimes(newAnimesFromMalUserAnimes(tgtList))

		LogDebug(ctx, "[%s] Got %d from AniList, %d from MAL", prefix, len(srcs), len(tgts))

		return srcs, tgts, nil
	}

	// manga
	srcList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	LogDebug(ctx, "[%s] Fetching MAL...", prefix)
	tgtList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMediaListGroups(srcList, a.anilistScoreFormat))
	tgts := newTargetsFromMangas(newMangasFromMalUserMangas(tgtList))

	LogDebug(ctx, "[%s] Got %d from AniList, %d from MAL", prefix, len(srcs), len(tgts))

	return srcs, tgts, nil
}

// fetchFromMALToAnilist fetches data for MAL -> AniList sync.
// The structure of this function intentionally mirrors fetchFromAnilistToMAL
// to keep the two sync directions explicit and symmetrical.
func (a *App) fetchFromMALToAnilist(ctx context.Context, mediaType string, prefix string) ([]Source, []Target, error) {
	LogDebug(ctx, "[%s] Fetching MAL...", prefix)

	if mediaType == "anime" {
		srcList, err := a.mal.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from mal: %w", err)
		}

		LogDebug(ctx, "[%s] Fetching AniList...", prefix)
		tgtList, err := a.anilist.GetUserAnimeList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting user anime list from anilist: %w", err)
		}

		srcs := newSourcesFromAnimes(newAnimesFromMalUserAnimes(srcList))
		tgts := newTargetsFromAnimes(newAnimesFromMediaListGroups(tgtList, a.anilistScoreFormat))

		LogDebug(ctx, "[%s] Got %d from MAL, %d from AniList", prefix, len(srcs), len(tgts))

		return srcs, tgts, nil
	}

	// manga
	srcList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	LogDebug(ctx, "[%s] Fetching AniList...", prefix)
	tgtList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMalUserMangas(srcList))
	tgts := newTargetsFromMangas(newMangasFromMediaListGroups(tgtList, a.anilistScoreFormat))

	LogDebug(ctx, "[%s] Got %d from MAL, %d from AniList", prefix, len(srcs), len(tgts))

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
