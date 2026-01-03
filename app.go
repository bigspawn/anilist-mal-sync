package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/rl404/verniy"
)

type App struct {
	config Config
	opts   SyncOptions

	mal     *MyAnimeListClient
	anilist *AnilistClient

	animeUpdater        *Updater
	mangaUpdater        *Updater
	reverseAnimeUpdater *Updater
	reverseMangaUpdater *Updater
}

// NewApp creates a new App instance with configured clients and updaters
//
//nolint:funlen //ok
func NewApp(ctx context.Context, config Config, forceSync bool, dryRun bool, verbose bool, mangaSync bool, allSync bool, direction string) (*App, error) {
	opts := NewSyncOptions(forceSync, dryRun, verbose, mangaSync, allSync, direction, config)

	oauthMAL, err := NewMyAnimeListOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MyAnimeList OAuth client: %w", err)
	}

	log.Println("Got MAL token")

	malClient := NewMyAnimeListClient(ctx, oauthMAL, config.MyAnimeList.Username)

	log.Println("MAL client created")

	oauthAnilist, err := NewAnilistOAuth(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AniList OAuth client: %w", err)
	}

	log.Println("Got Anilist token")

	anilistClient := NewAnilistClient(ctx, oauthAnilist, config.Anilist.Username)

	log.Println("Anilist client created")

	ignoreAnimeTitles := buildIgnoreTitlesMap(config.GetIgnoreAnimeTitlesMap(), DefaultIgnoreAnimeTitles)
	ignoreMangaTitles := buildIgnoreTitlesMap(config.GetIgnoreMangaTitlesMap(), DefaultIgnoreMangaTitles)

	animeUpdater := NewUpdater(
		UpdaterPrefixAnilistToMALAnime,
		new(Statistics),
		ignoreAnimeTitles,
		NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := malClient.GetAnimeByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("failed to get anime by ID %d from MAL: %w", id, err)
					}
					malAnime, ok := safeDerefPtr(resp)
					if !ok {
						return nil, fmt.Errorf("GetAnimeByID returned nil for ID %d", id)
					}
					ani, err := newAnimeFromMalAnime(malAnime, false)
					if err != nil {
						return nil, fmt.Errorf("failed to convert MAL anime (ID: %d) to internal format: %w", id, err)
					}
					return &targetAdapter{t: ani}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := malClient.GetAnimesByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("failed to search anime by name '%s' in MAL: %w", name, err)
					}
					return wrapTargets(newTargetsFromAnimes(newAnimesFromMalAnimes(resp, false))), nil
				},
			},
		),
		opts.Verbose,
		opts.DryRun,
		opts.ForceSync,
	)

	animeUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id TargetID, src Source) error {
		unwrapped, ok := safeUnwrapSourceAdapter(src)
		if !ok {
			return fmt.Errorf("internal error: failed to unwrap sourceAdapter (type: %T)", src)
		}
		a, ok := unwrapped.(Anime)
		if !ok {
			return fmt.Errorf("internal error: source is not an Anime (type: %T)", unwrapped)
		}
		if err := malClient.UpdateAnimeByIDAndOptions(ctx, int(id), a.GetUpdateOptions()); err != nil {
			return fmt.Errorf("failed to update anime (MAL ID: %d, title: %s): %w", id, a.GetTitle(), err)
		}
		return nil
	}

	mangaUpdater := NewUpdater(
		UpdaterPrefixAnilistToMALManga,
		new(Statistics),
		ignoreMangaTitles,
		NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := malClient.GetMangaByID(ctx, int(id))
					if err != nil {
						return nil, fmt.Errorf("failed to get manga by ID %d from MAL: %w", id, err)
					}
					malManga, ok := safeDerefPtr(resp)
					if !ok {
						return nil, fmt.Errorf("GetMangaByID returned nil for ID %d", id)
					}
					manga, err := newMangaFromMalManga(malManga, false)
					if err != nil {
						return nil, fmt.Errorf("failed to convert MAL manga (ID: %d) to internal format: %w", id, err)
					}
					return &targetAdapter{t: manga}, nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := malClient.GetMangasByName(ctx, name)
					if err != nil {
						return nil, fmt.Errorf("failed to search manga by name '%s' in MAL: %w", name, err)
					}
					return wrapTargets(newTargetsFromMangas(newMangasFromMalMangas(resp, false))), nil
				},
			},
		),
		opts.Verbose,
		opts.DryRun,
		opts.ForceSync,
	)

	mangaUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id TargetID, src Source) error {
		unwrapped, ok := safeUnwrapSourceAdapter(src)
		if !ok {
			return fmt.Errorf("internal error: failed to unwrap sourceAdapter (type: %T)", src)
		}
		m, ok := unwrapped.(Manga)
		if !ok {
			return fmt.Errorf("internal error: source is not a Manga (type: %T)", unwrapped)
		}
		if err := malClient.UpdateMangaByIDAndOptions(ctx, int(id), m.GetUpdateOptions()); err != nil {
			return fmt.Errorf("failed to update manga (MAL ID: %d, title: %s): %w", id, m.GetTitle(), err)
		}
		return nil
	}

	// Reverse updaters for MAL -> AniList sync
	reverseAnimeUpdater := NewUpdater(
		UpdaterPrefixMALToAnilistAnime,
		new(Statistics),
		ignoreAnimeTitles,
		NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := anilistClient.GetAnimeByID(ctx, int(id), UpdaterPrefixMALToAnilistAnime)
					if err != nil {
						return nil, fmt.Errorf("failed to get anime by ID %d from AniList: %w", id, err)
					}
					verniyMedia, ok := safeDerefPtr(resp)
					if !ok {
						return nil, fmt.Errorf("GetAnimeByID returned nil for ID %d", id)
					}
					ani, err := newAnimeFromVerniyMedia(verniyMedia, false)
					if err != nil {
						return nil, fmt.Errorf("failed to convert AniList anime (ID: %d) to internal format: %w", id, err)
					}
					return &targetAdapter{t: ani}, nil
				},
				GetTargetsByMALIDFunc: func(ctx context.Context, malID int) ([]Target, error) {
					resp, err := anilistClient.GetAnimesByMALID(ctx, malID, UpdaterPrefixMALToAnilistAnime)
					if err != nil {
						return nil, fmt.Errorf("failed to search anime by MAL ID %d in AniList: %w", malID, err)
					}
					return wrapTargets(newTargetsFromAnimes(newAnimesFromVerniyMedias(resp, false))), nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := anilistClient.GetAnimesByName(ctx, name, UpdaterPrefixMALToAnilistAnime)
					if err != nil {
						return nil, fmt.Errorf("failed to search anime by name '%s' in AniList: %w", name, err)
					}
					return wrapTargets(newTargetsFromAnimes(newAnimesFromVerniyMedias(resp, false))), nil
				},
			},
		),
		opts.Verbose,
		opts.DryRun,
		opts.ForceSync,
	)

	reverseAnimeUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id TargetID, src Source) error {
		unwrapped, ok := safeUnwrapSourceAdapter(src)
		if !ok {
			return fmt.Errorf("internal error: failed to unwrap sourceAdapter (type: %T)", src)
		}
		a, ok := unwrapped.(Anime)
		if !ok {
			return fmt.Errorf("internal error: source is not an Anime (type: %T)", unwrapped)
		}
		if err := anilistClient.UpdateAnimeEntry(
			ctx,
			int(id),
			a.Status.GetAnilistStatus(),
			a.Progress,
			int(a.Score),
			UpdaterPrefixMALToAnilistAnime); err != nil {
			return fmt.Errorf("failed to update anime (AniList ID: %d, title: %s): %w", id, a.GetTitle(), err)
		}
		return nil
	}

	reverseMangaUpdater := NewUpdater(
		UpdaterPrefixMALToAnilistManga,
		new(Statistics),
		ignoreMangaTitles,
		NewStrategyChain(
			IDStrategy{},
			TitleStrategy{},
			APISearchStrategy{
				GetTargetByIDFunc: func(ctx context.Context, id TargetID) (Target, error) {
					resp, err := anilistClient.GetMangaByID(ctx, int(id), UpdaterPrefixMALToAnilistManga)
					if err != nil {
						return nil, fmt.Errorf("failed to get manga by ID %d from AniList: %w", id, err)
					}
					verniyMedia, ok := safeDerefPtr(resp)
					if !ok {
						return nil, fmt.Errorf("GetMangaByID returned nil for ID %d", id)
					}
					manga, err := newMangaFromVerniyMedia(verniyMedia, false)
					if err != nil {
						return nil, fmt.Errorf("failed to convert AniList manga (ID: %d) to internal format: %w", id, err)
					}
					return &targetAdapter{t: manga}, nil
				},
				GetTargetsByMALIDFunc: func(ctx context.Context, malID int) ([]Target, error) {
					resp, err := anilistClient.GetMangasByMALID(ctx, malID, UpdaterPrefixMALToAnilistManga)
					if err != nil {
						return nil, fmt.Errorf("failed to search manga by MAL ID %d in AniList: %w", malID, err)
					}
					return wrapTargets(newTargetsFromMangas(newMangasFromVerniyMedias(resp, false))), nil
				},
				GetTargetsByNameFunc: func(ctx context.Context, name string) ([]Target, error) {
					resp, err := anilistClient.GetMangasByName(ctx, name, UpdaterPrefixMALToAnilistManga)
					if err != nil {
						return nil, fmt.Errorf("failed to search manga by name '%s' in AniList: %w", name, err)
					}
					return wrapTargets(newTargetsFromMangas(newMangasFromVerniyMedias(resp, false))), nil
				},
			},
		),
		opts.Verbose,
		opts.DryRun,
		opts.ForceSync,
	)

	reverseMangaUpdater.UpdateTargetBySourceFunc = func(ctx context.Context, id TargetID, src Source) error {
		unwrapped, ok := safeUnwrapSourceAdapter(src)
		if !ok {
			return fmt.Errorf("internal error: failed to unwrap sourceAdapter (type: %T)", src)
		}
		m, ok := unwrapped.(Manga)
		if !ok {
			return fmt.Errorf("internal error: source is not a Manga (type: %T)", unwrapped)
		}
		if err := anilistClient.UpdateMangaEntry(
			ctx,
			int(id),
			m.Status.GetAnilistStatus(),
			m.Progress,
			m.ProgressVolumes,
			int(m.Score),
			UpdaterPrefixMALToAnilistManga); err != nil {
			return fmt.Errorf("failed to update manga (AniList ID: %d, title: %s): %w", id, m.GetTitle(), err)
		}
		return nil
	}

	return &App{
		config:              config,
		opts:                opts,
		mal:                 malClient,
		anilist:             anilistClient,
		animeUpdater:        animeUpdater,
		mangaUpdater:        mangaUpdater,
		reverseAnimeUpdater: reverseAnimeUpdater,
		reverseMangaUpdater: reverseMangaUpdater,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a.opts.IsBidirectional() {
		log.Println("Starting bidirectional sync (AniList ↔ MAL)...")

		// Run normal sync first (AniList → MAL)
		log.Println("Phase 1: Syncing AniList → MAL...")
		if err := a.runNormalSync(ctx); err != nil {
			return fmt.Errorf("failed normal sync (AniList → MAL): %w", err)
		}

		// Run reverse sync second (MAL → AniList)
		log.Println("Phase 2: Syncing MAL → AniList...")
		if err := a.runReverseSync(ctx); err != nil {
			return fmt.Errorf("failed reverse sync (MAL → AniList): %w", err)
		}

		log.Println("Bidirectional sync completed successfully!")
		return nil
	}

	if a.opts.IsReverseDirection() {
		return a.runReverseSync(ctx)
	}
	return a.runNormalSync(ctx)
}

func (a *App) runNormalSync(ctx context.Context) error {
	if a.opts.IsMangaSync() {
		if err := a.syncManga(ctx); err != nil {
			return fmt.Errorf("failed to sync manga (AniList -> MAL): %w", err)
		}
	}

	if a.opts.IsAnimeSync() {
		if err := a.syncAnime(ctx); err != nil {
			return fmt.Errorf("failed to sync anime (AniList -> MAL): %w", err)
		}
	}

	return nil
}

func (a *App) runReverseSync(ctx context.Context) error {
	if a.opts.IsMangaSync() {
		if err := a.reverseSyncManga(ctx); err != nil {
			return fmt.Errorf("failed to sync manga (MAL -> AniList): %w", err)
		}
	}

	if a.opts.IsAnimeSync() {
		if err := a.reverseSyncAnime(ctx); err != nil {
			return fmt.Errorf("failed to sync anime (MAL -> AniList): %w", err)
		}
	}

	return nil
}

// syncAnimeFromAnilistToMAL syncs anime from AniList to MAL
func (a *App) syncAnime(ctx context.Context) error {
	return a.performSync(ctx, MediaTypeAnime, false, a.animeUpdater)
}

// syncMangaFromAnilistToMAL syncs manga from AniList to MAL
func (a *App) syncManga(ctx context.Context) error {
	return a.performSync(ctx, MediaTypeManga, false, a.mangaUpdater)
}

// reverseSyncAnimeFromMALToAnilist syncs anime from MAL to AniList
func (a *App) reverseSyncAnime(ctx context.Context) error {
	return a.performSync(ctx, MediaTypeAnime, true, a.reverseAnimeUpdater)
}

// reverseSyncMangaFromMALToAnilist syncs manga from MAL to AniList
func (a *App) reverseSyncManga(ctx context.Context) error {
	return a.performSync(ctx, MediaTypeManga, true, a.reverseMangaUpdater)
}

// performSync handles syncing for a given media type and direction
func (a *App) performSync(ctx context.Context, mediaType string, reverse bool, updater *Updater) error {
	if updater == nil {
		return fmt.Errorf("updater is nil")
	}
	if updater.Statistics == nil {
		return fmt.Errorf("Statistics is not set for updater: %s", updater.Prefix)
	}
	updater.Statistics.Reset()

	// reverse=true: MAL->AniList (fromAnilist=false)
	// reverse=false: AniList->MAL (fromAnilist=true)
	srcs, tgts, err := a.fetchSyncData(ctx, mediaType, updater.Prefix, !reverse)
	if err != nil {
		return err
	}

	updater.Update(ctx, wrapSources(srcs), wrapTargets(tgts))
	updater.Statistics.Print(updater.Prefix)

	return nil
}

// mediaFetchers contains type-specific functions for fetching and converting media data
type mediaFetchers struct {
	getAnilistList          func(context.Context) (interface{}, error)
	getMALList              func(context.Context) (interface{}, error)
	convertAnilistToSources func(interface{}, bool) []Source
	convertMALToSources     func(interface{}, bool) []Source
	convertAnilistToTargets func(interface{}, bool) []Target
	convertMALToTargets     func(interface{}, bool) []Target
	mediaTypeName           string
}

// fetchSyncData fetches source and target data for the specified sync direction
func (a *App) fetchSyncData(ctx context.Context, mediaType string, prefix string, fromAnilist bool) ([]Source, []Target, error) {
	var srcService, tgtService string

	if fromAnilist {
		srcService = ServiceAnilist
		tgtService = ServiceMAL
		log.Printf("[%s] Fetching %s...", prefix, ServiceAnilist)
	} else {
		srcService = ServiceMAL
		tgtService = ServiceAnilist
		log.Printf("[%s] Fetching %s...", prefix, ServiceMAL)
	}

	var fetchers mediaFetchers
	if mediaType == MediaTypeAnime {
		fetchers = mediaFetchers{
			getAnilistList: func(ctx context.Context) (interface{}, error) {
				return a.anilist.GetUserAnimeList(ctx)
			},
			getMALList: func(ctx context.Context) (interface{}, error) {
				return a.mal.GetUserAnimeList(ctx)
			},
			convertAnilistToSources: func(data interface{}, _ bool) []Source {
				groups, ok := data.([]verniy.MediaListGroup)
				if !ok {
					log.Printf("Warning: unexpected type for AniList anime data: %T", data)
					return nil
				}
				return newSourcesFromAnimes(newAnimesFromMediaListGroups(groups, false))
			},
			convertMALToSources: func(data interface{}, reverse bool) []Source {
				animes, ok := data.([]mal.UserAnime)
				if !ok {
					log.Printf("Warning: unexpected type for MAL anime data: %T", data)
					return nil
				}
				return newSourcesFromAnimes(newAnimesFromMalUserAnimes(animes, reverse))
			},
			convertAnilistToTargets: func(data interface{}, _ bool) []Target {
				groups, ok := data.([]verniy.MediaListGroup)
				if !ok {
					log.Printf("Warning: unexpected type for AniList anime targets: %T", data)
					return nil
				}
				return newTargetsFromAnimes(newAnimesFromMediaListGroups(groups, false))
			},
			convertMALToTargets: func(data interface{}, _ bool) []Target {
				animes, ok := data.([]mal.UserAnime)
				if !ok {
					log.Printf("Warning: unexpected type for MAL anime targets: %T", data)
					return nil
				}
				return newTargetsFromAnimes(newAnimesFromMalUserAnimes(animes, false))
			},
			mediaTypeName: MediaTypeAnime,
		}
	} else {
		fetchers = mediaFetchers{
			getAnilistList: func(ctx context.Context) (interface{}, error) {
				return a.anilist.GetUserMangaList(ctx)
			},
			getMALList: func(ctx context.Context) (interface{}, error) {
				return a.mal.GetUserMangaList(ctx)
			},
			convertAnilistToSources: func(data interface{}, _ bool) []Source {
				groups, ok := data.([]verniy.MediaListGroup)
				if !ok {
					log.Printf("Warning: unexpected type for AniList manga data: %T", data)
					return nil
				}
				return newSourcesFromMangas(newMangasFromMediaListGroups(groups, false))
			},
			convertMALToSources: func(data interface{}, reverse bool) []Source {
				mangas, ok := data.([]mal.UserManga)
				if !ok {
					log.Printf("Warning: unexpected type for MAL manga data: %T", data)
					return nil
				}
				return newSourcesFromMangas(newMangasFromMalUserMangas(mangas, reverse))
			},
			convertAnilistToTargets: func(data interface{}, _ bool) []Target {
				groups, ok := data.([]verniy.MediaListGroup)
				if !ok {
					log.Printf("Warning: unexpected type for AniList manga targets: %T", data)
					return nil
				}
				return newTargetsFromMangas(newMangasFromMediaListGroups(groups, false))
			},
			convertMALToTargets: func(data interface{}, _ bool) []Target {
				mangas, ok := data.([]mal.UserManga)
				if !ok {
					log.Printf("Warning: unexpected type for MAL manga targets: %T", data)
					return nil
				}
				return newTargetsFromMangas(newMangasFromMalUserMangas(mangas, false))
			},
			mediaTypeName: MediaTypeManga,
		}
	}

	return a.fetchMediaSyncData(ctx, prefix, fromAnilist, srcService, tgtService, fetchers)
}

// fetchMediaSyncData is a generic function that handles fetching for both anime and manga
func (a *App) fetchMediaSyncData(ctx context.Context, prefix string, fromAnilist bool, srcService, tgtService string, fetchers mediaFetchers) ([]Source, []Target, error) {
	var srcs []Source
	var tgts []Target

	if fromAnilist {
		srcList, err := fetchers.getAnilistList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("[%s] failed to fetch %s list from AniList: %w", prefix, fetchers.mediaTypeName, err)
		}

		log.Printf("[%s] Fetching %s...", prefix, tgtService)
		tgtList, err := fetchers.getMALList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("[%s] failed to fetch %s list from MAL: %w", prefix, fetchers.mediaTypeName, err)
		}

		srcs = fetchers.convertAnilistToSources(srcList, false)
		tgts = fetchers.convertMALToTargets(tgtList, false)
	} else {
		srcList, err := fetchers.getMALList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("[%s] failed to fetch %s list from MAL: %w", prefix, fetchers.mediaTypeName, err)
		}

		log.Printf("[%s] Fetching %s...", prefix, tgtService)
		tgtList, err := fetchers.getAnilistList(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("[%s] failed to fetch %s list from AniList: %w", prefix, fetchers.mediaTypeName, err)
		}

		srcs = fetchers.convertMALToSources(srcList, a.opts.IsReverseDirection())
		tgts = fetchers.convertAnilistToTargets(tgtList, false)
	}

	log.Printf("[%s] Got %d from %s", prefix, len(srcs), srcService)
	log.Printf("[%s] Got %d from %s", prefix, len(tgts), tgtService)

	return srcs, tgts, nil
}

// buildIgnoreTitlesMap builds an ignore titles map from config and defaults
func buildIgnoreTitlesMap(configMap map[string]struct{}, defaults []string) map[string]struct{} {
	if len(configMap) > 0 {
		return configMap
	}
	if len(defaults) == 0 {
		return configMap
	}
	ignoreMap := make(map[string]struct{}, len(defaults))
	for _, title := range defaults {
		ignoreMap[strings.ToLower(title)] = struct{}{}
	}
	return ignoreMap
}

// Adapters bridge Source/Target interfaces for the sync system

type targetAdapter struct {
	t Target
}

func (ta *targetAdapter) GetTargetID() TargetID {
	if isNil(ta.t) {
		return TargetID(0)
	}
	return ta.t.GetTargetID()
}

func (ta *targetAdapter) GetTitle() string {
	if isNil(ta.t) {
		return ""
	}
	return ta.t.GetTitle()
}

func (ta *targetAdapter) String() string {
	if isNil(ta.t) {
		return "targetAdapter{nil}"
	}
	return ta.t.String()
}

type sourceAdapter struct {
	s Source
}

func (sa *sourceAdapter) GetStatusString() string {
	if isNil(sa.s) {
		return ""
	}
	return sa.s.GetStatusString()
}

func (sa *sourceAdapter) GetTargetID() TargetID {
	if isNil(sa.s) {
		return TargetID(0)
	}
	return sa.s.GetTargetID()
}

func (sa *sourceAdapter) GetTitle() string {
	if isNil(sa.s) {
		return ""
	}
	return sa.s.GetTitle()
}

func (sa *sourceAdapter) GetStringDiffWithTarget(t Target) string {
	if isNil(sa.s) {
		return "Diff{sourceAdapter has nil source}"
	}
	// Try to unwrap targetAdapter using safe helper
	if unwrapped, ok := safeUnwrapTargetAdapter(t); ok {
		return sa.s.GetStringDiffWithTarget(unwrapped)
	}
	// Fallback: pass target directly (will return "Diff{undefined}" if type doesn't match)
	return sa.s.GetStringDiffWithTarget(t)
}

func (sa *sourceAdapter) SameProgressWithTarget(t Target) bool {
	if isNil(sa.s) {
		return false
	}
	if unwrapped, ok := safeUnwrapTargetAdapter(t); ok {
		return sa.s.SameProgressWithTarget(unwrapped)
	}
	return false
}

func (sa *sourceAdapter) SameTypeWithTarget(t Target) bool {
	if isNil(sa.s) {
		return false
	}
	if unwrapped, ok := safeUnwrapTargetAdapter(t); ok {
		return sa.s.SameTypeWithTarget(unwrapped)
	}
	return false
}

func (sa *sourceAdapter) SameTitleWithTarget(t Target) bool {
	if isNil(sa.s) {
		return false
	}
	if unwrapped, ok := safeUnwrapTargetAdapter(t); ok {
		return sa.s.SameTitleWithTarget(unwrapped)
	}
	return false
}

func (sa *sourceAdapter) String() string {
	if isNil(sa.s) {
		return "sourceAdapter{nil}"
	}
	return sa.s.String()
}

func wrapTargets(ts []Target) []Target {
	res := make([]Target, 0, len(ts))
	for _, t := range ts {
		if !isNil(t) {
			res = append(res, &targetAdapter{t: t})
		}
	}
	return res
}

func wrapSources(ss []Source) []Source {
	res := make([]Source, 0, len(ss))
	for _, s := range ss {
		if !isNil(s) {
			res = append(res, &sourceAdapter{s: s})
		}
	}
	return res
}
