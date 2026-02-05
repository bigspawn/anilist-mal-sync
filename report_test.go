package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncReport(t *testing.T) {
	report := NewSyncReport()

	if report.Warnings == nil {
		t.Error("Warnings slice should be initialized")
	}
	if len(report.Warnings) != 0 {
		t.Error("Warnings should be empty initially")
	}
}

func TestSyncReport_AddWarning(t *testing.T) {
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
	report := NewSyncReport()
	report.AddWarning("Test Title", "Test Reason", "Test Detail", "Test Media")

	warning := report.Warnings[0]

	assert.Equal(t, "Test Title", warning.Title)
	assert.Equal(t, "Test Reason", warning.Reason)
	assert.Equal(t, "Test Detail", warning.Detail)
	assert.Equal(t, "Test Media", warning.MediaType)
}
