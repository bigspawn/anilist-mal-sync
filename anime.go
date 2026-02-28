package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/rl404/verniy"
)

var errStatusUnknown = errors.New("status unknown")

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
	case StatusUnknown:
		return "", errStatusUnknown
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
	case StatusUnknown:
		return ""
	default:
		return ""
	}
}

type Anime struct {
	NumEpisodes int
	IDAnilist   int
	IDMal       int
	Progress    int
	Score       int
	SeasonYear  int
	Status      Status
	TitleEN     string
	TitleJP     string
	TitleRomaji string
	StartedAt   *time.Time
	FinishedAt  *time.Time
	IsFavourite bool
	isReverse   bool // true when used in reverse sync (MAL → AniList)
}

func (a Anime) GetTargetID() TargetID {
	if a.isReverse {
		return TargetID(a.IDAnilist)
	}
	return TargetID(a.IDMal)
}

// GetAniListID returns the AniList ID
func (a Anime) GetAniListID() TargetID {
	return TargetID(a.IDAnilist)
}

// GetMALID returns the MAL ID
func (a Anime) GetMALID() TargetID {
	return TargetID(a.IDMal)
}

func (a Anime) GetSourceID() int {
	if a.isReverse {
		return a.IDMal
	}
	return a.IDAnilist
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
		"StartedAt", a.StartedAt, b.StartedAt,
		"FinishedAt", a.FinishedAt, b.FinishedAt,
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
		return false
	}
	if a.Score != b.Score {
		return false
	}
	if !sameDates(a.StartedAt, b.StartedAt) {
		return false
	}
	// Compare FinishedAt only when status is COMPLETED.
	// Non-completed entries may have stale FinishedAt on AniList
	// that MAL ignores — comparing them would cause infinite
	// update loops since MAL never accepts the date.
	if a.Status == StatusCompleted && !sameDates(a.FinishedAt, b.FinishedAt) {
		return false
	}
	progress := a.Progress == b.Progress
	if a.NumEpisodes == b.NumEpisodes {
		return progress
	}
	if a.NumEpisodes == 0 || b.NumEpisodes == 0 {
		return progress
	}
	if progress && (a.NumEpisodes-b.NumEpisodes != 0) {
		return true
	}

	aa := (a.NumEpisodes - a.Progress)
	bb := (b.NumEpisodes - b.Progress)

	return aa == bb
}

func (a Anime) SameTypeWithTarget(t Target) bool {
	// Type assertion to ensure we're comparing with another Anime
	b, ok := t.(Anime)
	if !ok {
		return false
	}

	// Check if MAL IDs match (critical for reverse sync)
	if a.IDMal > 0 && b.IDMal > 0 && a.IDMal == b.IDMal {
		return true
	}

	// Check if AniList IDs match
	if a.IDAnilist > 0 && b.IDAnilist > 0 && a.IDAnilist == b.IDAnilist {
		return true
	}

	// Use the comprehensive title matching logic
	return a.SameTitleWithTarget(t)
}

func (a Anime) SameTitleWithTarget(t Target) bool {
	b, ok := t.(Anime)
	if !ok {
		return false
	}

	// Check if titles match
	if !titleMatchingLevels(
		a.TitleEN, a.TitleJP, a.TitleRomaji,
		b.TitleEN, b.TitleJP, b.TitleRomaji,
	) {
		return false
	}

	// Additional validation: check episode count if both are known
	// This prevents matching movies (1 ep) with TV series (13+ eps)
	if a.NumEpisodes > 0 && b.NumEpisodes > 0 {
		minEps := a.NumEpisodes
		maxEps := b.NumEpisodes
		if minEps > maxEps {
			minEps, maxEps = maxEps, minEps
		}

		// Calculate percentage difference
		percentDiff := float64(maxEps-minEps) / float64(maxEps) * 100

		// Reject if difference is more than 20%
		if percentDiff > 20.0 {
			return false
		}
	}

	return true
}

// IsPotentiallyIncorrectMatch checks if a match might be incorrect
// Returns true if the match should be rejected
func (a Anime) IsPotentiallyIncorrectMatch(t Target) bool {
	b, ok := t.(Anime)
	if !ok {
		return false
	}

	// If source has a valid MAL ID that matches, trust it
	srcID := a.IDMal
	tgtID := b.IDMal
	if srcID > 0 && srcID == tgtID {
		return false // Valid MAL ID match
	}

	// If source has no MAL ID but target does, and titles don't match exactly
	// This prevents matching random specials/ovies with different titles
	if srcID == 0 && tgtID > 0 && !a.IdenticalTitleMatch(b) {
		return true // Likely incorrect match - different titles
	}

	// Check episode count mismatch
	// If source has 0/unknown episodes but target has many (> 4)
	if (a.NumEpisodes == 0 || a.NumEpisodes == 1) && b.NumEpisodes > 4 {
		// Check if titles are actually different (not just one being a substring)
		if !a.IdenticalTitleMatch(b) {
			return true // Likely incorrect match
		}
	}

	return false
}

// IdenticalTitleMatch checks if titles are truly identical (not just similar)
func (a Anime) IdenticalTitleMatch(b Anime) bool {
	// Exact match on any title field
	if a.TitleEN != "" && a.TitleEN == b.TitleEN {
		return true
	}
	if a.TitleJP != "" && a.TitleJP == b.TitleJP {
		return true
	}
	if a.TitleRomaji != "" && a.TitleRomaji == b.TitleRomaji {
		return true
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
	}

	if a.Status == StatusCompleted && a.FinishedAt != nil {
		opts = append(opts, mal.FinishDate(*a.FinishedAt))
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
	var sb strings.Builder
	sb.WriteString("Anime{")
	fmt.Fprintf(&sb, "IDAnilist: %d, ", a.IDAnilist)
	fmt.Fprintf(&sb, "IDMal: %d, ", a.IDMal)
	fmt.Fprintf(&sb, "TitleEN: %s, ", a.TitleEN)
	fmt.Fprintf(&sb, "TitleJP: %s, ", a.TitleJP)
	fmt.Fprintf(&sb, "MediaListStatus: %s, ", a.Status)
	fmt.Fprintf(&sb, "Score: %d, ", a.Score)
	fmt.Fprintf(&sb, "Progress: %d, ", a.Progress)
	fmt.Fprintf(&sb, "EpisodeNumber: %d, ", a.NumEpisodes)
	fmt.Fprintf(&sb, "SeasonYear: %d, ", a.SeasonYear)
	fmt.Fprintf(&sb, "StartedAt: %s, ", a.StartedAt)
	fmt.Fprintf(&sb, "FinishedAt: %s", a.FinishedAt)
	sb.WriteString("}")
	return sb.String()
}

// newAnimesFromMediaListGroups converts AniList media list groups to domain Anime list.
// reverse=false: entries are forward-sync sources; reverse=true: reverse-sync targets.
func newAnimesFromMediaListGroups(groups []verniy.MediaListGroup, scoreFormat verniy.ScoreFormat, reverse bool) []Anime {
	res := make([]Anime, 0, len(groups))
	for _, group := range groups {
		for _, mediaList := range group.Entries {
			a, err := newAnimeFromMediaListEntry(mediaList, scoreFormat, reverse)
			if err != nil {
				log.Printf("Error creating anime from media list entry: %v", err)
				continue
			}

			res = append(res, a)
		}
	}
	return res
}

func newAnimeFromMediaListEntry(mediaList verniy.MediaList, scoreFormat verniy.ScoreFormat, reverse bool) (Anime, error) {
	if mediaList.Media == nil {
		return Anime{}, errors.New("media is nil")
	}

	if mediaList.Status == nil {
		return Anime{}, errors.New("status is nil")
	}

	if mediaList.Media.Title == nil {
		return Anime{}, errors.New("title is nil")
	}

	var score int
	if mediaList.Score != nil {
		// Normalize AniList score to MAL format (0-10)
		score = normalizeScoreForMAL(*mediaList.Score, scoreFormat)
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

	var isFavourite bool
	if mediaList.Media.IsFavourite != nil {
		isFavourite = *mediaList.Media.IsFavourite
	}

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
		IsFavourite: isFavourite,
		isReverse:   reverse,
	}, nil
}

// newAnimesFromMalAnimes converts MAL anime list to domain Anime list.
// reverse=false: entries are forward-sync targets (MAL IDs as target IDs).
// reverse=true: entries are reverse-sync sources (AniList IDs as target IDs).
func newAnimesFromMalAnimes(malAnimes []mal.Anime, reverse bool) []Anime {
	res := make([]Anime, 0, len(malAnimes))
	for _, malAnime := range malAnimes {
		a, err := newAnimeFromMalAnime(malAnime, reverse)
		if err != nil {
			log.Printf("failed to convert mal anime to anime: %v", err)
			continue
		}
		res = append(res, a)
	}
	return res
}

// newAnimesFromMalUserAnimes converts MAL user anime list to domain Anime list.
// reverse=false: entries are forward-sync targets (MAL IDs as target IDs).
// reverse=true: entries are reverse-sync sources (AniList IDs as target IDs).
func newAnimesFromMalUserAnimes(malAnimes []mal.UserAnime, reverse bool) []Anime {
	res := make([]Anime, 0, len(malAnimes))
	for _, malAnime := range malAnimes {
		a, err := newAnimeFromMalAnime(malAnime.Anime, reverse)
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

func newAnimeFromMalAnime(malAnime mal.Anime, reverse bool) (Anime, error) {
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

	// In reverse sync, IDAnilist=0 triggers name-based search in the strategy chain.
	// In forward sync, IDAnilist=-1 indicates "unknown" (MAL entries as targets don't need it).
	anilistID := -1
	if reverse {
		anilistID = 0
	}

	return Anime{
		NumEpisodes: malAnime.NumEpisodes,
		IDAnilist:   anilistID,
		IDMal:       malAnime.ID,
		Progress:    malAnime.MyListStatus.NumEpisodesWatched,
		Score:       malAnime.MyListStatus.Score, // MAL score is already 0-10 int
		SeasonYear:  malAnime.StartSeason.Year,
		Status:      mapMalAnimeStatusToStatus(malAnime.MyListStatus.Status),
		TitleEN:     titleEN,
		TitleJP:     titleJP,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		IsFavourite: false, // MAL API v2 does not provide favorites
		isReverse:   reverse,
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

// newAnimesFromVerniyMedias converts AniList API search results to domain Anime list.
// reverse=true: entries will be used as reverse-sync targets (AniList IDs as target IDs).
func newAnimesFromVerniyMedias(medias []verniy.Media, reverse bool) []Anime {
	res := make([]Anime, 0, len(medias))
	for _, media := range medias {
		a, err := newAnimeFromVerniyMedia(media, reverse)
		if err != nil {
			log.Printf("failed to convert verniy media to anime: %v", err)
			continue
		}
		res = append(res, a)
	}
	return res
}

func newAnimeFromVerniyMedia(media verniy.Media, reverse bool) (Anime, error) {
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
		StartedAt:   nil,   // Will be set from MAL source
		FinishedAt:  nil,   // Will be set from MAL source
		IsFavourite: false, // Verniy media from search doesn't contain user favorite status
		isReverse:   reverse,
	}, nil
}

// sameDates compares two date pointers at day-level granularity.
// Used to detect if dates need syncing between source and target.
//
// Behavior:
//
//	Source (a) | Target (b) | Result | Action
//	-----------|------------|--------|-----------------------------
//	nil        | nil        | true   | no dates, nothing to do
//	nil        | set        | true   | don't clear existing target date
//	set        | nil        | false  | sync date to target
//	same       | same       | true   | already in sync
//	differ     | differ     | false  | update target date
func sameDates(a, b *time.Time) bool {
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func buildDiffString(pairs ...any) string {
	if len(pairs)%3 != 0 {
		return "Diff{invalid params}"
	}

	sb := strings.Builder{}
	sb.WriteString("Diff{")

	for i := 0; i < len(pairs); i += 3 {
		field, ok := pairs[i].(string)
		if !ok {
			continue
		}
		a := pairs[i+1]
		b := pairs[i+2]

		if !reflect.DeepEqual(a, b) {
			fmt.Fprintf(&sb, "%s: %v -> %v, ", field, a, b)
		}
	}

	sb.WriteString("}")
	return sb.String()
}
