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

	animeUpdater *Updater
	mangaUpdater *Updater
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
		Prefix:     "Anime",
		Statistics: new(Statistics),
		IgnoreTitles: map[string]struct{}{ // in lowercase, TODO: move to config
			"scott pilgrim takes off":       {}, // this anime is not in MAL
			"bocchi the rock! recap part 2": {}, // this anime is not in MAL
		},

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
		Prefix:       "Manga",
		Statistics:   new(Statistics),
		IgnoreTitles: map[string]struct{}{},

		GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
			resp, err := malClient.GetMangaByID(ctx, int(id))
			if err != nil {
				return nil, fmt.Errorf("error getting anime by id: %w", err)
			}
			ani, err := newMangaFromMalManga(*resp)
			if err != nil {
				return nil, fmt.Errorf("error creating anime from mal anime: %w", err)
			}
			return ani, nil
		},

		GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
			resp, err := malClient.GetMangasByName(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("error getting anime by name: %w", err)
			}
			return newTargetsFromMangas(newMangasFromMalMangas(resp)), nil
		},

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

	return &App{
		config:       config,
		mal:          malClient,
		anilist:      anilistClient,
		animeUpdater: animeUpdater,
		mangaUpdater: mangaUpdater,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
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
