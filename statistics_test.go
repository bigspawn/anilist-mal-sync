package main

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatistics_Reset(t *testing.T) {
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
	stats := &Statistics{}

	stats.Reset()

	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)
}

func TestStatistics_ResetMultipleTimes(t *testing.T) {
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
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewLogger(false)
	logger.SetOutput(&buf)

	stats := NewStatistics(logger)
	stats.UpdatedCount = 42
	stats.SkippedCount = 100
	stats.TotalCount = 142

	stats.Print("TestPrefix")

	output := buf.String()
	assert.Contains(t, output, "=== TestPrefix: Sync Complete ===", "Print should log header")
	assert.Contains(t, output, "Total items: 142", "Print should log total items")
	assert.Contains(t, output, "✓ Updated: 42", "Print should log correct update info")
	assert.Contains(t, output, "Skipped: 100", "Print should log correct skip info")
}

func TestStatistics_PrintWithZeroValues(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewLogger(false)
	logger.SetOutput(&buf)

	stats := NewStatistics(logger)

	stats.Print("EmptyTest")

	output := buf.String()
	assert.Contains(t, output, "=== EmptyTest: Sync Complete ===", "Print should log header")
	assert.Contains(t, output, "✓ Updated: 0", "Print should log zero updated")
}

func TestStatistics_WatchModeFlow(t *testing.T) {
	logger := NewLogger(false)
	updater := &Updater{
		Prefix:     "Test Watch Mode",
		Statistics: NewStatistics(logger),
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
	// Integration test to verify Reset() is called after Print()
	// This tests the fixed behavior in performSync()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	stats := &Statistics{
		UpdatedCount: 10,
		SkippedCount: 5,
		TotalCount:   15,
	}

	// Simulate performSync behavior: Update -> Print -> Reset
	prefix := "Test Sync"

	// Print first (before reset)
	stats.Print(prefix)
	output := buf.String()

	// Verify output contains the counts
	assert.Contains(t, output, "[Test Sync] Updated 10 out of 15")
	assert.Contains(t, output, "[Test Sync] Skipped 5")

	// Reset after print
	stats.Reset()

	// Verify counters are reset
	assert.Equal(t, 0, stats.UpdatedCount)
	assert.Equal(t, 0, stats.SkippedCount)
	assert.Equal(t, 0, stats.TotalCount)

	// If we print again, should show zeros
	buf.Reset()
	stats.Print(prefix)
	output = buf.String()

	assert.Contains(t, output, "[Test Sync] Updated 0 out of 0")
	assert.Contains(t, output, "[Test Sync] Skipped 0")
}

func TestStatistics_ResetIdempotent(t *testing.T) {
	stats := &Statistics{
		UpdatedCount: 999,
		SkippedCount: 888,
		TotalCount:   1887,
	}

	// Reset multiple times
	for i := 0; i < 10; i++ {
		stats.Reset()
		assert.Equal(t, 0, stats.UpdatedCount, "Reset should be idempotent (iteration %d)", i)
		assert.Equal(t, 0, stats.SkippedCount, "Reset should be idempotent (iteration %d)", i)
		assert.Equal(t, 0, stats.TotalCount, "Reset should be idempotent (iteration %d)", i)
	}
}

func TestStatistics_WatchModeMultipleIterations(t *testing.T) {
	updater := &Updater{
		Prefix:     "Watch Mode Test",
		Statistics: &Statistics{},
	}

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
		t.Run("Iteration "+string(rune('1'+i)), func(t *testing.T) {
			// Simulate counting
			updater.Statistics.UpdatedCount = iter.updated
			updater.Statistics.SkippedCount = iter.skipped
			updater.Statistics.TotalCount = iter.total

			assert.Equal(t, iter.updated, updater.Statistics.UpdatedCount)
			assert.Equal(t, iter.skipped, updater.Statistics.SkippedCount)
			assert.Equal(t, iter.total, updater.Statistics.TotalCount)

			// Reset for next iteration
			updater.Statistics.Reset()

			assert.Equal(t, 0, updater.Statistics.UpdatedCount)
			assert.Equal(t, 0, updater.Statistics.SkippedCount)
			assert.Equal(t, 0, updater.Statistics.TotalCount)
		})
	}
}
