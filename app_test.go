package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_Refresh_ResetsPerRunState(t *testing.T) {
	t.Parallel()

	ctx := NewLogger(false).WithContext(t.Context())

	app := &App{
		config: Config{},
		offlineStrategy: &OfflineDatabaseStrategy{
			Database: nil,
		},
		animeUpdater:        newTestUpdater(),
		mangaUpdater:        newTestUpdater(),
		reverseAnimeUpdater: newTestUpdater(),
		reverseMangaUpdater: newTestUpdater(),
		syncReport:          NewSyncReport(),
	}

	// Simulate accumulated state from a previous run
	app.animeUpdater.Statistics.UpdatedCount = 5
	app.animeUpdater.UnmappedList = []UnmappedEntry{{Title: "test"}}
	app.mangaUpdater.Statistics.SkippedCount = 3
	app.reverseAnimeUpdater.Statistics.ErrorCount = 2
	app.reverseMangaUpdater.UnmappedList = []UnmappedEntry{{Title: "test2"}}
	app.syncReport.AddWarning("test", "warning", "detail", "anime")

	app.Refresh(ctx)

	// Statistics should be reset
	assert.Equal(t, 0, app.animeUpdater.Statistics.UpdatedCount)
	assert.Equal(t, 0, app.mangaUpdater.Statistics.SkippedCount)
	assert.Equal(t, 0, app.reverseAnimeUpdater.Statistics.ErrorCount)
	assert.Equal(t, 0, app.reverseMangaUpdater.Statistics.UpdatedCount)

	// UnmappedList should be nil
	assert.Nil(t, app.animeUpdater.UnmappedList)
	assert.Nil(t, app.reverseMangaUpdater.UnmappedList)

	// SyncReport should be fresh (no warnings)
	assert.Empty(t, app.syncReport.Warnings)
}

func newTestUpdater() *Updater {
	return &Updater{
		Statistics:    NewStatistics(),
		StrategyChain: NewStrategyChain(),
	}
}

func TestAppendCacheSaver_SkipsNilKeepsConstructed(t *testing.T) {
	t.Parallel()

	var nilHato *HatoClient
	hato := &HatoClient{}

	clients := appendCacheSaver(nil, nilHato)
	assert.Empty(t, clients, "a disabled source's nil client must not be collected")

	clients = appendCacheSaver(clients, hato)
	assert.Equal(t, []cacheSaver{hato}, clients, "an enabled source's client must be collected")
}

func TestLoadIDMappingStrategies_DisabledSourcesAreNoOps(t *testing.T) {
	t.Parallel()

	ctx := NewLogger(false).WithContext(t.Context())

	strategies := loadIDMappingStrategies(ctx, Config{}, true)

	assert.Nil(t, strategies.offline.Database)
	assert.Nil(t, strategies.hato.Client)
	assert.Nil(t, strategies.arm.Client)
	assert.Nil(t, strategies.mangaBaka.Client)
	assert.Nil(t, strategies.jikan.Client)
	assert.Nil(t, strategies.hatoClient)
	assert.Nil(t, strategies.mangaBakaClient)
	assert.Nil(t, strategies.jikanClient)
	assert.Empty(t, strategies.cacheClients)
}

// strategyNames extracts strategy names in order for chain-composition assertions.
func strategyNames(chain *StrategyChain) []string {
	names := make([]string, 0, len(chain.strategies))
	for _, s := range chain.strategies {
		names = append(names, s.Name())
	}
	return names
}

func TestBuildForwardMangaChain_MangaBakaAfterHato(t *testing.T) {
	t.Parallel()

	idStrategies := idMappingStrategies{
		hato:      HatoAPIStrategy{},
		mangaBaka: MangaBakaStrategy{},
		jikan:     JikanAPIStrategy{},
	}
	chain := buildForwardMangaChain(ManualMappingStrategy{}, idStrategies, nil)

	names := strategyNames(chain)
	hatoIdx := strategyIndexOf(names, "HatoAPIStrategy")
	mangaBakaIdx := strategyIndexOf(names, "MangaBakaStrategy")
	assert.NotEqual(t, -1, hatoIdx)
	assert.Equal(t, hatoIdx+1, mangaBakaIdx)
}

func TestBuildReverseMangaChain_MangaBakaAfterHato(t *testing.T) {
	t.Parallel()

	idStrategies := idMappingStrategies{
		hato:      HatoAPIStrategy{},
		mangaBaka: MangaBakaStrategy{},
		jikan:     JikanAPIStrategy{},
	}
	chain := buildReverseMangaChain(ManualMappingStrategy{}, idStrategies, nil)

	names := strategyNames(chain)
	hatoIdx := strategyIndexOf(names, "HatoAPIStrategy")
	mangaBakaIdx := strategyIndexOf(names, "MangaBakaStrategy")
	assert.NotEqual(t, -1, hatoIdx)
	assert.Equal(t, hatoIdx+1, mangaBakaIdx)
}

// MangaBaka maps manga only, so an anime chain must never reach for it.
func TestAnimeChains_DoNotIncludeMangaBaka(t *testing.T) {
	t.Parallel()

	idStrategies := idMappingStrategies{
		offline:   &OfflineDatabaseStrategy{},
		hato:      HatoAPIStrategy{},
		arm:       ARMAPIStrategy{},
		mangaBaka: MangaBakaStrategy{},
		jikan:     JikanAPIStrategy{},
	}

	forward := strategyNames(buildForwardAnimeChain(ManualMappingStrategy{}, idStrategies, nil))
	reverse := strategyNames(buildReverseAnimeChain(ManualMappingStrategy{}, idStrategies, nil))

	assert.Equal(t, -1, strategyIndexOf(forward, "MangaBakaStrategy"))
	assert.Equal(t, -1, strategyIndexOf(reverse, "MangaBakaStrategy"))
	// A chain that lost Hato would pass the assertion above vacuously.
	assert.NotEqual(t, -1, strategyIndexOf(forward, "HatoAPIStrategy"))
	assert.NotEqual(t, -1, strategyIndexOf(reverse, "HatoAPIStrategy"))
}

func strategyIndexOf(names []string, name string) int {
	for i, n := range names {
		if n == name {
			return i
		}
	}
	return -1
}

// fakeCacheSaver counts SaveCache calls and can be made to fail.
type fakeCacheSaver struct {
	calls int
	err   error
}

func (f *fakeCacheSaver) SaveCache(context.Context) error {
	f.calls++
	return f.err
}

func TestApp_SaveCaches_SkipsNilClientsAndCallsEnabledOnes(t *testing.T) {
	t.Parallel()

	ctx := NewLogger(false).WithContext(t.Context())

	// No cache-owning clients at all - must not panic or error.
	app := &App{cacheClients: nil}
	app.saveCaches(ctx)

	saver := &fakeCacheSaver{}
	app = &App{cacheClients: []cacheSaver{saver}}
	app.saveCaches(ctx)
	assert.Equal(t, 1, saver.calls)
}

func TestApp_SaveCaches_LogsAndContinuesOnError(t *testing.T) {
	t.Parallel()

	ctx := NewLogger(false).WithContext(t.Context())

	failing := &fakeCacheSaver{err: assert.AnError}
	succeeding := &fakeCacheSaver{}
	app := &App{cacheClients: []cacheSaver{failing, succeeding}}

	app.saveCaches(ctx)

	assert.Equal(t, 1, failing.calls)
	assert.Equal(t, 1, succeeding.calls)
}
