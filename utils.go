package main

import (
	"math"
	"regexp"
	"strings"
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

// Helper functions for title matching to reduce complexity

func exactMatch(t1, t2, titleType string) bool {
	if t1 == "" || t2 == "" {
		return false
	}
	if strings.EqualFold(t1, t2) {
		DPrintf("Exact match found %s: %s == %s", titleType, t1, t2)
		return true
	}
	return false
}

func normalizedMatch(t1, t2, titleType string) bool {
	if t1 == "" || t2 == "" {
		return false
	}
	normalizedA := normalizeTitle(t1)
	normalizedB := normalizeTitle(t2)
	if normalizedA == normalizedB {
		DPrintf("Normalized match found %s: '%s' == '%s' (original: '%s' vs '%s')",
			titleType, normalizedA, normalizedB, t1, t2)
		return true
	}
	return false
}

func fuzzyMatch(t1, t2, titleType string, threshold float64) bool {
	if t1 == "" || t2 == "" {
		return false
	}
	similarity := titleSimilarity(t1, t2)
	if similarity >= threshold {
		DPrintf("Fuzzy match found %s: '%s' ~= '%s' (similarity: %.2f)", titleType, t1, t2, similarity)
		return true
	}
	return false
}

func levenshteinMatch(t1, t2, titleType string, threshold float64) bool {
	if t1 == "" || t2 == "" {
		return false
	}
	similarity := titleLevenshteinSimilarity(t1, t2)
	if similarity >= threshold {
		DPrintf("Levenshtein match found %s: '%s' ~= '%s' (similarity: %.2f)", titleType, t1, t2, similarity)
		return true
	}
	return false
}

// titleMatchingLevels performs multi-level title matching
func titleMatchingLevels(titleEN1, titleJP1, titleRomaji1, titleEN2, titleJP2, titleRomaji2 string) bool {
	// Level 1: Exact case-insensitive title matching
	if exactMatch(titleEN1, titleEN2, "TitleEN") {
		return true
	}
	if exactMatch(titleJP1, titleJP2, "TitleJP") {
		return true
	}
	if exactMatch(titleRomaji1, titleRomaji2, "TitleRomaji") {
		return true
	}

	// Level 2: Normalized exact matching (removes punctuation, brackets, etc.)
	if normalizedMatch(titleEN1, titleEN2, "TitleEN") {
		return true
	}
	if normalizedMatch(titleJP1, titleJP2, "TitleJP") {
		return true
	}
	if normalizedMatch(titleRomaji1, titleRomaji2, "TitleRomaji") {
		return true
	}

	// Level 3: Fuzzy matching with similarity threshold
	const similarityThreshold = 98.0
	if fuzzyMatch(titleEN1, titleEN2, "TitleEN", similarityThreshold) {
		return true
	}
	if fuzzyMatch(titleJP1, titleJP2, "TitleJP", similarityThreshold) {
		return true
	}
	if fuzzyMatch(titleRomaji1, titleRomaji2, "TitleRomaji", similarityThreshold) {
		return true
	}

	// Level 4: Levenshtein distance-based matching
	const levenshteinThreshold = 98.0
	if levenshteinMatch(titleEN1, titleEN2, "TitleEN", levenshteinThreshold) {
		return true
	}
	if levenshteinMatch(titleJP1, titleJP2, "TitleJP", levenshteinThreshold) {
		return true
	}
	if levenshteinMatch(titleRomaji1, titleRomaji2, "TitleRomaji", levenshteinThreshold) {
		return true
	}

	return false
}
