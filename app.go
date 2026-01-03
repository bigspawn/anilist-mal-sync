package main

import (
	"context"
	"fmt"
	"log"

	syncer "github.com/bigspawn/anilist-mal-sync/internal/sync"
)

type App struct {
	config Config

	mal     *MyAnimeListClient
	anilist *AnilistClient

	animeUpdater        *syncer.Updater
	mangaUpdater        *syncer.Updater
	reverseAnimeUpdater *syncer.Updater
	reverseMangaUpdater *syncer.Updater
}

// NewApp creates a new App instance with configured clients and updaters
//
//nolint:funlen //ok
func NewApp(ctx context.Context, config Config, forceSync bool, dryRun bool, verbose bool) (*App, error) {
	// Apply config-driven defaults
	EnableScoreNormalization = *config.ScoreNormalization

	// Wire flags to syncer package
	syncer.ForceSync = forceSync
	syncer.DryRun = dryRun
	syncer.Verbose = verbose

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

	animeUpdater := syncer.NewUpdater(
		"AniList to MAL Anime",
		new(syncer.Statistics),
		map[string]struct{}{ // in lowercase, TODO: move to config
			"scott pilgrim takes off":       {}, // this anime is not in MAL
			"bocchi the rock! recap part 2": {}, // this anime is not in MAL
		},
		syncer.NewStrategyChain(
			syncer.IDStrategy{},
			syncer.TitleStrategy{},
			syncer.APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id syncer.TargetID) (syncer.Target, error) {
					resp, err := malClient.GetAnimeByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("error getting anime by id: %w", err)
					}
					ani, err := newAnimeFromMalAnime(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating anime from mal anime: %w", err)
					}
					return &targetAdapter{t: ani}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]syncer.Target, error) {
					resp, err := malClient.GetAnimesByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("error getting anime by name: %w", err)
					}
					return wrapTargets(newTargetsFromAnimes(newAnimesFromMalAnimes(resp))), nil
				},
			},
		),
	)

	animeUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id syncer.TargetID, src syncer.Source) error {
		sa, ok := src.(*sourceAdapter)
		if !ok {
			return fmt.Errorf("source is not an adapter")
		}
		a, ok := sa.s.(Anime)
		if !ok {
			return fmt.Errorf("source is not an anime")
		}
		if err := malClient.UpdateAnimeByIDAndOptions(ctx, int(id), a.GetUpdateOptions()); err != nil {
			return fmt.Errorf("error updating anime by id and options: %w", err)
		}
		return nil
	}

	mangaUpdater := syncer.NewUpdater(
		"AniList to MAL Manga",
		new(syncer.Statistics),
		map[string]struct{}{},
		syncer.NewStrategyChain(
			syncer.IDStrategy{},
			syncer.TitleStrategy{},
			syncer.APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id syncer.TargetID) (syncer.Target, error) {
					resp, err := malClient.GetMangaByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("error getting manga by id: %w", err)
					}
					manga, err := newMangaFromMalManga(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating manga from mal manga: %w", err)
					}
					return &targetAdapter{t: manga}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]syncer.Target, error) {
					resp, err := malClient.GetMangasByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("error getting manga by name: %w", err)
					}
					return wrapTargets(newTargetsFromMangas(newMangasFromMalMangas(resp))), nil
				},
			},
		),
	)

	mangaUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id syncer.TargetID, src syncer.Source) error {
		sa, ok := src.(*sourceAdapter)
		if !ok {
			return fmt.Errorf("source is not an adapter")
		}
		m, ok := sa.s.(Manga)
		if !ok {
			return fmt.Errorf("source is not a manga")
		}
		if err := malClient.UpdateMangaByIDAndOptions(ctx, int(id), m.GetUpdateOptions()); err != nil {
			return fmt.Errorf("error updating manga by id and options: %w", err)
		}
		return nil
	}

	// Reverse updaters for MAL -> AniList sync
	reverseAnimeUpdater := syncer.NewUpdater(
		"MAL to AniList Anime",
		new(syncer.Statistics),
		map[string]struct{}{},
		syncer.NewStrategyChain(
			syncer.IDStrategy{},
			syncer.TitleStrategy{},
			syncer.APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id syncer.TargetID) (syncer.Target, error) {
					resp, err := anilistClient.GetAnimeByID(ctx, int(id), "MAL to AniList Anime")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist anime by id: %w", err)
					}
					ani, err := newAnimeFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating anime from anilist media: %w", err)
					}
					return &targetAdapter{t: ani}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]syncer.Target, error) {
					resp, err := anilistClient.GetAnimesByName(ctx, name, "MAL to AniList Anime")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist anime by name: %w", err)
					}
					return wrapTargets(newTargetsFromAnimes(newAnimesFromVerniyMedias(resp))), nil
				},
			},
		),
	)

	reverseAnimeUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id syncer.TargetID, src syncer.Source) error {
		sa, ok := src.(*sourceAdapter)
		if !ok {
			return fmt.Errorf("source is not an adapter")
		}
		a, ok := sa.s.(Anime)
		if !ok {
			return fmt.Errorf("source is not an anime")
		}
		if err := anilistClient.UpdateAnimeEntry(
			ctx,
			int(id),
			a.Status.GetAnilistStatus(),
			a.Progress,
			int(a.Score),
			"MAL to AniList Anime"); err != nil {
			return fmt.Errorf("error updating anilist anime: %w", err)
		}
		return nil
	}

	reverseMangaUpdater := syncer.NewUpdater(
		"MAL to AniList Manga",
		new(syncer.Statistics),
		map[string]struct{}{},
		syncer.NewStrategyChain(
			syncer.IDStrategy{},
			syncer.TitleStrategy{},
			syncer.APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id syncer.TargetID) (syncer.Target, error) {
					resp, err := anilistClient.GetMangaByID(ctx, int(id), "MAL to AniList Manga")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist manga by id: %w", err)
					}
					manga, err := newMangaFromVerniyMedia(*resp)
					if err != nil {
						return nil, fmt.Errorf("error creating manga from anilist media: %w", err)
					}
					return &targetAdapter{t: manga}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]syncer.Target, error) {
					resp, err := anilistClient.GetMangasByName(ctx, name, "MAL to AniList Manga")
					if err != nil {
						return nil, fmt.Errorf("error getting anilist manga by name: %w", err)
					}
					return wrapTargets(newTargetsFromMangas(newMangasFromVerniyMedias(resp))), nil
				},
			},
		),
	)

	reverseMangaUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id syncer.TargetID, src syncer.Source) error {
		sa, ok := src.(*sourceAdapter)
		if !ok {
			return fmt.Errorf("source is not an adapter")
		}
		m, ok := sa.s.(Manga)
		if !ok {
			return fmt.Errorf("source is not a manga")
		}
		if err := anilistClient.UpdateMangaEntry(
			ctx,
			int(id),
			m.Status.GetAnilistStatus(),
			m.Progress,
			m.ProgressVolumes,
			int(m.Score),
			"MAL to AniList Manga"); err != nil {
			return fmt.Errorf("error updating anilist manga: %w", err)
		}
		return nil
	}

	return &App{
		config:              config,
		mal:                 malClient,
		anilist:             anilistClient,
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
func (a *App) performSync(ctx context.Context, mediaType string, reverse bool, updater *syncer.Updater) error {
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

	updater.Update(ctx, wrapSources(srcs), wrapTargets(tgts))
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

// fetchFromAnilistToMAL fetches data for AniList -> MAL sync
//
//nolint:dupl // ok
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

		srcs := newSourcesFromAnimes(newAnimesFromMediaListGroups(srcList))
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

	srcs := newSourcesFromMangas(newMangasFromMediaListGroups(srcList))
	tgts := newTargetsFromMangas(newMangasFromMalUserMangas(tgtList))

	log.Printf("[%s] Got %d from AniList", prefix, len(srcs))
	log.Printf("[%s] Got %d from Mal", prefix, len(tgts))

	return srcs, tgts, nil
}

// fetchFromMALToAnilist fetches data for MAL -> AniList sync
//
//nolint:dupl // ok
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
		tgts := newTargetsFromAnimes(newAnimesFromMediaListGroups(tgtList))

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
	tgts := newTargetsFromMangas(newMangasFromMediaListGroups(tgtList))

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

// Adapters to bridge root Source/Target to syncer.Source/syncer.Target

type targetAdapter struct {
	t Target
}

func (ta *targetAdapter) GetTargetID() syncer.TargetID {
	return syncer.TargetID(ta.t.GetTargetID())
}

func (ta *targetAdapter) GetTitle() string {
	return ta.t.GetTitle()
}

func (ta *targetAdapter) String() string {
	return ta.t.String()
}

type sourceAdapter struct {
	s Source
}

func (sa *sourceAdapter) GetStatusString() string {
	return sa.s.GetStatusString()
}

func (sa *sourceAdapter) GetTargetID() syncer.TargetID {
	return syncer.TargetID(sa.s.GetTargetID())
}

func (sa *sourceAdapter) GetTitle() string {
	return sa.s.GetTitle()
}

func (sa *sourceAdapter) GetStringDiffWithTarget(t syncer.Target) string {
	// try to unwrap targetAdapter
	if ta, ok := t.(*targetAdapter); ok {
		return sa.s.GetStringDiffWithTarget(ta.t)
	}
	// fallback: best-effort string compare
	return sa.s.GetStringDiffWithTarget(nil)
}

func (sa *sourceAdapter) SameProgressWithTarget(t syncer.Target) bool {
	if ta, ok := t.(*targetAdapter); ok {
		return sa.s.SameProgressWithTarget(ta.t)
	}
	return false
}

func (sa *sourceAdapter) SameTypeWithTarget(t syncer.Target) bool {
	if ta, ok := t.(*targetAdapter); ok {
		return sa.s.SameTypeWithTarget(ta.t)
	}
	return false
}

func (sa *sourceAdapter) SameTitleWithTarget(t syncer.Target) bool {
	if ta, ok := t.(*targetAdapter); ok {
		return sa.s.SameTitleWithTarget(ta.t)
	}
	return false
}

func (sa *sourceAdapter) String() string {
	return sa.s.String()
}

func wrapTargets(ts []Target) []syncer.Target {
	res := make([]syncer.Target, len(ts))
	for i, t := range ts {
		res[i] = &targetAdapter{t: t}
	}
	return res
}

func wrapSources(ss []Source) []syncer.Source {
	res := make([]syncer.Source, len(ss))
	for i, s := range ss {
		res[i] = &sourceAdapter{s: s}
	}
	return res
}
