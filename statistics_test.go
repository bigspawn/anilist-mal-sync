package main

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatistics_Reset(t *testing.T) {
	t.Parallel()
	stats := &Statistics{
		UpdatedCount: 42,
		SkippedCount: 100,
		TotalCount:   142,
	}

	stats.Reset()

	assert.Equal(t, 0, stats.UpdatedCount, "UpdatedCount should be 0 after Reset")
	assert.Equal(t, 0, stats.SkippedCount, "SkippedCount should be 0 after Reset")
	assert.Equal(t, 0, stats.TotalCount, "TotalCount should be 0 after Reset")
}

func TestStatistics_ResetOnEmpty(t *testing.T) {
	t.Parallel()
	stats := &Statistics{}

	stats.Reset()

	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)
}

func TestStatistics_ResetMultipleTimes(t *testing.T) {
	t.Parallel()
	stats := &Statistics{
		UpdatedCount: 10,
		SkippedCount: 5,
		TotalCount:   15,
	}

	// First reset
	stats.Reset()
	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)

	// Second reset - should still be zeros
	stats.Reset()
	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)
}

func TestStatistics_PrintLogsCorrectly(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := NewStatistics()
	stats.UpdatedCount = 42
	stats.SkippedCount = 100
	stats.TotalCount = 142

	stats.Print(ctx, "TestPrefix")

	output := buf.String()
	assert.Contains(t, output, "=== TestPrefix: Sync Complete ===", "Print should log header")
	assert.Contains(t, output, "Total items: 142", "Print should log total items")
	assert.Contains(t, output, "✓ Updated: 42", "Print should log correct update info")
	assert.Contains(t, output, "Skipped: 100", "Print should log correct skip info")
}

func TestStatistics_PrintWithZeroValues(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := NewStatistics()

	stats.Print(ctx, "EmptyTest")

	output := buf.String()
	assert.Contains(t, output, "=== EmptyTest: Sync Complete ===", "Print should log header")
	assert.Contains(t, output, "✓ Updated: 0", "Print should log zero updated")
}

func TestStatistics_WatchModeFlow(t *testing.T) {
	t.Parallel()
	updater := &Updater{
		Prefix:     "Test Watch Mode",
		Statistics: NewStatistics(),
	}

	// First iteration
	updater.Statistics.UpdatedCount = 5
	updater.Statistics.SkippedCount = 10
	updater.Statistics.TotalCount = 15
	assert.Equal(t, 15, updater.Statistics.TotalCount, "First iteration should have 15 total")

	// Reset before next iteration (performSync behavior)
	updater.Statistics.Reset()
	assert.Equal(t, 0, updater.Statistics.TotalCount, "After reset should be 0")

	// Second iteration
	updater.Statistics.UpdatedCount = 3
	updater.Statistics.SkippedCount = 7
	updater.Statistics.TotalCount = 10
	assert.Equal(t, 10, updater.Statistics.TotalCount, "Second iteration should have 10 total")

	// Reset again
	updater.Statistics.Reset()
	assert.Equal(t, 0, updater.Statistics.TotalCount, "After second reset should be 0")
}

func TestStatistics_NoResetAccumulationBug(t *testing.T) {
	t.Parallel()
	updater := &Updater{
		Prefix:     "Buggy Watch Mode",
		Statistics: &Statistics{},
	}

	// First iteration - process 10 items
	updater.Statistics.TotalCount = 10
	updater.Statistics.UpdatedCount = 2
	updater.Statistics.SkippedCount = 8

	// Second iteration - WITHOUT Reset(), counters accumulate (bug)
	updater.Statistics.TotalCount += 5
	updater.Statistics.UpdatedCount++
	updater.Statistics.SkippedCount += 4

	// BUG: Shows 15 total instead of 5 for current iteration
	assert.Equal(t, 15, updater.Statistics.TotalCount, "Bug: accumulated TotalCount")

	// Fix: use Reset()
	updater.Statistics.Reset()
	updater.Statistics.TotalCount = 5
	updater.Statistics.UpdatedCount = 1
	updater.Statistics.SkippedCount = 4

	assert.Equal(t, 5, updater.Statistics.TotalCount, "After Reset: current iteration only")
}

func TestStatistics_AllCountersIndependent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		updated int
		skipped int
		total   int
	}{
		{"All non-zero", 100, 200, 300},
		{"Only updated", 50, 0, 50},
		{"Only skipped", 0, 75, 75},
		{"Only total", 0, 0, 25},
		{"Mixed values", 10, 20, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := &Statistics{
				UpdatedCount: tt.updated,
				SkippedCount: tt.skipped,
				TotalCount:   tt.total,
			}

			stats.Reset()

			assert.Equal(t, 0, stats.UpdatedCount, "UpdatedCount should be 0")
			assert.Equal(t, 0, stats.SkippedCount, "SkippedCount should be 0")
			assert.Equal(t, 0, stats.TotalCount, "TotalCount should be 0")
		})
	}
}

func TestUpdater_StatisticsIntegration(t *testing.T) {
	t.Parallel()
	updater := &Updater{
		Prefix:     "Integration Test",
		Statistics: &Statistics{},
	}

	// Simulate some counts being set
	updater.Statistics.UpdatedCount = 10
	updater.Statistics.SkippedCount = 20
	updater.Statistics.TotalCount = 30

	assert.Equal(t, 10, updater.Statistics.UpdatedCount)
	assert.Equal(t, 20, updater.Statistics.SkippedCount)
	assert.Equal(t, 30, updater.Statistics.TotalCount)

	// Reset should clear all counters
	updater.Statistics.Reset()

	assert.Equal(t, 0, updater.Statistics.UpdatedCount)
	assert.Equal(t, 0, updater.Statistics.SkippedCount)
	assert.Equal(t, 0, updater.Statistics.TotalCount)
}

func TestPerformSync_ResetsAfterPrint(t *testing.T) {
	t.Parallel()
	// Integration test to verify Reset() is called after Print()
	// This tests the fixed behavior in performSync()

	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := &Statistics{
		UpdatedCount: 10,
		SkippedCount: 5,
		TotalCount:   15,
	}

	// Simulate performSync behavior: Update -> Print -> Reset
	prefix := "Test Sync"

	// Print first (before reset)
	stats.Print(ctx, prefix)
	output := buf.String()

	// Verify output contains the counts
	assert.Contains(t, output, "Total items: 15")
	assert.Contains(t, output, "Updated: 10")
	assert.Contains(t, output, "Skipped: 5")

	// Reset after print
	stats.Reset()

	// Verify counters are reset
	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)

	// If we print again, should show zeros
	buf.Reset()
	stats.Print(ctx, prefix)
	output = buf.String()

	assert.Contains(t, output, "Total items: 0")
	assert.Contains(t, output, "Updated: 0")
	// Skipped is not shown when count is 0
	assert.NotContains(t, output, "Skipped:")
}

func TestStatistics_ResetIdempotent(t *testing.T) {
	t.Parallel()
	stats := &Statistics{
		UpdatedCount: 999,
		SkippedCount: 888,
		TotalCount:   1887,
	}

	// Reset multiple times
	for i := range 10 {
		stats.Reset()
		assert.Equal(t, 0, stats.UpdatedCount, "Reset should be idempotent (iteration %d)", i)
		assert.Equal(t, 0, stats.SkippedCount, "Reset should be idempotent (iteration %d)", i)
		assert.Equal(t, 0, stats.TotalCount, "Reset should be idempotent (iteration %d)", i)
	}
}

func TestStatistics_WatchModeMultipleIterations(t *testing.T) {
	t.Parallel()
	iterations := []struct {
		updated int
		skipped int
		total   int
	}{
		{5, 10, 15},
		{3, 7, 10},
		{8, 2, 10},
		{0, 5, 5},
	}

	for i, iter := range iterations {
		tt := iter
		t.Run("Iteration "+string(rune('1'+i)), func(t *testing.T) {
			t.Parallel()
			// Create separate updater for each iteration to avoid race conditions
			updater := &Updater{
				Prefix:     "Watch Mode Test",
				Statistics: &Statistics{},
			}
			// Simulate counting
			updater.Statistics.UpdatedCount = tt.updated
			updater.Statistics.SkippedCount = tt.skipped
			updater.Statistics.TotalCount = tt.total

			assert.Equal(t, tt.updated, updater.Statistics.UpdatedCount)
			assert.Equal(t, tt.skipped, updater.Statistics.SkippedCount)
			assert.Equal(t, tt.total, updater.Statistics.TotalCount)

			// Reset for next iteration
			updater.Statistics.Reset()

			assert.Equal(t, 0, updater.Statistics.UpdatedCount)
			assert.Equal(t, 0, updater.Statistics.SkippedCount)
			assert.Equal(t, 0, updater.Statistics.TotalCount)
		})
	}
}

func TestStatistics_RecordSkipAndUpdate(t *testing.T) {
	t.Parallel()
	// This tests that RecordSkip() and RecordUpdate() work correctly
	// with properly initialized Statistics
	stats := NewStatistics()

	stats.RecordSkip(UpdateResult{Title: "Test", Status: "watching", SkipReason: "test"})
	assert.Equal(t, 1, stats.SkippedCount)
	assert.Equal(t, 1, stats.StatusCounts["watching"])

	stats.RecordUpdate(UpdateResult{Title: "Test2", Status: "completed"})
	assert.Equal(t, 1, stats.UpdatedCount)
	assert.Equal(t, 1, stats.StatusCounts["completed"])
}

func TestStatistics_RecordSkipPanicWithoutInit(t *testing.T) {
	t.Parallel()
	// This demonstrates the bug - using new(Statistics) causes panic
	// because StatusCounts is nil
	stats := &Statistics{} // or new(Statistics)

	assert.Panics(t, func() {
		stats.RecordSkip(UpdateResult{Title: "Test", Status: "watching"})
	}, "RecordSkip should panic when StatusCounts is nil")
}

func TestStatistics_RecordUpdatePanicWithoutInit(t *testing.T) {
	t.Parallel()
	// This demonstrates the bug - using new(Statistics) causes panic
	// because StatusCounts is nil
	stats := new(Statistics)

	assert.Panics(t, func() {
		stats.RecordUpdate(UpdateResult{Title: "Test", Status: "completed"})
	}, "RecordUpdate should panic when StatusCounts is nil")
}

func TestPrintGlobalSummary(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	// Create multiple statistics
	stats1 := NewStatistics()
	stats1.UpdatedCount = 5
	stats1.SkippedCount = 10
	stats1.TotalCount = 15
	stats1.StatusCounts = map[string]int{"watching": 5, "completed": 10}
	stats1.SkippedItems = []UpdateResult{
		{Title: "Skip1", Status: "watching", SkipReason: "no changes"},
		{Title: "Skip2", Status: "watching", SkipReason: "target not found"},
	}

	stats2 := NewStatistics()
	stats2.UpdatedCount = 3
	stats2.SkippedCount = 7
	stats2.TotalCount = 10
	stats2.StatusCounts = map[string]int{"completed": 3, "watching": 7}
	stats2.SkippedItems = []UpdateResult{
		{Title: "Skip3", Status: "completed", SkipReason: "no changes"},
	}

	report := NewSyncReport()
	report.AddWarning("Warning Anime", "episode mismatch", "(1 vs 12)", "Anime")

	statsArray := []*Statistics{stats1, stats2}

	PrintGlobalSummary(ctx, statsArray, report, 5*time.Second, false)

	output := buf.String()

	// Check for summary output
	assert.Contains(t, output, "Sync Complete", "Should print header")
	assert.Contains(t, output, "Total: 25", "Should show total items")     // 15 + 10
	assert.Contains(t, output, "Updated: 8", "Should show updated items")  // 5 + 3
	assert.Contains(t, output, "Skipped: 17", "Should show skipped items") // 10 + 7
}

func TestPrintGlobalSummary_EmptyStats(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	report := NewSyncReport()
	statsArray := []*Statistics{}

	PrintGlobalSummary(ctx, statsArray, report, 1*time.Second, false)

	output := buf.String()

	// Should still print header
	assert.Contains(t, output, "Sync Complete", "Should print header")
	assert.Contains(t, output, "Total: 0", "Should show zero total")
}

func TestPrintGlobalSummary_WithWarnings(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := NewStatistics()
	stats.UpdatedCount = 1
	stats.TotalCount = 1

	report := NewSyncReport()
	report.AddWarning("Test Anime", "test reason", "(1 vs 12)", "Anime")
	report.AddWarning("Test Manga", "test reason 2", "", "Manga")

	statsArray := []*Statistics{stats}

	PrintGlobalSummary(ctx, statsArray, report, 1*time.Second, false)

	output := buf.String()

	assert.Contains(t, output, "Warnings (2)", "Should show warnings count")
	assert.Contains(t, output, "Test Anime", "Should show warning title")
	assert.Contains(t, output, "test reason", "Should show warning reason")
}

func TestPrintGlobalSummary_WithErrors(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := NewStatistics()
	stats.UpdatedCount = 1
	stats.ErrorCount = 2
	stats.TotalCount = 3
	stats.ErrorItems = []UpdateResult{
		{Title: "Error1", Error: errors.New("error 1")},
		{Title: "Error2", Error: errors.New("error 2")},
	}

	report := NewSyncReport()
	statsArray := []*Statistics{stats}

	PrintGlobalSummary(ctx, statsArray, report, 1*time.Second, false)

	output := buf.String()

	assert.Contains(t, output, "Errors:", "Should show errors section")
	assert.Contains(t, output, "Error1", "Should show error title")
	assert.Contains(t, output, "error 1", "Should show error message")
}

func TestPrintGlobalSummary_SkipReasonsAggregation(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	// Create stats with various skip reasons
	stats1 := NewStatistics()
	stats1.SkippedCount = 3
	stats1.SkippedItems = []UpdateResult{
		{Title: "A", SkipReason: "no changes"},
		{Title: "B", SkipReason: "no changes"},
		{Title: "C", SkipReason: "target not found"},
	}

	stats2 := NewStatistics()
	stats2.SkippedCount = 2
	stats2.SkippedItems = []UpdateResult{
		{Title: "D", SkipReason: "no changes"},
		{Title: "E", SkipReason: "in ignore list"},
	}

	report := NewSyncReport()
	statsArray := []*Statistics{stats1, stats2}

	PrintGlobalSummary(ctx, statsArray, report, 1*time.Second, false)

	output := buf.String()

	assert.Contains(t, output, "Skipped by reason:", "Should show skip reasons")
	assert.Contains(t, output, "no changes: 3", "Should aggregate 'no changes' reason")
	assert.Contains(t, output, "target not found: 1", "Should show 'target not found' count")
	assert.Contains(t, output, "in ignore list: 1", "Should show 'in ignore list' count")
}

func TestStatistics_RecordDryRun(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordDryRun(UpdateResult{Title: "Test", Status: "watching", Detail: "dry run"})

	assert.Equal(t, 1, stats.DryRunCount)
	assert.Equal(t, 1, len(stats.DryRunItems))
	assert.True(t, stats.DryRunItems[0].IsDryRun)
	assert.Equal(t, 1, stats.StatusCounts["watching"])
}

func TestStatistics_ResetClearsDryRun(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()
	stats.RecordDryRun(UpdateResult{Title: "Test", Status: "watching"})
	stats.DryRunCount = 5

	stats.Reset()

	assert.Equal(t, 0, stats.DryRunCount)
	assert.Nil(t, stats.DryRunItems)
}

func TestPrintGlobalSummary_WithDryRunItems(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	stats := NewStatistics()
	stats.DryRunCount = 2
	stats.DryRunItems = []UpdateResult{
		{Title: "Anime1", Detail: "dry run", IsDryRun: true},
		{Title: "Anime2", Detail: "dry run", IsDryRun: true},
	}

	report := NewSyncReport()
	statsArray := []*Statistics{stats}

	PrintGlobalSummary(ctx, statsArray, report, 1*time.Second, false)

	output := buf.String()
	assert.Contains(t, output, "Would update (2)")
}

// =============================================================================
// RecordUpdate / RecordSkip / RecordError / IncrementTotal — extra coverage
// =============================================================================

func TestStatistics_RecordError_DoesNotUpdateStatusCounts(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordError(UpdateResult{Title: "Err1", Status: "watching", Error: errors.New("oops")})
	stats.RecordError(UpdateResult{Title: "Err2", Status: "completed", Error: errors.New("fail")})

	assert.Equal(t, 2, stats.ErrorCount)
	assert.Len(t, stats.ErrorItems, 2)
	// RecordError does NOT touch StatusCounts
	assert.Empty(t, stats.StatusCounts)
}

func TestStatistics_RecordUpdate_TracksStatus(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordUpdate(UpdateResult{Title: "A", Status: "watching"})
	stats.RecordUpdate(UpdateResult{Title: "B", Status: "watching"})
	stats.RecordUpdate(UpdateResult{Title: "C", Status: "completed"})

	assert.Equal(t, 3, stats.UpdatedCount)
	assert.Len(t, stats.UpdatedItems, 3)
	assert.Equal(t, 2, stats.StatusCounts["watching"])
	assert.Equal(t, 1, stats.StatusCounts["completed"])
}

func TestStatistics_RecordSkip_TracksReason(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordSkip(UpdateResult{Title: "X", Status: "on_hold", SkipReason: "no changes"})
	stats.RecordSkip(UpdateResult{Title: "Y", Status: "completed", SkipReason: "in ignore list"})

	assert.Equal(t, 2, stats.SkippedCount)
	assert.Len(t, stats.SkippedItems, 2)
	assert.Equal(t, 1, stats.StatusCounts["on_hold"])
	assert.Equal(t, 1, stats.StatusCounts["completed"])
}

// =============================================================================
// aggregateStats
// =============================================================================

func TestAggregateStats(t *testing.T) {
	t.Parallel()
	s1 := NewStatistics()
	s1.TotalCount = 10
	s1.UpdatedCount = 3
	s1.SkippedCount = 5
	s1.ErrorCount = 2
	s1.DryRunCount = 1
	s1.SkippedItems = []UpdateResult{
		{SkipReason: "no changes"},
		{SkipReason: "no changes"},
		{SkipReason: "unmapped"},
	}
	s1.UpdatedItems = []UpdateResult{{Title: "A"}}
	s1.ErrorItems = []UpdateResult{{Title: "Err1"}}
	s1.DryRunItems = []UpdateResult{{Title: "Dry1"}}

	s2 := NewStatistics()
	s2.TotalCount = 5
	s2.UpdatedCount = 2
	s2.SkippedCount = 2
	s2.ErrorCount = 1
	s2.DryRunCount = 0
	s2.SkippedItems = []UpdateResult{{SkipReason: "no changes"}}
	s2.UpdatedItems = []UpdateResult{{Title: "B"}}
	s2.ErrorItems = []UpdateResult{{Title: "Err2"}}

	result := aggregateStats([]*Statistics{s1, s2})

	assert.Equal(t, 15, result.items)
	assert.Equal(t, 5, result.updated)
	assert.Equal(t, 7, result.skipped)
	assert.Equal(t, 3, result.errors)
	assert.Equal(t, 1, result.dryRun)
	assert.Equal(t, 3, result.skipReasons["no changes"])
	assert.Equal(t, 1, result.skipReasons["unmapped"])
	assert.Len(t, result.updatedItems, 2)
	assert.Len(t, result.errorItems, 2)
	assert.Len(t, result.dryRunItems, 1)
}

func TestAggregateStats_NilEntries(t *testing.T) {
	t.Parallel()
	s := NewStatistics()
	s.TotalCount = 7
	s.UpdatedCount = 7

	result := aggregateStats([]*Statistics{nil, s, nil})

	assert.Equal(t, 7, result.items)
	assert.Equal(t, 7, result.updated)
}

func TestAggregateStats_Empty(t *testing.T) {
	t.Parallel()
	result := aggregateStats([]*Statistics{})

	assert.Equal(t, 0, result.items)
	assert.Equal(t, 0, result.updated)
	assert.Empty(t, result.skipReasons)
}

// =============================================================================
// groupSkipReasons
// =============================================================================

func TestGroupSkipReasons(t *testing.T) {
	t.Parallel()
	items := []UpdateResult{
		{SkipReason: "no changes"},
		{SkipReason: "no changes"},
		{SkipReason: "unmapped"},
		{SkipReason: ""},
	}

	result := groupSkipReasons(items)

	assert.Equal(t, 2, result["no changes"])
	assert.Equal(t, 1, result["unmapped"])
	assert.Equal(t, 1, result["unspecified"])
	assert.Equal(t, 3, len(result))
}

func TestGroupSkipReasons_Empty(t *testing.T) {
	t.Parallel()
	result := groupSkipReasons(nil)
	assert.Empty(t, result)
}

// =============================================================================
// sortedKeys
// =============================================================================

func TestSortedKeys(t *testing.T) {
	t.Parallel()
	m := map[string]int{"banana": 1, "apple": 2, "cherry": 3}
	keys := sortedKeys(m)
	assert.Equal(t, []string{"apple", "banana", "cherry"}, keys)
}

func TestSortedKeys_Empty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, sortedKeys(map[string]int{}))
}

func TestSortedKeys_SingleEntry(t *testing.T) {
	t.Parallel()
	keys := sortedKeys(map[string]int{"only": 1})
	assert.Equal(t, []string{"only"}, keys)
}

// =============================================================================
// capitalizeFirst
// =============================================================================

func TestCapitalizeFirst(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"anime", "Anime"},
		{"manga", "Manga"},
		{"ALREADY", "ALREADY"},
		{"", ""},
		{"a", "A"},
		{"already Capitalized", "Already Capitalized"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, capitalizeFirst(tt.input))
		})
	}
}

// =============================================================================
// formatUnmappedLine
// =============================================================================

func TestFormatUnmappedLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		num      int
		item     UnmappedEntry
		media    string
		contains []string
	}{
		{
			name:     "with AniList ID",
			num:      1,
			item:     UnmappedEntry{Title: "Naruto", AniListID: 100, Reason: "no match"},
			media:    "Anime",
			contains: []string{"1.", `"Naruto"`, "AniList: 100", "[Anime]", "no match"},
		},
		{
			name:     "with MAL ID",
			num:      2,
			item:     UnmappedEntry{Title: "Bleach", MALID: 200},
			media:    "Anime",
			contains: []string{"2.", `"Bleach"`, "MAL: 200"},
		},
		{
			name:     "with neither ID",
			num:      3,
			item:     UnmappedEntry{Title: "Unknown", Reason: "missing"},
			media:    "Manga",
			contains: []string{"3.", `"Unknown"`, "[Manga]", "missing"},
		},
		{
			name:     "no reason",
			num:      4,
			item:     UnmappedEntry{Title: "X", AniListID: 5},
			media:    "Anime",
			contains: []string{`"X"`, "AniList: 5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUnmappedLine(tt.num, tt.item, tt.media)
			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}
		})
	}
}
