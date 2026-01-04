package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/rl404/verniy"
)

var errMangaStatusUnknown = errors.New("manga status unknown")

type MangaStatus string

const (
	MangaStatusReading    MangaStatus = "reading"
	MangaStatusCompleted  MangaStatus = "completed"
	MangaStatusOnHold     MangaStatus = "on_hold"
	MangaStatusDropped    MangaStatus = "dropped"
	MangaStatusPlanToRead MangaStatus = "plan_to_read"
	MangaStatusRepeating  MangaStatus = "repeating" // AniList-specific: rereading
	MangaStatusUnknown    MangaStatus = "unknown"
)

func (s MangaStatus) GetMalStatus() (mal.MangaStatus, error) {
	switch s {
	case MangaStatusReading:
		return mal.MangaStatusReading, nil
	case MangaStatusCompleted:
		return mal.MangaStatusCompleted, nil
	case MangaStatusOnHold:
		return mal.MangaStatusOnHold, nil
	case MangaStatusDropped:
		return mal.MangaStatusDropped, nil
	case MangaStatusPlanToRead:
		return mal.MangaStatusPlanToRead, nil
	case MangaStatusRepeating:
		return mal.MangaStatusReading, nil
	case MangaStatusUnknown:
		return "", errMangaStatusUnknown
	default:
		return "", errMangaStatusUnknown
	}
}

func (s MangaStatus) GetAnilistStatus() string {
	switch s {
	case MangaStatusReading:
		return AnilistStatusCurrent
	case MangaStatusCompleted:
		return AnilistStatusCompleted
	case MangaStatusOnHold:
		return AnilistStatusPaused
	case MangaStatusDropped:
		return AnilistStatusDropped
	case MangaStatusPlanToRead:
		return AnilistStatusPlanning
	case MangaStatusRepeating:
		return AnilistStatusRepeating
	case MangaStatusUnknown:
		return ""
	default:
		return ""
	}
}

type Manga struct {
	IDAnilist        int
	IDMal            int
	Progress         int
	ProgressVolumes  int
	Score            float64
	Status           MangaStatus
	TitleEN          string
	TitleJP          string
	TitleRomaji      string
	Chapters         int
	Volumes          int
	StartedAt        *time.Time
	FinishedAt       *time.Time
	reverseDirection bool // true when syncing MAL -> AniList
}

func (m Manga) GetTargetID() TargetID {
	if m.reverseDirection {
		return TargetID(m.IDAnilist)
	}
	return TargetID(m.IDMal)
}

func (m Manga) GetStatusString() string {
	return string(m.Status)
}

func (m Manga) GetStringDiffWithTarget(t Target) string {
	b, ok := t.(Manga)
	if !ok {
		return "Diff{undefined}"
	}

	return buildDiffString(
		"Status", m.Status, b.Status,
		"Score", m.Score, b.Score,
		"Progress", m.Progress, b.Progress,
		"ProgressVolumes", m.ProgressVolumes, b.ProgressVolumes,
	)
}

func (m Manga) SameProgressWithTarget(t Target) bool {
	b, ok := t.(Manga)
	if !ok {
		return false
	}

	if m.Status != b.Status {
		return false
	}
	if m.Score != b.Score {
		return false
	}
	if m.Progress != b.Progress {
		return false
	}
	if m.ProgressVolumes != b.ProgressVolumes {
		return false
	}

	return true
}

func (m Manga) SameTypeWithTarget(t Target) bool {
	// First check: Compare target IDs (respects reverseDirection)
	if m.GetTargetID() == t.GetTargetID() {
		return true
	}

	// Type assertion to ensure we're comparing with another Manga
	b, ok := t.(Manga)
	if !ok {
		return false
	}

	// Use the comprehensive title matching logic
	if m.SameTitleWithTarget(t) {
		return true
	}

	// Fallback: Check if chapters and volumes match
	if m.Chapters == b.Chapters && m.Volumes == b.Volumes {
		// NOTE: some mangas are joined in MAL in the same entry in Volumes, but it is separated in Anilist.
		// Skip it for now.
		return true
	}

	return false
}

func (m Manga) SameTitleWithTarget(t Target) bool {
	b, ok := t.(Manga)
	if !ok {
		return false
	}

	return titleMatchingLevels(m.TitleEN, m.TitleJP, m.TitleRomaji, b.TitleEN, b.TitleJP, b.TitleRomaji)
}

// GetTitle returns the best available title, preferring English, then Japanese, then Romaji.
// Returns an empty string if all title fields are empty.
func (m Manga) GetTitle() string {
	if m.TitleEN != "" {
		return m.TitleEN
	}
	if m.TitleJP != "" {
		return m.TitleJP
	}
	return m.TitleRomaji
}

func (m Manga) String() string {
	sb := strings.Builder{}
	sb.WriteString("Manga{")
	fmt.Fprintf(&sb, "IDAnilist: %d, ", m.IDAnilist)
	fmt.Fprintf(&sb, "IDMal: %d, ", m.IDMal)
	fmt.Fprintf(&sb, "TitleEN: %s, ", m.TitleEN)
	fmt.Fprintf(&sb, "TitleJP: %s, ", m.TitleJP)
	fmt.Fprintf(&sb, "Status: %s, ", m.Status)
	fmt.Fprintf(&sb, "Score: %f, ", m.Score)
	fmt.Fprintf(&sb, "Progress: %d, ", m.Progress)
	fmt.Fprintf(&sb, "ProgressVolumes: %d, ", m.ProgressVolumes)
	fmt.Fprintf(&sb, "Chapters: %d, ", m.Chapters)
	fmt.Fprintf(&sb, "Volumes: %d, ", m.Volumes)
	fmt.Fprintf(&sb, "StartedAt: %s, ", m.StartedAt)
	fmt.Fprintf(&sb, "FinishedAt: %s", m.FinishedAt)
	sb.WriteString("}")
	return sb.String()
}

func (m Manga) GetUpdateOptions() []mal.UpdateMyMangaListStatusOption {
	st, err := m.Status.GetMalStatus()
	if err != nil {
		log.Printf("Error getting MAL status for manga '%s' (status: %s): %v", m.GetTitle(), m.Status, err)
		// Return empty slice instead of nil to prevent issues, but log the error
		// The update will be skipped by the caller if opts is empty
		return []mal.UpdateMyMangaListStatusOption{}
	}

	// Always normalize scores for MAL (MAL only accepts 0-10 integer scores)
	// If score is 0, don't send it (MAL treats 0 as "no score")
	var scoreOption mal.Score
	if m.Score > 0 {
		scoreOption = normalizeScoreForMAL(m.Score)
	} else {
		scoreOption = mal.Score(0)
	}

	// Pre-allocate with capacity 6 (base 4 + start date + finish date)
	opts := make([]mal.UpdateMyMangaListStatusOption, 4, 6)
	opts[0] = st
	opts[1] = scoreOption
	opts[2] = mal.NumChaptersRead(m.Progress)
	opts[3] = mal.NumVolumesRead(m.ProgressVolumes)

	if m.StartedAt != nil {
		opts = append(opts, mal.StartDate(*m.StartedAt))
	} else {
		opts = append(opts, mal.StartDate(time.Time{}))
	}

	if m.Status == MangaStatusCompleted && m.FinishedAt != nil {
		opts = append(opts, mal.FinishDate(*m.FinishedAt))
	} else {
		opts = append(opts, mal.FinishDate(time.Time{}))
	}

	return opts
}

func newMangaFromMediaListEntry(mediaList verniy.MediaList, reverseDirection bool) (Manga, error) {
	if mediaList.Media == nil {
		return Manga{}, errors.New("media is nil")
	}

	if mediaList.Status == nil {
		return Manga{}, errors.New("status is nil")
	}

	if mediaList.Media.Title == nil {
		return Manga{}, errors.New("title is nil")
	}

	var score float64
	if mediaList.Score != nil {
		score = *mediaList.Score
	}

	var progress int
	if mediaList.Progress != nil {
		progress = *mediaList.Progress
	}

	var progressVolumes int
	if mediaList.ProgressVolumes != nil {
		progressVolumes = *mediaList.ProgressVolumes
	}

	var titleEN string
	if mediaList.Media.Title.English != nil {
		titleEN = *mediaList.Media.Title.English
	}

	var titleJP string
	if mediaList.Media.Title.Native != nil {
		titleJP = *mediaList.Media.Title.Native
	}

	var idMal int
	if mediaList.Media.IDMAL != nil {
		idMal = *mediaList.Media.IDMAL
	}

	var romajiTitle string
	if mediaList.Media.Title.Romaji != nil {
		romajiTitle = *mediaList.Media.Title.Romaji
	}

	var chapters int
	if mediaList.Media.Chapters != nil {
		chapters = *mediaList.Media.Chapters
	}

	var volumes int
	if mediaList.Media.Volumes != nil {
		volumes = *mediaList.Media.Volumes
	}

	startedAt := convertFuzzyDateToTimeOrNow(mediaList.StartedAt)
	finishedAt := convertFuzzyDateToTimeOrNow(mediaList.CompletedAt)

	return Manga{
		IDAnilist:        mediaList.Media.ID,
		IDMal:            idMal,
		Progress:         progress,
		ProgressVolumes:  progressVolumes,
		Score:            score,
		Status:           mapAnilistMangaStatusToStatus(*mediaList.Status),
		TitleEN:          titleEN,
		TitleJP:          titleJP,
		TitleRomaji:      romajiTitle,
		Chapters:         chapters,
		Volumes:          volumes,
		StartedAt:        startedAt,
		FinishedAt:       finishedAt,
		reverseDirection: reverseDirection,
	}, nil
}

func newMangaFromMalManga(manga mal.Manga, reverseDirection bool) (Manga, error) {
	if manga.ID == 0 {
		return Manga{}, errors.New("ID is 0")
	}

	startedAt := parseDateOrNow(manga.MyListStatus.StartDate)
	finishedAt := parseDateOrNow(manga.MyListStatus.FinishDate)

	titleEN := manga.Title
	if manga.AlternativeTitles.En != "" {
		titleEN = manga.AlternativeTitles.En
	}

	titleJP := manga.Title
	if manga.AlternativeTitles.Ja != "" {
		titleJP = manga.AlternativeTitles.Ja
	}

	// In reverse sync mode, we need to leave AniList ID as 0 so the updater can find it by name
	anilistID := -1
	if reverseDirection {
		anilistID = 0 // This will trigger name-based search in reverse sync
	}

	return Manga{
		IDAnilist:        anilistID,
		IDMal:            manga.ID,
		Progress:         manga.MyListStatus.NumChaptersRead,
		ProgressVolumes:  manga.MyListStatus.NumVolumesRead,
		Score:            float64(manga.MyListStatus.Score),
		Status:           mapMalMangaStatusToStatus(manga.MyListStatus.Status),
		TitleEN:          titleEN,
		TitleJP:          titleJP,
		TitleRomaji:      "",
		Chapters:         manga.NumChapters,
		Volumes:          manga.NumVolumes,
		StartedAt:        startedAt,
		FinishedAt:       finishedAt,
		reverseDirection: reverseDirection,
	}, nil
}

func mapMalMangaStatusToStatus(s mal.MangaStatus) MangaStatus {
	switch s {
	case mal.MangaStatusReading:
		return MangaStatusReading
	case mal.MangaStatusCompleted:
		return MangaStatusCompleted
	case mal.MangaStatusOnHold:
		return MangaStatusOnHold
	case mal.MangaStatusDropped:
		return MangaStatusDropped
	case mal.MangaStatusPlanToRead:
		return MangaStatusPlanToRead
	default:
		return MangaStatusUnknown
	}
}

func mapAnilistMangaStatusToStatus(s verniy.MediaListStatus) MangaStatus {
	switch s {
	case verniy.MediaListStatusCurrent:
		return MangaStatusReading
	case verniy.MediaListStatusCompleted:
		return MangaStatusCompleted
	case verniy.MediaListStatusPaused:
		return MangaStatusOnHold
	case verniy.MediaListStatusDropped:
		return MangaStatusDropped
	case verniy.MediaListStatusPlanning:
		return MangaStatusPlanToRead
	case verniy.MediaListStatusRepeating:
		return MangaStatusRepeating
	default:
		return MangaStatusUnknown
	}
}

func newMangasFromMediaListGroups(groups []verniy.MediaListGroup, reverseDirection bool) []Manga {
	res := make([]Manga, 0, len(groups))
	for _, group := range groups {
		for _, mediaList := range group.Entries {
			r, err := newMangaFromMediaListEntry(mediaList, reverseDirection)
			if err != nil {
				log.Printf("Error creating manga from media list entry: %v", err)
				continue
			}

			res = append(res, r)
		}
	}
	return res
}

func newMangasFromMalUserMangas(mangas []mal.UserManga, reverseDirection bool) []Manga {
	res := make([]Manga, 0, len(mangas))
	for _, manga := range mangas {
		r, err := newMangaFromMalManga(manga.Manga, reverseDirection)
		if err != nil {
			log.Printf("Error creating manga from mal user manga: %v", err)
			continue
		}

		res = append(res, r)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].GetStatusString() < res[j].GetStatusString()
	})
	return res
}

func newMangasFromMalMangas(mangas []mal.Manga, reverseDirection bool) []Manga {
	res := make([]Manga, 0, len(mangas))
	for _, manga := range mangas {
		r, err := newMangaFromMalManga(manga, reverseDirection)
		if err != nil {
			log.Printf("Error creating manga from mal manga: %v", err)
			continue
		}

		res = append(res, r)
	}
	return res
}

func newTargetsFromMangas(mangas []Manga) []Target {
	res := make([]Target, 0, len(mangas))
	for _, manga := range mangas {
		res = append(res, manga)
	}
	return res
}

func newSourcesFromMangas(mangas []Manga) []Source {
	res := make([]Source, 0, len(mangas))
	for _, manga := range mangas {
		res = append(res, manga)
	}
	return res
}

func newMangasFromVerniyMedias(medias []verniy.Media, reverseDirection bool) []Manga {
	res := make([]Manga, 0, len(medias))
	for _, media := range medias {
		m, err := newMangaFromVerniyMedia(media, reverseDirection)
		if err != nil {
			log.Printf("failed to convert verniy media to manga: %v", err)
			continue
		}
		res = append(res, m)
	}
	return res
}

func newMangaFromVerniyMedia(media verniy.Media, reverseDirection bool) (Manga, error) {
	if media.ID == 0 {
		return Manga{}, errors.New("ID is 0")
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

	var chapters int
	if media.Chapters != nil {
		chapters = *media.Chapters
	}

	var volumes int
	if media.Volumes != nil {
		volumes = *media.Volumes
	}

	var idMal int
	if media.IDMAL != nil {
		idMal = *media.IDMAL
	}

	return Manga{
		IDAnilist:        media.ID,
		IDMal:            idMal,
		Progress:         0,                  // Will be set from MAL source
		ProgressVolumes:  0,                  // Will be set from MAL source
		Score:            0,                  // Will be set from MAL source
		Status:           MangaStatusUnknown, // Will be set from MAL source
		TitleEN:          titleEN,
		TitleJP:          titleJP,
		TitleRomaji:      romajiTitle,
		Chapters:         chapters,
		Volumes:          volumes,
		StartedAt:        nil, // Will be set from MAL source
		FinishedAt:       nil, // Will be set from MAL source
		reverseDirection: reverseDirection,
	}, nil
}
