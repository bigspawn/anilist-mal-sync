package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncReport(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	if report.Warnings == nil {
		t.Error("Warnings slice should be initialized")
	}
	if len(report.Warnings) != 0 {
		t.Error("Warnings should be empty initially")
	}
}

func TestSyncReport_AddWarning(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		title     string
		reason    string
		detail    string
		mediaType string
	}{
		{
			name:      "add warning with all fields",
			title:     "Test Anime",
			reason:    "episode count mismatch",
			detail:    "(1 vs 12)",
			mediaType: "Anime",
		},
		{
			name:      "add warning with minimal fields",
			title:     "Test Manga",
			reason:    "different MAL IDs",
			detail:    "",
			mediaType: "Manga",
		},
		{
			name:      "add warning with empty detail",
			title:     "Test Anime 2",
			reason:    "title mismatch",
			detail:    "",
			mediaType: "Anime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewSyncReport()
			report.AddWarning(tt.title, tt.reason, tt.detail, tt.mediaType)

			assert.Equal(t, 1, len(report.Warnings), "Should have exactly one warning")

			warning := report.Warnings[0]
			assert.Equal(t, tt.title, warning.Title, "Title should match")
			assert.Equal(t, tt.reason, warning.Reason, "Reason should match")
			assert.Equal(t, tt.detail, warning.Detail, "Detail should match")
			assert.Equal(t, tt.mediaType, warning.MediaType, "MediaType should match")
		})
	}
}

func TestSyncReport_AddMultipleWarnings(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	report.AddWarning("Anime 1", "reason 1", "detail 1", "Anime")
	report.AddWarning("Anime 2", "reason 2", "detail 2", "Anime")
	report.AddWarning("Manga 1", "reason 3", "detail 3", "Manga")

	assert.Equal(t, 3, len(report.Warnings), "Should have three warnings")

	assert.Equal(t, "Anime 1", report.Warnings[0].Title)
	assert.Equal(t, "Anime 2", report.Warnings[1].Title)
	assert.Equal(t, "Manga 1", report.Warnings[2].Title)
}

func TestSyncReport_HasWarnings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(*SyncReport)
		expected bool
	}{
		{
			name:     "empty report has no warnings",
			setup:    func(_ *SyncReport) {},
			expected: false,
		},
		{
			name: "report with one warning",
			setup: func(r *SyncReport) {
				r.AddWarning("Test", "reason", "detail", "Anime")
			},
			expected: true,
		},
		{
			name: "report with multiple warnings",
			setup: func(r *SyncReport) {
				r.AddWarning("Test 1", "reason 1", "detail 1", "Anime")
				r.AddWarning("Test 2", "reason 2", "detail 2", "Manga")
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewSyncReport()
			tt.setup(report)

			got := report.HasWarnings()
			assert.Equal(t, tt.expected, got, "HasWarnings should return expected value")
		})
	}
}

func TestSyncReport_WarningsPreserveOrder(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	// Add warnings in specific order
	report.AddWarning("Third", "reason 3", "detail 3", "Anime")
	report.AddWarning("First", "reason 1", "detail 1", "Anime")
	report.AddWarning("Second", "reason 2", "detail 2", "Manga")

	assert.Equal(t, 3, len(report.Warnings))
	assert.Equal(t, "Third", report.Warnings[0].Title)
	assert.Equal(t, "First", report.Warnings[1].Title)
	assert.Equal(t, "Second", report.Warnings[2].Title)
}

func TestSyncReport_WarningStruct(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()
	report.AddWarning("Test Title", "Test Reason", "Test Detail", "Test Media")

	warning := report.Warnings[0]

	assert.Equal(t, "Test Title", warning.Title)
	assert.Equal(t, "Test Reason", warning.Reason)
	assert.Equal(t, "Test Detail", warning.Detail)
	assert.Equal(t, "Test Media", warning.MediaType)
}

// =============================================================================
// AddUnmappedItems / HasUnmappedItems
// =============================================================================

func TestSyncReport_AddUnmappedItems(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	items := []UnmappedEntry{
		{Title: "Anime A", AniListID: 1, MediaType: "anime"},
		{Title: "Manga B", MALID: 2, MediaType: "manga"},
	}
	report.AddUnmappedItems(items)

	assert.Len(t, report.UnmappedItems, 2)
	assert.Equal(t, "Anime A", report.UnmappedItems[0].Title)
	assert.Equal(t, "Manga B", report.UnmappedItems[1].Title)
}

func TestSyncReport_AddUnmappedItems_Empty(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()
	report.AddUnmappedItems(nil)
	report.AddUnmappedItems([]UnmappedEntry{})

	assert.Empty(t, report.UnmappedItems)
}

func TestSyncReport_AddUnmappedItems_Accumulates(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()
	report.AddUnmappedItems([]UnmappedEntry{{Title: "A"}})
	report.AddUnmappedItems([]UnmappedEntry{{Title: "B"}, {Title: "C"}})

	assert.Len(t, report.UnmappedItems, 3)
}

func TestSyncReport_HasUnmappedItems(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(*SyncReport)
		expected bool
	}{
		{"empty report", func(_ *SyncReport) {}, false},
		{"one item", func(r *SyncReport) {
			r.AddUnmappedItems([]UnmappedEntry{{Title: "X"}})
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSyncReport()
			tt.setup(r)
			assert.Equal(t, tt.expected, r.HasUnmappedItems())
		})
	}
}

// =============================================================================
// AddDuplicateConflict / HasDuplicateConflicts
// =============================================================================

func TestSyncReport_AddDuplicateConflict(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	report.AddDuplicateConflict("Loser", "Winner", "Target", "stratA", "stratB", "anime")

	assert.Len(t, report.DuplicateConflicts, 1)
	c := report.DuplicateConflicts[0]
	assert.Equal(t, "Loser", c.LoserTitle)
	assert.Equal(t, "Winner", c.WinnerTitle)
	assert.Equal(t, "Target", c.TargetTitle)
	assert.Equal(t, "stratA", c.LoserStrat)
	assert.Equal(t, "stratB", c.WinnerStrat)
	assert.Equal(t, "anime", c.MediaType)
}

func TestSyncReport_AddDuplicateConflict_Multiple(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	report.AddDuplicateConflict("L1", "W1", "T1", "s1", "s2", "anime")
	report.AddDuplicateConflict("L2", "W2", "T2", "s3", "s4", "manga")

	assert.Len(t, report.DuplicateConflicts, 2)
	assert.Equal(t, "L2", report.DuplicateConflicts[1].LoserTitle)
}

func TestSyncReport_HasDuplicateConflicts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(*SyncReport)
		expected bool
	}{
		{"empty", func(_ *SyncReport) {}, false},
		{"one conflict", func(r *SyncReport) {
			r.AddDuplicateConflict("L", "W", "T", "s1", "s2", "anime")
		}, true},
		{"two conflicts", func(r *SyncReport) {
			r.AddDuplicateConflict("L1", "W1", "T1", "s1", "s2", "anime")
			r.AddDuplicateConflict("L2", "W2", "T2", "s3", "s4", "manga")
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSyncReport()
			tt.setup(r)
			assert.Equal(t, tt.expected, r.HasDuplicateConflicts())
		})
	}
}

// =============================================================================
// AddFavoritesResult / HasFavoritesMismatches
// =============================================================================

func TestSyncReport_AddFavoritesResult(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	result := FavoritesResult{
		Added: 3,
		Mismatches: []FavoriteMismatch{
			{Title: "Anime X", AniListID: 1, MALID: 10, MediaType: "anime", OnAniList: true, OnMAL: false},
		},
	}
	report.AddFavoritesResult(result)

	assert.Equal(t, 3, report.FavoritesAdded)
	assert.Len(t, report.FavoritesMismatches, 1)
	assert.Equal(t, "Anime X", report.FavoritesMismatches[0].Title)
}

func TestSyncReport_AddFavoritesResult_Accumulates(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	report.AddFavoritesResult(FavoritesResult{Added: 2, Mismatches: []FavoriteMismatch{{Title: "A"}}})
	report.AddFavoritesResult(FavoritesResult{Added: 5, Mismatches: []FavoriteMismatch{{Title: "B"}, {Title: "C"}}})

	assert.Equal(t, 7, report.FavoritesAdded)
	assert.Len(t, report.FavoritesMismatches, 3)
}

func TestSyncReport_HasFavoritesMismatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(*SyncReport)
		expected bool
	}{
		{"empty", func(_ *SyncReport) {}, false},
		{"only added, no mismatches", func(r *SyncReport) {
			r.AddFavoritesResult(FavoritesResult{Added: 5})
		}, false},
		{"with mismatches", func(r *SyncReport) {
			r.AddFavoritesResult(FavoritesResult{
				Mismatches: []FavoriteMismatch{{Title: "X"}},
			})
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSyncReport()
			tt.setup(r)
			assert.Equal(t, tt.expected, r.HasFavoritesMismatches())
		})
	}
}

// =============================================================================
// Thread-safety: SyncReport mutex protection
// =============================================================================

func TestSyncReport_ConcurrentAccess(t *testing.T) {
	// Run with -race to verify mutex protection on all Add/Has methods.
	report := NewSyncReport()

	const goroutines = 20
	done := make(chan struct{})

	for range goroutines {
		go func() {
			report.AddWarning("T", "r", "d", "anime")
			report.AddDuplicateConflict("L", "W", "T", "s1", "s2", "anime")
			report.AddUnmappedItems([]UnmappedEntry{{Title: "X"}})
			report.AddFavoritesResult(FavoritesResult{Added: 1})
			_ = report.HasWarnings()
			_ = report.HasDuplicateConflicts()
			_ = report.HasUnmappedItems()
			_ = report.HasFavoritesMismatches()
			done <- struct{}{}
		}()
	}

	for range goroutines {
		<-done
	}

	assert.Equal(t, goroutines, len(report.Warnings))
	assert.Equal(t, goroutines, report.FavoritesAdded)
}
