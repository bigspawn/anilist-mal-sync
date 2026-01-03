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

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

var betweenBracketsRegexp = regexp.MustCompile(`\(.*\)`)

// normalizeTitle normalizes a title for comparison by removing punctuation and brackets
func normalizeTitle(title string) string {
	normalized := strings.ToLower(title)
	normalized = betweenBracketsRegexp.ReplaceAllString(normalized, "")
	replacer := strings.NewReplacer(
		":", "", "!", "", "?", "", ".", "",
		"-", " ", "_", " ", "  ", " ",
	)
	normalized = strings.TrimSpace(replacer.Replace(normalized))

	return normalized
}

// titleSimilarity calculates title similarity percentage between two strings
func titleSimilarity(title1, title2 string) float64 {
	normalized1 := normalizeTitle(title1)
	normalized2 := normalizeTitle(title2)

	if normalized1 == normalized2 {
		return 100.0
	}

	// Calculate similarity based on common words
	words1 := strings.Fields(normalized1)
	words2 := strings.Fields(normalized2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	commonWords := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				commonWords++
				break
			}
		}
	}

	// Calculate similarity percentage
	totalWords := len(words1) + len(words2)
	if totalWords == 0 {
		return 0.0
	}

	similarity := (float64(commonWords*2) / float64(totalWords)) * 100.0

	return similarity
}

// titleLevenshteinSimilarity calculates similarity using Levenshtein distance
func titleLevenshteinSimilarity(title1, title2 string) float64 {
	normalized1 := normalizeTitle(title1)
	normalized2 := normalizeTitle(title2)

	if normalized1 == normalized2 {
		return 100.0
	}

	distance := levenshteinDistance(normalized1, normalized2)
	maxLen := len(normalized1)
	if len(normalized2) > maxLen {
		maxLen = len(normalized2)
	}

	if maxLen == 0 {
		return 100.0
	}

	similarity := (1.0 - float64(distance)/float64(maxLen)) * 100.0
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// normalizeScoreForMAL converts AniList scores to MAL 0-10 integer scores.
// MAL only accepts integer scores from 0-10, so this function ensures
// all scores are properly normalized regardless of AniList's scoring system.
//
// Normalization rules:
// - If score <= 0 -> 0 (MAL treats 0 as "no score")
// - If score > 10 (e.g. 100-point scale) -> divide by 10 and round
// - If score > 10 after division -> clamp to 10
// - Otherwise round to nearest integer (0-10)
func normalizeScoreForMAL(score float64) mal.Score {
	if score <= 0 {
		return mal.Score(0)
	}

	s := score
	// Handle 100-point scale (or any scale > 10)
	if s > 10 {
		s /= 10.0
	}

	// Clamp to valid MAL range (0-10)
	if s < 0 {
		s = 0
	}
	if s > 10 {
		s = 10
	}

	// Round to nearest integer (MAL only accepts integers)
	normalized := int(math.Round(s))

	// Final safety check
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 10 {
		normalized = 10
	}

	return mal.Score(normalized)
}

// titleMatchingLevels performs multi-level title matching with increasing flexibility
func titleMatchingLevels(titleEN1, titleJP1, titleRomaji1, titleEN2, titleJP2, titleRomaji2 string) bool {
	// Level 1: Exact case-insensitive match
	if titleEN1 != "" && titleEN2 != "" && strings.EqualFold(titleEN1, titleEN2) {
		return true
	}

	if titleJP1 != "" && titleJP2 != "" && strings.EqualFold(titleJP1, titleJP2) {
		return true
	}

	if titleRomaji1 != "" && titleRomaji2 != "" && strings.EqualFold(titleRomaji1, titleRomaji2) {
		return true
	}

	// Level 2: Normalized match (removes punctuation, brackets)
	if titleEN1 != "" && titleEN2 != "" {
		normalizedA := normalizeTitle(titleEN1)
		normalizedB := normalizeTitle(titleEN2)
		if normalizedA == normalizedB {
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		normalizedA := normalizeTitle(titleJP1)
		normalizedB := normalizeTitle(titleJP2)
		if normalizedA == normalizedB {
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		normalizedA := normalizeTitle(titleRomaji1)
		normalizedB := normalizeTitle(titleRomaji2)
		if normalizedA == normalizedB {
			return true
		}
	}

	// Level 3: Fuzzy match (word-based similarity)
	const similarityThreshold = 98.0
	if titleEN1 != "" && titleEN2 != "" {
		similarity := titleSimilarity(titleEN1, titleEN2)
		if similarity >= similarityThreshold {
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		similarity := titleSimilarity(titleJP1, titleJP2)
		if similarity >= similarityThreshold {
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		similarity := titleSimilarity(titleRomaji1, titleRomaji2)
		if similarity >= similarityThreshold {
			return true
		}
	}

	// Level 4: Levenshtein distance match
	const levenshteinThreshold = 98.0
	if titleEN1 != "" && titleEN2 != "" {
		similarity := titleLevenshteinSimilarity(titleEN1, titleEN2)
		if similarity >= levenshteinThreshold {
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		similarity := titleLevenshteinSimilarity(titleJP1, titleJP2)
		if similarity >= levenshteinThreshold {
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		similarity := titleLevenshteinSimilarity(titleRomaji1, titleRomaji2)
		if similarity >= levenshteinThreshold {
			return true
		}
	}

	return false
}
