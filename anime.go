package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/rl404/verniy"
)

const (
	similarityThreshold  = 98.0
	levenshteinThreshold = 98.0
)

var errStatusUnknown = errors.New("status unknown")

var betweenBraketsRegexp = regexp.MustCompile(`\(.*\)`)

// normalizeTitle normalizes a title for better comparison
func normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Remove content in brackets/parentheses
	normalized = betweenBraketsRegexp.ReplaceAllString(normalized, "")

	// Remove common punctuation and special characters
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, "!", "")
	normalized = strings.ReplaceAll(normalized, "?", "")
	normalized = strings.ReplaceAll(normalized, "'", "")
	normalized = strings.ReplaceAll(normalized, "\"", "")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.ReplaceAll(normalized, ".", " ")
	normalized = strings.ReplaceAll(normalized, ",", " ")

	// Replace multiple spaces with single space
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	// Trim whitespace
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// titleSimilarity calculates prefix similarity between two titles
func titleSimilarity(title1, title2 string) float64 {
	if title1 == "" || title2 == "" {
		return 0.0
	}

	norm1 := strings.ReplaceAll(normalizeTitle(title1), " ", "")
	norm2 := strings.ReplaceAll(normalizeTitle(title2), " ", "")

	if norm1 == norm2 {
		return 100.0
	}

	// Ensure s1 is the longer string
	if len(norm1) < len(norm2) {
		norm1, norm2 = norm2, norm1
	}

	if len(norm2) == 0 {
		return 0.0
	}

	// Calculate character-by-character prefix match
	matchCount := 0
	for i, r := range norm1 {
		if i >= len(norm2) {
			break
		}
		if r == rune(norm2[i]) {
			matchCount = i + 1
		} else {
			break
		}
	}

	return float64(matchCount) / float64(len(norm1)) * 100.0
}

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
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// titleLevenshteinSimilarity calculates similarity based on Levenshtein distance
func titleLevenshteinSimilarity(title1, title2 string) float64 {
	if title1 == "" || title2 == "" {
		return 0.0
	}

	norm1 := normalizeTitle(title1)
	norm2 := normalizeTitle(title2)

	if norm1 == norm2 {
		return 100.0
	}

	distance := levenshteinDistance(norm1, norm2)
	maxLen := len(norm1)
	if len(norm2) > maxLen {
		maxLen = len(norm2)
	}

	if maxLen == 0 {
		return 100.0
	}

	// Convert distance to similarity percentage
	similarity := (1.0 - float64(distance)/float64(maxLen)) * 100.0
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

type Status string

const (
	StatusWatching    Status = "watching"
	StatusCompleted   Status = "completed"
	StatusOnHold      Status = "on_hold"
	StatusDropped     Status = "dropped"
	StatusPlanToWatch Status = "plan_to_watch"
	StatusUnknown     Status = "unknown"
)

func (s Status) GetMalStatus() (mal.AnimeStatus, error) {
	switch s {
	case StatusWatching:
		return mal.AnimeStatusWatching, nil
	case StatusCompleted:
		return mal.AnimeStatusCompleted, nil
	case StatusOnHold:
		return mal.AnimeStatusOnHold, nil
	case StatusDropped:
		return mal.AnimeStatusDropped, nil
	case StatusPlanToWatch:
		return mal.AnimeStatusPlanToWatch, nil
	default:
		return "", errStatusUnknown
	}
}

func (s Status) GetAnilistStatus() string {
	switch s {
	case StatusWatching:
		return "CURRENT"
	case StatusCompleted:
		return "COMPLETED"
	case StatusOnHold:
		return "PAUSED"
	case StatusDropped:
		return "DROPPED"
	case StatusPlanToWatch:
		return "PLANNING"
	default:
		return "PLANNING"
	}
}

type Anime struct {
	NumEpisodes int
	IDAnilist   int
	IDMal       int
	Progress    int
	Score       float64
	SeasonYear  int
	Status      Status
	TitleEN     string
	TitleJP     string
	TitleRomaji string
	StartedAt   *time.Time
	FinishedAt  *time.Time
}

func (a Anime) GetTargetID() TargetID {
	if *reverseDirection {
		return TargetID(a.IDAnilist)
	}
	return TargetID(a.IDMal)
}

func (a Anime) GetStatusString() string {
	return string(a.Status)
}

func (a Anime) GetStringDiffWithTarget(t Target) string {
	b, ok := t.(Anime)
	if !ok {
		return "Diff{undefined}"
	}

	return buildDiffString(
		"Status", a.Status, b.Status,
		"Score", a.Score, b.Score,
		"Progress", a.Progress, b.Progress,
		"NumEpisodes", a.NumEpisodes, b.NumEpisodes,
		"TitleEN", a.TitleEN, b.TitleEN,
		"TitleJP", a.TitleJP, b.TitleJP,
		"TitleRomaji", a.TitleRomaji, b.TitleRomaji,
	)
}

func (a Anime) SameProgressWithTarget(t Target) bool {
	b, ok := t.(Anime)
	if !ok {
		return false
	}

	if a.Status != b.Status {
		DPrintf("Status: %s != %s", a.Status, b.Status)
		return false
	}
	if a.Score != b.Score {
		DPrintf("Score: %f != %f", a.Score, b.Score)
		return false
	}
	progress := a.Progress == b.Progress
	if a.NumEpisodes == b.NumEpisodes {
		DPrintf("Equal number of episodes: %d == %d", a.NumEpisodes, b.NumEpisodes)
		DPrintf("Progress: %t", progress)
		return progress
	}
	if a.NumEpisodes == 0 || b.NumEpisodes == 0 {
		DPrintf("One of the anime has 0 episodes: %d, %d", a.NumEpisodes, b.NumEpisodes)
		DPrintf("Progress: %t", progress)
		return progress
	}
	if progress && (a.NumEpisodes-b.NumEpisodes != 0) {
		DPrintf("Both anime have 0 progress but different number of episodes: %d, %d", a.NumEpisodes, b.NumEpisodes)
		return true
	}

	aa := (a.NumEpisodes - a.Progress)
	bb := (b.NumEpisodes - b.Progress)

	DPrintf("Number of episodes: %d, %d", a.NumEpisodes, b.NumEpisodes)
	DPrintf("Progress: %d, %d", a.Progress, b.Progress)
	DPrintf("Progress: %d == %d", aa, bb)

	return aa == bb
}

func (a Anime) SameTypeWithTarget(t Target) bool {
	// First check: Compare target IDs
	if a.GetTargetID() == t.GetTargetID() {
		return true
	}

	// Type assertion to ensure we're comparing with another Anime
	_, ok := t.(Anime)
	if !ok {
		return false
	}

	// Use the comprehensive title matching logic
	return a.SameTitleWithTarget(t)
}

func (a Anime) SameTitleWithTarget(t Target) bool {
	b, ok := t.(Anime)
	if !ok {
		return false
	}

	// Level 1: Exact case-insensitive title matching
	if a.TitleEN != "" && b.TitleEN != "" && strings.EqualFold(a.TitleEN, b.TitleEN) {
		DPrintf("Exact match found TitleEN: %s == %s", a.TitleEN, b.TitleEN)
		return true
	}

	if a.TitleJP != "" && b.TitleJP != "" && strings.EqualFold(a.TitleJP, b.TitleJP) {
		DPrintf("Exact match found TitleJP: %s == %s", a.TitleJP, b.TitleJP)
		return true
	}

	if a.TitleRomaji != "" && b.TitleRomaji != "" && strings.EqualFold(a.TitleRomaji, b.TitleRomaji) {
		DPrintf("Exact match found TitleRomaji: %s == %s", a.TitleRomaji, b.TitleRomaji)
		return true
	}

	// Level 2: Normalized exact matching (removes punctuation, brackets, etc.)
	if a.TitleEN != "" && b.TitleEN != "" {
		normalizedA := normalizeTitle(a.TitleEN)
		normalizedB := normalizeTitle(b.TitleEN)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleEN: '%s' == '%s' (original: '%s' vs '%s')", normalizedA, normalizedB, a.TitleEN, b.TitleEN)
			return true
		}
	}

	if a.TitleJP != "" && b.TitleJP != "" {
		normalizedA := normalizeTitle(a.TitleJP)
		normalizedB := normalizeTitle(b.TitleJP)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleJP: '%s' == '%s' (original: '%s' vs '%s')", normalizedA, normalizedB, a.TitleJP, b.TitleJP)
			return true
		}
	}

	if a.TitleRomaji != "" && b.TitleRomaji != "" {
		normalizedA := normalizeTitle(a.TitleRomaji)
		normalizedB := normalizeTitle(b.TitleRomaji)
		if normalizedA == normalizedB {
			DPrintf("Normalized match found TitleRomaji: '%s' == '%s' (original: '%s' vs '%s')", normalizedA, normalizedB, a.TitleRomaji, b.TitleRomaji)
			return true
		}
	}

	// Level 4: Fuzzy matching with similarity threshold
	if a.TitleEN != "" && b.TitleEN != "" {
		similarity := titleSimilarity(a.TitleEN, b.TitleEN)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleEN: '%s' ~= '%s' (similarity: %.2f)", a.TitleEN, b.TitleEN, similarity)
			return true
		}
	}

	if a.TitleJP != "" && b.TitleJP != "" {
		similarity := titleSimilarity(a.TitleJP, b.TitleJP)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleJP: '%s' ~= '%s' (similarity: %.2f)", a.TitleJP, b.TitleJP, similarity)
			return true
		}
	}

	if a.TitleRomaji != "" && b.TitleRomaji != "" {
		similarity := titleSimilarity(a.TitleRomaji, b.TitleRomaji)
		if similarity >= similarityThreshold {
			DPrintf("Fuzzy match found TitleRomaji: '%s' ~= '%s' (similarity: %.2f)", a.TitleRomaji, b.TitleRomaji, similarity)
			return true
		}
	}

	// Level 5: Levenshtein distance-based matching
	if a.TitleEN != "" && b.TitleEN != "" {
		similarity := titleLevenshteinSimilarity(a.TitleEN, b.TitleEN)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleEN: '%s' ~= '%s' (similarity: %.2f)", a.TitleEN, b.TitleEN, similarity)
			return true
		}
	}

	if a.TitleJP != "" && b.TitleJP != "" {
		similarity := titleLevenshteinSimilarity(a.TitleJP, b.TitleJP)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleJP: '%s' ~= '%s' (similarity: %.2f)", a.TitleJP, b.TitleJP, similarity)
			return true
		}
	}

	if a.TitleRomaji != "" && b.TitleRomaji != "" {
		similarity := titleLevenshteinSimilarity(a.TitleRomaji, b.TitleRomaji)
		if similarity >= levenshteinThreshold {
			DPrintf("Levenshtein match found TitleRomaji: '%s' ~= '%s' (similarity: %.2f)", a.TitleRomaji, b.TitleRomaji, similarity)
			return true
		}
	}

	return false
}

func (a Anime) GetUpdateOptions() []mal.UpdateMyAnimeListStatusOption {
	st, err := a.Status.GetMalStatus()
	if err != nil {
		log.Printf("Error getting MAL status: %v", err)
		return nil
	}

	opts := []mal.UpdateMyAnimeListStatusOption{
		st,
		mal.Score(a.Score),
		mal.NumEpisodesWatched(a.Progress),
	}

	if a.StartedAt != nil {
		opts = append(opts, mal.StartDate(*a.StartedAt))
	} else {
		opts = append(opts, mal.StartDate(time.Time{}))
	}

	if a.Status == StatusCompleted && a.FinishedAt != nil {
		opts = append(opts, mal.FinishDate(*a.FinishedAt))
	} else {
		opts = append(opts, mal.FinishDate(time.Time{}))
	}

	return opts
}

func (a Anime) GetTitle() string {
	if a.TitleEN != "" {
		return a.TitleEN
	}
	if a.TitleJP != "" {
		return a.TitleJP
	}
	return a.TitleRomaji
}

func (a Anime) String() string {
	sb := strings.Builder{}
	sb.WriteString("Anime{")
	sb.WriteString(fmt.Sprintf("IDAnilist: %d, ", a.IDAnilist))
	sb.WriteString(fmt.Sprintf("IDMal: %d, ", a.IDMal))
	sb.WriteString(fmt.Sprintf("TitleEN: %s, ", a.TitleEN))
	sb.WriteString(fmt.Sprintf("TitleJP: %s, ", a.TitleJP))
	sb.WriteString(fmt.Sprintf("MediaListStatus: %s, ", a.Status))
	sb.WriteString(fmt.Sprintf("Score: %f, ", a.Score))
	sb.WriteString(fmt.Sprintf("Progress: %d, ", a.Progress))
	sb.WriteString(fmt.Sprintf("EpisodeNumber: %d, ", a.NumEpisodes))
	sb.WriteString(fmt.Sprintf("SeasonYear: %d, ", a.SeasonYear))
	sb.WriteString(fmt.Sprintf("StartedAt: %s, ", a.StartedAt))
	sb.WriteString(fmt.Sprintf("FinishedAt: %s", a.FinishedAt))
	sb.WriteString("}")
	return sb.String()
}

func newAnimesFromMediaListGroups(groups []verniy.MediaListGroup) []Anime {
	res := make([]Anime, 0, len(groups))
	for _, group := range groups {
		for _, mediaList := range group.Entries {
			a, err := newAnimeFromMediaListEntry(mediaList)
			if err != nil {
				log.Printf("Error creating anime from media list entry: %v", err)
				continue
			}

			res = append(res, a)
		}
	}
	return res
}

func newAnimeFromMediaListEntry(mediaList verniy.MediaList) (Anime, error) {
	if mediaList.Media == nil {
		return Anime{}, errors.New("media is nil")
	}

	if mediaList.Status == nil {
		return Anime{}, errors.New("status is nil")
	}

	if mediaList.Media.Title == nil {
		return Anime{}, errors.New("title is nil")
	}

	var score float64
	if mediaList.Score != nil {
		score = *mediaList.Score
	}

	var progress int
	if mediaList.Progress != nil {
		progress = *mediaList.Progress
	}

	var titleEN string
	if mediaList.Media.Title.English != nil {
		titleEN = *mediaList.Media.Title.English
	}

	var titleJP string
	if mediaList.Media.Title.Native != nil {
		titleJP = *mediaList.Media.Title.Native
	}

	var episodeNumber int
	if mediaList.Media.Episodes != nil {
		episodeNumber = *mediaList.Media.Episodes
	}

	var year int
	if mediaList.Media.SeasonYear != nil {
		year = *mediaList.Media.SeasonYear
	}

	var idMal int
	if mediaList.Media.IDMAL != nil {
		idMal = *mediaList.Media.IDMAL
	}

	var romajiTitle string
	if mediaList.Media.Title.Romaji != nil {
		romajiTitle = *mediaList.Media.Title.Romaji
	}

	startedAt := convertFuzzyDateToTimeOrNow(mediaList.StartedAt)
	finishedAt := convertFuzzyDateToTimeOrNow(mediaList.CompletedAt)

	return Anime{
		NumEpisodes: episodeNumber,
		IDAnilist:   mediaList.Media.ID,
		IDMal:       idMal,
		Progress:    progress,
		Score:       score,
		SeasonYear:  year,
		Status:      mapVerniyStatusToStatus(*mediaList.Status),
		TitleEN:     titleEN,
		TitleJP:     titleJP,
		TitleRomaji: romajiTitle,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
	}, nil
}

func newAnimesFromMalAnimes(malAnimes []mal.Anime) []Anime {
	res := make([]Anime, 0, len(malAnimes))
	for _, malAnime := range malAnimes {
		a, err := newAnimeFromMalAnime(malAnime)
		if err != nil {
			log.Printf("failed to convert mal anime to anime: %v", err)
			continue
		}
		res = append(res, a)
	}
	return res
}

func newAnimesFromMalUserAnimes(malAnimes []mal.UserAnime) []Anime {
	res := make([]Anime, 0, len(malAnimes))
	for _, malAnime := range malAnimes {
		a, err := newAnimeFromMalAnime(malAnime.Anime)
		if err != nil {
			log.Printf("failed to convert mal anime to anime: %v", err)
			continue
		}
		res = append(res, a)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].GetStatusString() < res[j].GetStatusString()
	})
	return res
}

func newAnimeFromMalAnime(malAnime mal.Anime) (Anime, error) {
	if malAnime.ID == 0 {
		return Anime{}, errors.New("ID is nil")
	}

	startedAt := parseDateOrNow(malAnime.MyListStatus.StartDate)
	finishedAt := parseDateOrNow(malAnime.MyListStatus.FinishDate)

	titleEN := malAnime.Title
	if malAnime.AlternativeTitles.En != "" {
		titleEN = malAnime.AlternativeTitles.En
	}

	titleJP := malAnime.Title
	if malAnime.AlternativeTitles.Ja != "" {
		titleJP = malAnime.AlternativeTitles.Ja
	}

	// In reverse sync mode, we need to leave AniList ID as 0 so the updater can find it by name
	anilistID := -1
	if *reverseDirection {
		anilistID = 0 // This will trigger name-based search in reverse sync
	}

	return Anime{
		NumEpisodes: malAnime.NumEpisodes,
		IDAnilist:   anilistID,
		IDMal:       malAnime.ID,
		Progress:    malAnime.MyListStatus.NumEpisodesWatched,
		Score:       float64(malAnime.MyListStatus.Score),
		SeasonYear:  malAnime.StartSeason.Year,
		Status:      mapMalAnimeStatusToStatus(malAnime.MyListStatus.Status),
		TitleEN:     titleEN,
		TitleJP:     titleJP,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
	}, nil
}

func mapVerniyStatusToStatus(s verniy.MediaListStatus) Status {
	switch s {
	case verniy.MediaListStatusCurrent:
		return StatusWatching
	case verniy.MediaListStatusCompleted:
		return StatusCompleted
	case verniy.MediaListStatusPaused:
		return StatusOnHold
	case verniy.MediaListStatusDropped:
		return StatusDropped
	case verniy.MediaListStatusPlanning:
		return StatusPlanToWatch
	case verniy.MediaListStatusRepeating:
		return StatusWatching // TODO: handle repeating correctly
	default:
		return StatusUnknown
	}
}

func mapMalAnimeStatusToStatus(s mal.AnimeStatus) Status {
	switch s {
	case mal.AnimeStatusWatching:
		return StatusWatching
	case mal.AnimeStatusCompleted:
		return StatusCompleted
	case mal.AnimeStatusOnHold:
		return StatusOnHold
	case mal.AnimeStatusDropped:
		return StatusDropped
	case mal.AnimeStatusPlanToWatch:
		return StatusPlanToWatch
	default:
		return StatusUnknown
	}
}

func convertFuzzyDateToTimeOrNow(fd *verniy.FuzzyDate) *time.Time {
	if fd == nil || fd.Year == nil || fd.Month == nil || fd.Day == nil {
		return nil
	}
	d := time.Date(
		*fd.Year,
		time.Month(*fd.Month),
		*fd.Day,
		0, 0, 0, 0,
		time.UTC,
	)
	return &d
}

func parseDateOrNow(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	parsedTime, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return nil
	}
	parsedTime = parsedTime.UTC().Truncate(24 * time.Hour)
	return &parsedTime
}

func newTargetsFromAnimes(animes []Anime) []Target {
	res := make([]Target, 0, len(animes))
	for _, anime := range animes {
		res = append(res, anime)
	}
	return res
}

func newSourcesFromAnimes(animes []Anime) []Source {
	res := make([]Source, 0, len(animes))
	for _, anime := range animes {
		res = append(res, anime)
	}
	return res
}

func newAnimesFromVerniyMedias(medias []verniy.Media) []Anime {
	res := make([]Anime, 0, len(medias))
	for _, media := range medias {
		a, err := newAnimeFromVerniyMedia(media)
		if err != nil {
			log.Printf("failed to convert verniy media to anime: %v", err)
			continue
		}
		res = append(res, a)
	}
	return res
}

func newAnimeFromVerniyMedia(media verniy.Media) (Anime, error) {
	if media.ID == 0 {
		return Anime{}, errors.New("ID is 0")
	}

	var titleEN string
	if media.Title != nil && media.Title.English != nil {
		titleEN = *media.Title.English
	}

	var titleJP string
	if media.Title != nil && media.Title.Native != nil {
		titleJP = *media.Title.Native
	}

	var romajiTitle string
	if media.Title != nil && media.Title.Romaji != nil {
		romajiTitle = *media.Title.Romaji
	}

	var episodeNumber int
	if media.Episodes != nil {
		episodeNumber = *media.Episodes
	}

	var year int
	if media.SeasonYear != nil {
		year = *media.SeasonYear
	}

	var idMal int
	if media.IDMAL != nil {
		idMal = *media.IDMAL
	}

	return Anime{
		NumEpisodes: episodeNumber,
		IDAnilist:   media.ID,
		IDMal:       idMal,
		Progress:    0, // Will be set from MAL source
		Score:       0, // Will be set from MAL source
		SeasonYear:  year,
		Status:      StatusUnknown, // Will be set from MAL source
		TitleEN:     titleEN,
		TitleJP:     titleJP,
		TitleRomaji: romajiTitle,
		StartedAt:   nil, // Will be set from MAL source
		FinishedAt:  nil, // Will be set from MAL source
	}, nil
}

func buildDiffString(pairs ...any) string {
	if len(pairs)%3 != 0 {
		return "Diff{invalid params}"
	}

	sb := strings.Builder{}
	sb.WriteString("Diff{")

	for i := 0; i < len(pairs); i += 3 {
		field := pairs[i].(string)
		a := pairs[i+1]
		b := pairs[i+2]

		if !reflect.DeepEqual(a, b) {
			sb.WriteString(fmt.Sprintf("%s: %v -> %v, ", field, a, b))
		}
	}

	sb.WriteString("}")
	return sb.String()
}
