package main

import (
	"math"
	"regexp"
	"strings"

	"github.com/nstratos/go-myanimelist/mal"
)

// EnableScoreNormalization controls whether AniList scores are normalized to MAL scale.
var EnableScoreNormalization = true

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
	return int(math.Min(math.Min(float64(a), float64(b)), float64(c)))
}

var betweenBraketsRegexp = regexp.MustCompile(`\(.*\)`)

// normalizeTitle normalizes a title for better comparison
func normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Remove content in brackets/parentheses
	normalized = betweenBraketsRegexp.ReplaceAllString(normalized, "")

	// Remove common punctuation and extra spaces
	replacements := []string{
		":", "",
		"!", "",
		"?", "",
		".", "",
		"-", " ",
		"_", " ",
		"  ", " ",
	}
	replacer := strings.NewReplacer(replacements...)
	normalized = replacer.Replace(normalized)

	// Trim spaces
	normalized = strings.TrimSpace(normalized)

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
// - If score <= 0 -> 0
// - If score > 10 (e.g. 100-point scale) -> divide by 10 and round
// - Otherwise round to nearest integer
func normalizeScoreForMAL(score float64) mal.Score {
	if score <= 0 {
		return mal.Score(0)
	}

	s := score
	if s > 10 {
		s /= 10.0
	}

	// Clamp to 0..10 then round
	if s < 0 {
		s = 0
	}
	if s > 10 {
		s = 10
	}

	return mal.Score(int(math.Round(s)))
}

// titleMatchingLevels performs multi-level title matching
func titleMatchingLevels(titleEN1, titleJP1, titleRomaji1, titleEN2, titleJP2, titleRomaji2 string) bool {
	// Level 1: Exact case-insensitive title matching
	if titleEN1 != "" && titleEN2 != "" && strings.EqualFold(titleEN1, titleEN2) {
		DPrintf("Exact match found TitleEN: %s == %s", titleEN1, titleEN2)
		return true
	}

	if titleJP1 != "" && titleJP2 != "" && strings.EqualFold(titleJP1, titleJP2) {
		DPrintf("Exact match found TitleJP: %s == %s", titleJP1, titleJP2)
		return true
	}

	if titleRomaji1 != "" && titleRomaji2 != "" && strings.EqualFold(titleRomaji1, titleRomaji2) {
		DPrintf("Exact match found TitleRomaji: %s == %s", titleRomaji1, titleRomaji2)
		return true
	}

	// Level 2: Normalized exact matching (removes punctuation, brackets, etc.)
	if titleEN1 != "" && titleEN2 != "" {
		normalizedA := normalizeTitle(titleEN1)
		normalizedB := normalizeTitle(titleEN2)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleEN: '%s' == '%s' (original: '%s' vs '%s')", normalizedA,
				normalizedB, titleEN1, titleEN2)
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		normalizedA := normalizeTitle(titleJP1)
		normalizedB := normalizeTitle(titleJP2)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleJP: '%s' == '%s' (original: '%s' vs '%s')",
				normalizedA, normalizedB, titleJP1, titleJP2)
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		normalizedA := normalizeTitle(titleRomaji1)
		normalizedB := normalizeTitle(titleRomaji2)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleRomaji: '%s' == '%s' (original: '%s' vs '%s')",
				normalizedA, normalizedB, titleRomaji1, titleRomaji2)
			return true
		}
	}

	// Level 3: Fuzzy matching with similarity threshold
	const similarityThreshold = 98.0
	if titleEN1 != "" && titleEN2 != "" {
		similarity := titleSimilarity(titleEN1, titleEN2)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleEN: '%s' ~= '%s' (similarity: %.2f)", titleEN1, titleEN2, similarity)
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		similarity := titleSimilarity(titleJP1, titleJP2)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleJP: '%s' ~= '%s' (similarity: %.2f)", titleJP1, titleJP2, similarity)
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		similarity := titleSimilarity(titleRomaji1, titleRomaji2)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleRomaji: '%s' ~= '%s' (similarity: %.2f)", titleRomaji1,
				titleRomaji2, similarity)
			return true
		}
	}

	// Level 4: Levenshtein distance-based matching
	const levenshteinThreshold = 98.0
	if titleEN1 != "" && titleEN2 != "" {
		similarity := titleLevenshteinSimilarity(titleEN1, titleEN2)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleEN: '%s' ~= '%s' (similarity: %.2f)", titleEN1, titleEN2, similarity)
			return true
		}
	}

	if titleJP1 != "" && titleJP2 != "" {
		similarity := titleLevenshteinSimilarity(titleJP1, titleJP2)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleJP: '%s' ~= '%s' (similarity: %.2f)", titleJP1, titleJP2, similarity)
			return true
		}
	}

	if titleRomaji1 != "" && titleRomaji2 != "" {
		similarity := titleLevenshteinSimilarity(titleRomaji1, titleRomaji2)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleRomaji: '%s' ~= '%s' (similarity: %.2f)", titleRomaji1,
				titleRomaji2, similarity)
			return true
		}
	}

	return false
}
