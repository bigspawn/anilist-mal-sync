package main

import (
	"context"

	"github.com/rl404/verniy"
)

// normalizeScoreForMAL converts AniList score to 0-10 int format
func normalizeScoreForMAL(ctx context.Context, anilistScore float64, scoreFormat verniy.ScoreFormat) int {
	if anilistScore == 0 {
		return 0
	}

	var normalized float64

	switch scoreFormat {
	case verniy.ScoreFormatPoint100:
		normalized = anilistScore / 10.0
		LogDebug(ctx, "Score normalized (POINT_100): %.1f → %.1f", anilistScore, normalized)

	case verniy.ScoreFormatPoint100Decimal: // "POINT_10_DECIMAL"
		normalized = anilistScore

	case verniy.ScoreFormatPoint10:
		normalized = anilistScore

	case verniy.ScoreFormatPoint5:
		normalized = anilistScore * 2.0
		LogDebug(ctx, "Score normalized (POINT_5): %.1f → %.1f", anilistScore, normalized)

	case verniy.ScoreFormatPoint3:
		normalized = (anilistScore / 3.0) * 10.0
		LogDebug(ctx, "Score normalized (POINT_3): %.1f → %.1f", anilistScore, normalized)

	default:
		if anilistScore > 10 {
			normalized = anilistScore / 10.0
			LogDebug(ctx, "Score normalized (unknown): %.1f → %.1f", anilistScore, normalized)
		} else {
			normalized = anilistScore
		}
	}

	// Clamp 0-10
	if normalized > 10 {
		LogDebug(ctx, "Score %.1f > 10, clamping", normalized)
		normalized = 10
	}
	if normalized < 0 {
		normalized = 0
	}

	// Round using "half-up" strategy (0.5 always rounds up)
	// We use int(normalized + 0.5) instead of math.Round() because:
	// - math.Round() uses "banker's rounding" (round half to even): 8.5 → 8
	// - int(x + 0.5) uses "round half up": 8.5 → 9
	// This ensures consistent behavior: 4.5→5, 8.5→9, 9.5→10
	return int(normalized + 0.5)
}

// denormalizeScoreForAniList converts 0-10 int format back to AniList format
func denormalizeScoreForAniList(ctx context.Context, normalizedScore int, scoreFormat verniy.ScoreFormat) int {
	if normalizedScore == 0 {
		return 0
	}

	var result float64

	switch scoreFormat {
	case verniy.ScoreFormatPoint100:
		result = float64(normalizedScore) * 10.0
		LogDebug(ctx, "Score denormalized (POINT_100): %d → %.1f", normalizedScore, result)

	case verniy.ScoreFormatPoint100Decimal: // "POINT_10_DECIMAL"
		result = float64(normalizedScore)

	case verniy.ScoreFormatPoint10:
		result = float64(normalizedScore)

	case verniy.ScoreFormatPoint5:
		result = float64(normalizedScore) / 2.0
		LogDebug(ctx, "Score denormalized (POINT_5): %d → %.1f", normalizedScore, result)

	case verniy.ScoreFormatPoint3:
		result = (float64(normalizedScore) / 10.0) * 3.0
		if result < 1 && normalizedScore > 0 {
			result = 1
		}
		LogDebug(ctx, "Score denormalized (POINT_3): %d → %.1f", normalizedScore, result)

	default:
		result = float64(normalizedScore)
	}

	// Round using "half-up" strategy (see normalizeScoreForMAL comment)
	return int(result + 0.5)
}

// normalizeMangaScoreForMAL converts AniList manga score to 0-10 int format
// Same logic as normalizeScoreForMAL but for manga
func normalizeMangaScoreForMAL(ctx context.Context, anilistScore float64, scoreFormat verniy.ScoreFormat) int {
	return normalizeScoreForMAL(ctx, anilistScore, scoreFormat)
}

// denormalizeMangaScoreForAniList converts 0-10 int format back to AniList manga format
// Same logic as denormalizeScoreForAniList but for manga
func denormalizeMangaScoreForAniList(ctx context.Context, normalizedScore int, scoreFormat verniy.ScoreFormat) int {
	return denormalizeScoreForAniList(ctx, normalizedScore, scoreFormat)
}
