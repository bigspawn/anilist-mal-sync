package main

import (
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
