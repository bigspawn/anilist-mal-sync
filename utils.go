package main

import (
	"math"
	"regexp"
	"strings"

	"github.com/nstratos/go-myanimelist/mal"
)

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D slice for dynamic programming
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1, // deletion
				min(
					matrix[i][j-1]+1,      // insertion
					matrix[i-1][j-1]+cost, // substitution
				),
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

var (
	betweenBracketsRegexp = regexp.MustCompile(`\(.*\)`)
	titleReplacer         = strings.NewReplacer(
		":", "", "!", "", "?", "", ".", "",
		"-", " ", "_", " ", "  ", " ",
	)
)

// normalizeTitle normalizes a title for comparison by removing punctuation and brackets
func normalizeTitle(title string) string {
	normalized := strings.ToLower(title)
	normalized = betweenBracketsRegexp.ReplaceAllString(normalized, "")
	normalized = strings.TrimSpace(titleReplacer.Replace(normalized))

	return normalized
}

// titleSimilarity calculates title similarity percentage between two strings
func titleSimilarity(title1, title2 string) float64 {
	normalized1 := normalizeTitle(title1)
	normalized2 := normalizeTitle(title2)

	if normalized1 == normalized2 {
		return PerfectMatchThreshold
	}

	// Calculate similarity based on common words
	words1 := strings.Fields(normalized1)
	words2 := strings.Fields(normalized2)

	if len(words1) == 0 || len(words2) == 0 {
		return NoMatchThreshold
	}

	// Use a map for O(n+m) complexity instead of O(n*m)
	words2Set := make(map[string]struct{}, len(words2))
	for _, word2 := range words2 {
		words2Set[word2] = struct{}{}
	}

	commonWords := 0
	for _, word1 := range words1 {
		if _, exists := words2Set[word1]; exists {
			commonWords++
		}
	}

	// Calculate similarity percentage
	// Note: totalWords cannot be 0 here since we already checked for empty word slices above
	totalWords := len(words1) + len(words2)
	similarity := (float64(commonWords*2) / float64(totalWords)) * PercentMultiplier

	return similarity
}

// titleLevenshteinSimilarity calculates similarity using Levenshtein distance
func titleLevenshteinSimilarity(title1, title2 string) float64 {
	normalized1 := normalizeTitle(title1)
	normalized2 := normalizeTitle(title2)

	if normalized1 == normalized2 {
		return PerfectMatchThreshold
	}

	distance := levenshteinDistance(normalized1, normalized2)
	maxLen := max(len(normalized1), len(normalized2))

	if maxLen == 0 {
		return PerfectMatchThreshold
	}

	similarity := (1.0 - float64(distance)/float64(maxLen)) * PercentMultiplier
	similarity = max(similarity, NoMatchThreshold)

	return similarity
}

const (
	// MALScoreMax is the maximum score value accepted by MAL (0-10)
	MALScoreMax = 10
	// MALScoreScaleDivisor is used to convert 100-point scale to 10-point scale
	MALScoreScaleDivisor = 10.0
)

// normalizeScoreForMAL converts AniList scores to MAL 0-10 integer scores.
// MAL only accepts integer scores from 0-10, so this function ensures
// all scores are properly normalized regardless of AniList's scoring system.
//
// Normalization rules:
// - If score <= 0 -> 0 (MAL treats 0 as "no score")
// - If score > 10 (e.g. 100-point scale) -> divide by 10 and round
// - Otherwise round to nearest integer (0-10)
func normalizeScoreForMAL(score float64) mal.Score {
	if score <= 0 {
		return mal.Score(0)
	}

	s := score
	// Handle 100-point scale (or any scale > 10)
	if s > MALScoreMax {
		s /= MALScoreScaleDivisor
	}

	// Clamp to valid MAL range (0-10) and round to nearest integer
	s = min(s, float64(MALScoreMax))
	normalized := int(math.Round(s))

	// Clamp to valid range (shouldn't be needed after above checks, but safety)
	normalized = max(0, min(normalized, MALScoreMax))

	return mal.Score(normalized)
}

// checkTitlePair checks if two title strings match using the given comparison function
func checkTitlePair(title1, title2 string, compare func(string, string) bool) bool {
	return title1 != "" && title2 != "" && compare(title1, title2)
}

// titleMatchingLevels performs multi-level title matching with increasing flexibility
func titleMatchingLevels(titleEN1, titleJP1, titleRomaji1, titleEN2, titleJP2, titleRomaji2 string) bool {
	// Level 1: Exact case-insensitive match
	if checkTitlePair(titleEN1, titleEN2, strings.EqualFold) ||
		checkTitlePair(titleJP1, titleJP2, strings.EqualFold) ||
		checkTitlePair(titleRomaji1, titleRomaji2, strings.EqualFold) {
		return true
	}

	// Level 2: Normalized match (removes punctuation, brackets)
	normalizedMatch := func(t1, t2 string) bool {
		return normalizeTitle(t1) == normalizeTitle(t2)
	}
	if checkTitlePair(titleEN1, titleEN2, normalizedMatch) ||
		checkTitlePair(titleJP1, titleJP2, normalizedMatch) ||
		checkTitlePair(titleRomaji1, titleRomaji2, normalizedMatch) {
		return true
	}

	// Level 3: Fuzzy match (word-based similarity)
	fuzzyMatch := func(t1, t2 string) bool {
		return titleSimilarity(t1, t2) >= TitleSimilarityThreshold
	}
	if checkTitlePair(titleEN1, titleEN2, fuzzyMatch) ||
		checkTitlePair(titleJP1, titleJP2, fuzzyMatch) ||
		checkTitlePair(titleRomaji1, titleRomaji2, fuzzyMatch) {
		return true
	}

	// Level 4: Levenshtein distance match
	levenshteinMatch := func(t1, t2 string) bool {
		return titleLevenshteinSimilarity(t1, t2) >= TitleLevenshteinThreshold
	}
	if checkTitlePair(titleEN1, titleEN2, levenshteinMatch) ||
		checkTitlePair(titleJP1, titleJP2, levenshteinMatch) ||
		checkTitlePair(titleRomaji1, titleRomaji2, levenshteinMatch) {
		return true
	}

	return false
}
