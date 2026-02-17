package main

import (
	"testing"

	"github.com/rl404/verniy"
)

func TestNormalizeScoreForMAL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		score       float64
		format      verniy.ScoreFormat
		expected    int
		description string
	}{
		// POINT_100 format (0-100)
		{"POINT_100: 0", 0, verniy.ScoreFormatPoint100, 0, "zero score"},
		{"POINT_100: 50", 50, verniy.ScoreFormatPoint100, 5, "50/100 = 5/10"},
		{"POINT_100: 85", 85, verniy.ScoreFormatPoint100, 9, "85/100 = 8.5/10 rounds to 9"},
		{"POINT_100: 100", 100, verniy.ScoreFormatPoint100, 10, "100/100 = 10/10"},
		{"POINT_100: 95", 95, verniy.ScoreFormatPoint100, 10, "95/100 = 9.5/10 rounds to 10"},
		{"POINT_100: 45", 45, verniy.ScoreFormatPoint100, 5, "45/100 = 4.5/10 rounds to 5"},
		{"POINT_100: 44", 44, verniy.ScoreFormatPoint100, 4, "44/100 = 4.4/10 rounds to 4"},
		{"POINT_100: 1", 1, verniy.ScoreFormatPoint100, 0, "1/100 = 0.1/10 rounds to 0"},
		{"POINT_100: 5", 5, verniy.ScoreFormatPoint100, 1, "5/100 = 0.5/10 rounds to 1"},

		// POINT_10_DECIMAL format (0-10.0)
		{"POINT_10_DECIMAL: 0", 0, verniy.ScoreFormatPoint100Decimal, 0, "zero score"},
		{"POINT_10_DECIMAL: 5.0", 5.0, verniy.ScoreFormatPoint100Decimal, 5, "5.0/10.0 = 5/10"},
		{"POINT_10_DECIMAL: 8.5", 8.5, verniy.ScoreFormatPoint100Decimal, 9, "8.5/10.0 rounds to 9"},
		{"POINT_10_DECIMAL: 10.0", 10.0, verniy.ScoreFormatPoint100Decimal, 10, "10.0/10.0 = 10/10"},
		{"POINT_10_DECIMAL: 9.5", 9.5, verniy.ScoreFormatPoint100Decimal, 10, "9.5/10.0 rounds to 10"},
		{"POINT_10_DECIMAL: 4.4", 4.4, verniy.ScoreFormatPoint100Decimal, 4, "4.4/10.0 rounds to 4"},
		{"POINT_10_DECIMAL: 4.5", 4.5, verniy.ScoreFormatPoint100Decimal, 5, "4.5/10.0 rounds to 5"},
		{"POINT_10_DECIMAL: 0.1", 0.1, verniy.ScoreFormatPoint100Decimal, 0, "0.1/10.0 rounds to 0"},
		{"POINT_10_DECIMAL: 0.5", 0.5, verniy.ScoreFormatPoint100Decimal, 1, "0.5/10.0 rounds to 1"},

		// POINT_10 format (0-10)
		{"POINT_10: 0", 0, verniy.ScoreFormatPoint10, 0, "zero score"},
		{"POINT_10: 5", 5, verniy.ScoreFormatPoint10, 5, "5/10 = 5/10"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, 8, "8/10 = 8/10"},
		{"POINT_10: 10", 10, verniy.ScoreFormatPoint10, 10, "10/10 = 10/10"},
		{"POINT_10: 1", 1, verniy.ScoreFormatPoint10, 1, "1/10 = 1/10"},

		// POINT_5 format (0-5)
		{"POINT_5: 0", 0, verniy.ScoreFormatPoint5, 0, "zero score"},
		{"POINT_5: 1", 1, verniy.ScoreFormatPoint5, 2, "1/5 = 2/10"},
		{"POINT_5: 2", 2, verniy.ScoreFormatPoint5, 4, "2/5 = 4/10"},
		{"POINT_5: 3", 3, verniy.ScoreFormatPoint5, 6, "3/5 = 6/10"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, 8, "4/5 = 8/10"},
		{"POINT_5: 5", 5, verniy.ScoreFormatPoint5, 10, "5/5 = 10/10"},

		// POINT_3 format (1-3)
		{"POINT_3: 0", 0, verniy.ScoreFormatPoint3, 0, "zero score (special case)"},
		{"POINT_3: 1", 1, verniy.ScoreFormatPoint3, 3, "1/3 = 3.33/10 rounds to 3"},
		{"POINT_3: 2", 2, verniy.ScoreFormatPoint3, 7, "2/3 = 6.67/10 rounds to 7"},
		{"POINT_3: 3", 3, verniy.ScoreFormatPoint3, 10, "3/3 = 10/10"},

		// Edge cases - values over 10 for unknown format
		{"Unknown format: 150", 150, "UNKNOWN_FORMAT", 10, "150 > 10, normalize to 10/10 max"},
		{"Unknown format: 50", 50, "UNKNOWN_FORMAT", 5, "50 > 10, divide by 10 = 5"},
		{"Unknown format: 8", 8, "UNKNOWN_FORMAT", 8, "8 <= 10, keep as is"},

		// Clamping edge cases
		{"POINT_100: over max", 110, verniy.ScoreFormatPoint100, 10, "110/100 = 11 clamped to 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeScoreForMAL(tt.score, tt.format)
			if result != tt.expected {
				t.Errorf("normalizeScoreForMAL(%v, %v) = %v; want %v (%s)",
					tt.score, tt.format, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDenormalizeScoreForAniList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		score       int
		format      verniy.ScoreFormat
		expected    int
		description string
	}{
		// POINT_100 format (0-100)
		{"POINT_100: 0", 0, verniy.ScoreFormatPoint100, 0, "zero score"},
		{"POINT_100: 5", 5, verniy.ScoreFormatPoint100, 50, "5/10 = 50/100"},
		{"POINT_100: 9", 9, verniy.ScoreFormatPoint100, 90, "9/10 = 90/100"},
		{"POINT_100: 10", 10, verniy.ScoreFormatPoint100, 100, "10/10 = 100/100"},
		{"POINT_100: 1", 1, verniy.ScoreFormatPoint100, 10, "1/10 = 10/100"},
		{"POINT_100: 8", 8, verniy.ScoreFormatPoint100, 80, "8/10 = 80/100"},

		// POINT_10_DECIMAL format (0-10.0)
		{"POINT_10_DECIMAL: 0", 0, verniy.ScoreFormatPoint100Decimal, 0, "zero score"},
		{"POINT_10_DECIMAL: 5", 5, verniy.ScoreFormatPoint100Decimal, 5, "5/10 = 5/10"},
		{"POINT_10_DECIMAL: 9", 9, verniy.ScoreFormatPoint100Decimal, 9, "9/10 = 9/10"},
		{"POINT_10_DECIMAL: 10", 10, verniy.ScoreFormatPoint100Decimal, 10, "10/10 = 10/10"},
		{"POINT_10_DECIMAL: 1", 1, verniy.ScoreFormatPoint100Decimal, 1, "1/10 = 1/10"},

		// POINT_10 format (0-10)
		{"POINT_10: 0", 0, verniy.ScoreFormatPoint10, 0, "zero score"},
		{"POINT_10: 5", 5, verniy.ScoreFormatPoint10, 5, "5/10 = 5/10"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, 8, "8/10 = 8/10"},
		{"POINT_10: 10", 10, verniy.ScoreFormatPoint10, 10, "10/10 = 10/10"},
		{"POINT_10: 1", 1, verniy.ScoreFormatPoint10, 1, "1/10 = 1/10"},

		// POINT_5 format (0-5)
		{"POINT_5: 0", 0, verniy.ScoreFormatPoint5, 0, "zero score"},
		{"POINT_5: 2", 2, verniy.ScoreFormatPoint5, 1, "2/10 = 1/5"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, 2, "4/10 = 2/5"},
		{"POINT_5: 6", 6, verniy.ScoreFormatPoint5, 3, "6/10 = 3/5"},
		{"POINT_5: 8", 8, verniy.ScoreFormatPoint5, 4, "8/10 = 4/5"},
		{"POINT_5: 10", 10, verniy.ScoreFormatPoint5, 5, "10/10 = 5/5"},
		{"POINT_5: 5", 5, verniy.ScoreFormatPoint5, 3, "5/10 = 2.5/5 rounds to 3"},
		{"POINT_5: 7", 7, verniy.ScoreFormatPoint5, 4, "7/10 = 3.5/5 rounds to 4"},
		{"POINT_5: 3", 3, verniy.ScoreFormatPoint5, 2, "3/10 = 1.5/5 rounds to 2"},

		// POINT_3 format (1-3)
		{"POINT_3: 0", 0, verniy.ScoreFormatPoint3, 0, "zero score (special case)"},
		{"POINT_3: 1", 1, verniy.ScoreFormatPoint3, 1, "1/10 = 0.3/3 but min is 1"},
		{"POINT_3: 3", 3, verniy.ScoreFormatPoint3, 1, "3/10 = 0.9/3 rounds to 1"},
		{"POINT_3: 4", 4, verniy.ScoreFormatPoint3, 1, "4/10 = 1.2/3 rounds to 1"},
		{"POINT_3: 5", 5, verniy.ScoreFormatPoint3, 2, "5/10 = 1.5/3 rounds to 2"},
		{"POINT_3: 7", 7, verniy.ScoreFormatPoint3, 2, "7/10 = 2.1/3 rounds to 2"},
		{"POINT_3: 8", 8, verniy.ScoreFormatPoint3, 2, "8/10 = 2.4/3 rounds to 2"},
		{"POINT_3: 9", 9, verniy.ScoreFormatPoint3, 3, "9/10 = 2.7/3 rounds to 3"},
		{"POINT_3: 10", 10, verniy.ScoreFormatPoint3, 3, "10/10 = 3/3"},

		// Unknown format
		{"Unknown format: 5", 5, "UNKNOWN_FORMAT", 5, "5/10 stays 5"},
		{"Unknown format: 8", 8, "UNKNOWN_FORMAT", 8, "8/10 stays 8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := denormalizeScoreForAniList(tt.score, tt.format)
			if result != tt.expected {
				t.Errorf("denormalizeScoreForAniList(%v, %v) = %v; want %v (%s)",
					tt.score, tt.format, result, tt.expected, tt.description)
			}
		})
	}
}

func TestNormalizeMangaScoreForMAL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		score       float64
		format      verniy.ScoreFormat
		expected    int
		description string
	}{
		// POINT_100 format (0-100)
		{"POINT_100: 0", 0, verniy.ScoreFormatPoint100, 0, "zero score"},
		{"POINT_100: 50", 50, verniy.ScoreFormatPoint100, 5, "50/100 = 5/10"},
		{"POINT_100: 85", 85, verniy.ScoreFormatPoint100, 9, "85/100 = 8.5/10 rounds to 9"},
		{"POINT_100: 100", 100, verniy.ScoreFormatPoint100, 10, "100/100 = 10/10"},
		{"POINT_100: 95", 95, verniy.ScoreFormatPoint100, 10, "95/100 = 9.5/10 rounds to 10"},
		{"POINT_100: 45", 45, verniy.ScoreFormatPoint100, 5, "45/100 = 4.5/10 rounds to 5"},
		{"POINT_100: 44", 44, verniy.ScoreFormatPoint100, 4, "44/100 = 4.4/10 rounds to 4"},
		{"POINT_100: 1", 1, verniy.ScoreFormatPoint100, 0, "1/100 = 0.1/10 rounds to 0"},
		{"POINT_100: 5", 5, verniy.ScoreFormatPoint100, 1, "5/100 = 0.5/10 rounds to 1"},

		// POINT_10_DECIMAL format (0-10.0)
		{"POINT_10_DECIMAL: 0", 0, verniy.ScoreFormatPoint100Decimal, 0, "zero score"},
		{"POINT_10_DECIMAL: 5.0", 5.0, verniy.ScoreFormatPoint100Decimal, 5, "5.0/10.0 = 5/10"},
		{"POINT_10_DECIMAL: 8.5", 8.5, verniy.ScoreFormatPoint100Decimal, 9, "8.5/10.0 rounds to 9"},
		{"POINT_10_DECIMAL: 10.0", 10.0, verniy.ScoreFormatPoint100Decimal, 10, "10.0/10.0 = 10/10"},
		{"POINT_10_DECIMAL: 9.5", 9.5, verniy.ScoreFormatPoint100Decimal, 10, "9.5/10.0 rounds to 10"},
		{"POINT_10_DECIMAL: 4.4", 4.4, verniy.ScoreFormatPoint100Decimal, 4, "4.4/10.0 rounds to 4"},
		{"POINT_10_DECIMAL: 4.5", 4.5, verniy.ScoreFormatPoint100Decimal, 5, "4.5/10.0 rounds to 5"},
		{"POINT_10_DECIMAL: 0.1", 0.1, verniy.ScoreFormatPoint100Decimal, 0, "0.1/10.0 rounds to 0"},
		{"POINT_10_DECIMAL: 0.5", 0.5, verniy.ScoreFormatPoint100Decimal, 1, "0.5/10.0 rounds to 1"},

		// POINT_10 format (0-10)
		{"POINT_10: 0", 0, verniy.ScoreFormatPoint10, 0, "zero score"},
		{"POINT_10: 5", 5, verniy.ScoreFormatPoint10, 5, "5/10 = 5/10"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, 8, "8/10 = 8/10"},
		{"POINT_10: 10", 10, verniy.ScoreFormatPoint10, 10, "10/10 = 10/10"},
		{"POINT_10: 1", 1, verniy.ScoreFormatPoint10, 1, "1/10 = 1/10"},

		// POINT_5 format (0-5)
		{"POINT_5: 0", 0, verniy.ScoreFormatPoint5, 0, "zero score"},
		{"POINT_5: 1", 1, verniy.ScoreFormatPoint5, 2, "1/5 = 2/10"},
		{"POINT_5: 2", 2, verniy.ScoreFormatPoint5, 4, "2/5 = 4/10"},
		{"POINT_5: 3", 3, verniy.ScoreFormatPoint5, 6, "3/5 = 6/10"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, 8, "4/5 = 8/10"},
		{"POINT_5: 5", 5, verniy.ScoreFormatPoint5, 10, "5/5 = 10/10"},

		// POINT_3 format (1-3)
		{"POINT_3: 0", 0, verniy.ScoreFormatPoint3, 0, "zero score (special case)"},
		{"POINT_3: 1", 1, verniy.ScoreFormatPoint3, 3, "1/3 = 3.33/10 rounds to 3"},
		{"POINT_3: 2", 2, verniy.ScoreFormatPoint3, 7, "2/3 = 6.67/10 rounds to 7"},
		{"POINT_3: 3", 3, verniy.ScoreFormatPoint3, 10, "3/3 = 10/10"},

		// Edge cases - values over 10 for unknown format
		{"Unknown format: 150", 150, "UNKNOWN_FORMAT", 10, "150 > 10, normalize to 10/10 max"},
		{"Unknown format: 50", 50, "UNKNOWN_FORMAT", 5, "50 > 10, divide by 10 = 5"},
		{"Unknown format: 8", 8, "UNKNOWN_FORMAT", 8, "8 <= 10, keep as is"},

		// Clamping edge cases
		{"POINT_100: over max", 110, verniy.ScoreFormatPoint100, 10, "110/100 = 11 clamped to 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeMangaScoreForMAL(tt.score, tt.format)
			if result != tt.expected {
				t.Errorf("normalizeMangaScoreForMAL(%v, %v) = %v; want %v (%s)",
					tt.score, tt.format, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDenormalizeMangaScoreForAniList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		score       int
		format      verniy.ScoreFormat
		expected    int
		description string
	}{
		// POINT_100 format (0-100)
		{"POINT_100: 0", 0, verniy.ScoreFormatPoint100, 0, "zero score"},
		{"POINT_100: 5", 5, verniy.ScoreFormatPoint100, 50, "5/10 = 50/100"},
		{"POINT_100: 9", 9, verniy.ScoreFormatPoint100, 90, "9/10 = 90/100"},
		{"POINT_100: 10", 10, verniy.ScoreFormatPoint100, 100, "10/10 = 100/100"},
		{"POINT_100: 1", 1, verniy.ScoreFormatPoint100, 10, "1/10 = 10/100"},
		{"POINT_100: 8", 8, verniy.ScoreFormatPoint100, 80, "8/10 = 80/100"},

		// POINT_10_DECIMAL format (0-10.0)
		{"POINT_10_DECIMAL: 0", 0, verniy.ScoreFormatPoint100Decimal, 0, "zero score"},
		{"POINT_10_DECIMAL: 5", 5, verniy.ScoreFormatPoint100Decimal, 5, "5/10 = 5/10"},
		{"POINT_10_DECIMAL: 9", 9, verniy.ScoreFormatPoint100Decimal, 9, "9/10 = 9/10"},
		{"POINT_10_DECIMAL: 10", 10, verniy.ScoreFormatPoint100Decimal, 10, "10/10 = 10/10"},
		{"POINT_10_DECIMAL: 1", 1, verniy.ScoreFormatPoint100Decimal, 1, "1/10 = 1/10"},

		// POINT_10 format (0-10)
		{"POINT_10: 0", 0, verniy.ScoreFormatPoint10, 0, "zero score"},
		{"POINT_10: 5", 5, verniy.ScoreFormatPoint10, 5, "5/10 = 5/10"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, 8, "8/10 = 8/10"},
		{"POINT_10: 10", 10, verniy.ScoreFormatPoint10, 10, "10/10 = 10/10"},
		{"POINT_10: 1", 1, verniy.ScoreFormatPoint10, 1, "1/10 = 1/10"},

		// POINT_5 format (0-5)
		{"POINT_5: 0", 0, verniy.ScoreFormatPoint5, 0, "zero score"},
		{"POINT_5: 2", 2, verniy.ScoreFormatPoint5, 1, "2/10 = 1/5"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, 2, "4/10 = 2/5"},
		{"POINT_5: 6", 6, verniy.ScoreFormatPoint5, 3, "6/10 = 3/5"},
		{"POINT_5: 8", 8, verniy.ScoreFormatPoint5, 4, "8/10 = 4/5"},
		{"POINT_5: 10", 10, verniy.ScoreFormatPoint5, 5, "10/10 = 5/5"},
		{"POINT_5: 5", 5, verniy.ScoreFormatPoint5, 3, "5/10 = 2.5/5 rounds to 3"},
		{"POINT_5: 7", 7, verniy.ScoreFormatPoint5, 4, "7/10 = 3.5/5 rounds to 4"},
		{"POINT_5: 3", 3, verniy.ScoreFormatPoint5, 2, "3/10 = 1.5/5 rounds to 2"},

		// POINT_3 format (1-3)
		{"POINT_3: 0", 0, verniy.ScoreFormatPoint3, 0, "zero score (special case)"},
		{"POINT_3: 1", 1, verniy.ScoreFormatPoint3, 1, "1/10 = 0.3/3 but min is 1"},
		{"POINT_3: 3", 3, verniy.ScoreFormatPoint3, 1, "3/10 = 0.9/3 rounds to 1"},
		{"POINT_3: 4", 4, verniy.ScoreFormatPoint3, 1, "4/10 = 1.2/3 rounds to 1"},
		{"POINT_3: 5", 5, verniy.ScoreFormatPoint3, 2, "5/10 = 1.5/3 rounds to 2"},
		{"POINT_3: 7", 7, verniy.ScoreFormatPoint3, 2, "7/10 = 2.1/3 rounds to 2"},
		{"POINT_3: 8", 8, verniy.ScoreFormatPoint3, 2, "8/10 = 2.4/3 rounds to 2"},
		{"POINT_3: 9", 9, verniy.ScoreFormatPoint3, 3, "9/10 = 2.7/3 rounds to 3"},
		{"POINT_3: 10", 10, verniy.ScoreFormatPoint3, 3, "10/10 = 3/3"},

		// Unknown format
		{"Unknown format: 5", 5, "UNKNOWN_FORMAT", 5, "5/10 stays 5"},
		{"Unknown format: 8", 8, "UNKNOWN_FORMAT", 8, "8/10 stays 8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := denormalizeMangaScoreForAniList(tt.score, tt.format)
			if result != tt.expected {
				t.Errorf("denormalizeMangaScoreForAniList(%v, %v) = %v; want %v (%s)",
					tt.score, tt.format, result, tt.expected, tt.description)
			}
		})
	}
}

func TestRoundTripNormalizeDenormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		originalScore float64
		format        verniy.ScoreFormat
		description   string
	}{
		{"POINT_100: 85", 85, verniy.ScoreFormatPoint100, "85 -> 9 -> 90 (loss due to rounding)"},
		{"POINT_100: 90", 90, verniy.ScoreFormatPoint100, "90 -> 9 -> 90 (perfect)"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, "8 -> 8 -> 8 (perfect)"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, "4 -> 8 -> 4 (perfect)"},
		{"POINT_3: 2", 2, verniy.ScoreFormatPoint3, "2 -> 7 -> 2 (perfect)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			normalized := normalizeScoreForMAL(tt.originalScore, tt.format)
			denormalized := denormalizeScoreForAniList(normalized, tt.format)

			t.Logf("Round trip: %.1f -> %d -> %d (%s)",
				tt.originalScore, normalized, denormalized, tt.description)

			// For formats with exact conversion, check round trip
			if tt.format == verniy.ScoreFormatPoint10 ||
				tt.format == verniy.ScoreFormatPoint5 ||
				tt.format == verniy.ScoreFormatPoint3 {
				if denormalized != int(tt.originalScore) {
					t.Errorf("Round trip failed: %.1f -> %d -> %d (expected %d)",
						tt.originalScore, normalized, denormalized, int(tt.originalScore))
				}
			}
		})
	}
}

func TestRoundTripNormalizeDenormalizeManga(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		originalScore float64
		format        verniy.ScoreFormat
		description   string
	}{
		{"POINT_100: 85", 85, verniy.ScoreFormatPoint100, "85 -> 9 -> 90 (loss due to rounding)"},
		{"POINT_100: 90", 90, verniy.ScoreFormatPoint100, "90 -> 9 -> 90 (perfect)"},
		{"POINT_10: 8", 8, verniy.ScoreFormatPoint10, "8 -> 8 -> 8 (perfect)"},
		{"POINT_5: 4", 4, verniy.ScoreFormatPoint5, "4 -> 8 -> 4 (perfect)"},
		{"POINT_3: 2", 2, verniy.ScoreFormatPoint3, "2 -> 7 -> 2 (perfect)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			normalized := normalizeMangaScoreForMAL(tt.originalScore, tt.format)
			denormalized := denormalizeMangaScoreForAniList(normalized, tt.format)

			t.Logf("Round trip: %.1f -> %d -> %d (%s)",
				tt.originalScore, normalized, denormalized, tt.description)

			// For formats with exact conversion, check round trip
			if tt.format == verniy.ScoreFormatPoint10 ||
				tt.format == verniy.ScoreFormatPoint5 ||
				tt.format == verniy.ScoreFormatPoint3 {
				if denormalized != int(tt.originalScore) {
					t.Errorf("Round trip failed: %.1f -> %d -> %d (expected %d)",
						tt.originalScore, normalized, denormalized, int(tt.originalScore))
				}
			}
		})
	}
}
