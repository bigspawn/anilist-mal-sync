package main

import (
	"context"
	"fmt"
	"log"
)

type App struct {
	config Config

	mal     *MyAnimeListClient
	anilist *AnilistClient

	animeUpdater        *Updater
	mangaUpdater        *Updater
	reverseAnimeUpdater *Updater
	reverseMangaUpdater *Updater
}

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
			mediaID, status, progress, score := a.GetAnilistUpdateParams()
			if err := anilistClient.UpdateAnimeEntry(ctx, mediaID, status, progress, score, "MAL to AniList Anime"); err != nil {
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
			APISearchStrategy{},
		),

		UpdateTargetBySourceFunc: func(ctx context.Context, id TargetID, src Source) error {
			m, ok := src.(Manga)
			if !ok {
				return fmt.Errorf("source is not a manga")
			}
			mediaID, status, progress, progressVolumes, score := m.GetAnilistUpdateParams()
			if err := anilistClient.UpdateMangaEntry(ctx, mediaID, status, progress, progressVolumes, score, "MAL to AniList Manga"); err != nil {
				return fmt.Errorf("error updating anilist manga: %w", err)
			}
			return nil
		},
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

func (a *App) syncAnime(ctx context.Context) error {
	log.Printf("[%s] Fetching AniList...", a.animeUpdater.Prefix)

	srcList, err := a.anilist.GetUserAnimeList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user anime list from anilist: %w", err)
	}

	log.Printf("[%s] Fetching MAL...", a.animeUpdater.Prefix)

	tgtList, err := a.mal.GetUserAnimeList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user anime list from mal: %w", err)
	}

	srcAnimes := newSourcesFromAnimes(newAnimesFromMediaListGroups(srcList))
	tgtAnimes := newTargetsFromAnimes(newAnimesFromMalUserAnimes(tgtList))

	log.Printf("[%s] Got %d from AniList", a.animeUpdater.Prefix, len(srcAnimes))
	log.Printf("[%s] Got %d from Mal", a.animeUpdater.Prefix, len(tgtAnimes))

	a.animeUpdater.Update(ctx, srcAnimes, tgtAnimes)
	a.animeUpdater.Statistics.Print(a.animeUpdater.Prefix)

	return nil
}

func (a *App) syncManga(ctx context.Context) error {
	log.Printf("[%s] Fetching AniList...", a.mangaUpdater.Prefix)

	srcList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	log.Printf("[%s] Fetching MAL...", a.mangaUpdater.Prefix)

	tgtList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMediaListGroups(srcList))
	tgts := newTargetsFromMangas(newMangasFromMalUserMangas(tgtList))

	log.Printf("[%s] Got %d from AniList", a.mangaUpdater.Prefix, len(srcs))
	log.Printf("[%s] Got %d from Mal", a.mangaUpdater.Prefix, len(tgts))

	a.mangaUpdater.Update(ctx, srcs, tgts)
	a.mangaUpdater.Statistics.Print(a.mangaUpdater.Prefix)

	return nil
}

func (a *App) reverseSyncAnime(ctx context.Context) error {
	log.Printf("[%s] Fetching MAL...", a.reverseAnimeUpdater.Prefix)

	srcList, err := a.mal.GetUserAnimeList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user anime list from mal: %w", err)
	}

	log.Printf("[%s] Fetching AniList...", a.reverseAnimeUpdater.Prefix)

	tgtList, err := a.anilist.GetUserAnimeList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user anime list from anilist: %w", err)
	}

	srcAnimes := newSourcesFromAnimes(newAnimesFromMalUserAnimes(srcList))
	tgtAnimes := newTargetsFromAnimes(newAnimesFromMediaListGroups(tgtList))

	log.Printf("[%s] Got %d from MAL", a.reverseAnimeUpdater.Prefix, len(srcAnimes))
	log.Printf("[%s] Got %d from AniList", a.reverseAnimeUpdater.Prefix, len(tgtAnimes))

	a.reverseAnimeUpdater.Update(ctx, srcAnimes, tgtAnimes)
	a.reverseAnimeUpdater.Statistics.Print(a.reverseAnimeUpdater.Prefix)

	return nil
}

func (a *App) reverseSyncManga(ctx context.Context) error {
	log.Printf("[%s] Fetching MAL...", a.reverseMangaUpdater.Prefix)

	srcList, err := a.mal.GetUserMangaList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user manga list from mal: %w", err)
	}

	log.Printf("[%s] Fetching AniList...", a.reverseMangaUpdater.Prefix)

	tgtList, err := a.anilist.GetUserMangaList(ctx)
	if err != nil {
		return fmt.Errorf("error getting user manga list from anilist: %w", err)
	}

	srcs := newSourcesFromMangas(newMangasFromMalUserMangas(srcList))
	tgts := newTargetsFromMangas(newMangasFromMediaListGroups(tgtList))

	log.Printf("[%s] Got %d from MAL", a.reverseMangaUpdater.Prefix, len(srcs))
	log.Printf("[%s] Got %d from AniList", a.reverseMangaUpdater.Prefix, len(tgts))

	a.reverseMangaUpdater.Update(ctx, srcs, tgts)
	a.reverseMangaUpdater.Statistics.Print(a.reverseMangaUpdater.Prefix)

	return nil
}
