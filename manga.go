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
	case MangaStatusUnknown:
		return "", errMangaStatusUnknown
	default:
		return "", errMangaStatusUnknown
	}
}

func (s MangaStatus) GetAnilistStatus() string {
	switch s {
	case MangaStatusReading:
		return "CURRENT"
	case MangaStatusCompleted:
		return "COMPLETED"
	case MangaStatusOnHold:
		return "PAUSED"
	case MangaStatusDropped:
		return "DROPPED"
	case MangaStatusPlanToRead:
		return "PLANNING"
	case MangaStatusUnknown:
		return ""
	default:
		return ""
	}
}

type Manga struct {
	IDAnilist       int
	IDMal           int
	Progress        int
	ProgressVolumes int
	Score           int
	Status          MangaStatus
	TitleEN         string
	TitleJP         string
	TitleRomaji     string
	Chapters        int
	Volumes         int
	StartedAt       *time.Time
	FinishedAt      *time.Time
}

func (m Manga) GetTargetID() TargetID {
	if *reverseDirection {
		return TargetID(m.IDAnilist)
	}
	return TargetID(m.IDMal)
}

func (m Manga) GetSourceID() int {
	if *reverseDirection {
		return m.IDMal
	}
	return m.IDAnilist
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
		DPrintf("Status: %s != %s", m.Status, b.Status)
		return false
	}
	if m.Score != b.Score {
		DPrintf("Score: %d != %d", m.Score, b.Score)
		return false
	}
	if m.Progress != b.Progress {
		DPrintf("Progress: %d != %d", m.Progress, b.Progress)
		return false
	}
	if m.ProgressVolumes != b.ProgressVolumes {
		DPrintf("ProgressVolumes: %d != %d", m.ProgressVolumes, b.ProgressVolumes)
		return false
	}

	return true
}

func (m Manga) SameTypeWithTarget(t Target) bool {
	b, ok := t.(Manga)
	if !ok {
		return false
	}

	// Check if MAL IDs match (critical for reverse sync)
	if m.IDMal > 0 && b.IDMal > 0 && m.IDMal == b.IDMal {
		return true
	}

	// Check if AniList IDs match
	if m.IDAnilist > 0 && b.IDAnilist > 0 && m.IDAnilist == b.IDAnilist {
		return true
	}

	// Use the comprehensive title matching logic
	if m.SameTitleWithTarget(t) {
		return true
	}

	// Fallback: Check if chapters and volumes match (only if both are known)
	if (m.Chapters > 0 || m.Volumes > 0) && m.Chapters == b.Chapters && m.Volumes == b.Volumes {
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

func (m Manga) GetUpdateMyAnimeListStatusOption() []mal.UpdateMyAnimeListStatusOption {
	return nil
}

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
	sb.WriteString(fmt.Sprintf("IDAnilist: %d, ", m.IDAnilist))
	sb.WriteString(fmt.Sprintf("IDMal: %d, ", m.IDMal))
	sb.WriteString(fmt.Sprintf("TitleEN: %s, ", m.TitleEN))
	sb.WriteString(fmt.Sprintf("TitleJP: %s, ", m.TitleJP))
	sb.WriteString(fmt.Sprintf("Status: %s, ", m.Status))
	sb.WriteString(fmt.Sprintf("Score: %d, ", m.Score))
	sb.WriteString(fmt.Sprintf("Progress: %d, ", m.Progress))
	sb.WriteString(fmt.Sprintf("ProgressVolumes: %d, ", m.ProgressVolumes))
	sb.WriteString(fmt.Sprintf("Chapters: %d, ", m.Chapters))
	sb.WriteString(fmt.Sprintf("Volumes: %d, ", m.Volumes))
	sb.WriteString(fmt.Sprintf("StartedAt: %s, ", m.StartedAt))
	sb.WriteString(fmt.Sprintf("FinishedAt: %s", m.FinishedAt))
	sb.WriteString("}")
	return sb.String()
}

func (m Manga) GetUpdateOptions() []mal.UpdateMyMangaListStatusOption {
	st, err := m.Status.GetMalStatus()
	if err != nil {
		log.Printf("Error getting MAL status: %v", err)
		return nil
	}

	opts := []mal.UpdateMyMangaListStatusOption{
		st,
		mal.Score(m.Score),
		mal.NumChaptersRead(m.Progress),
		mal.NumVolumesRead(m.ProgressVolumes),
	}

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

func newMangaFromMediaListEntry(mediaList verniy.MediaList, scoreFormat verniy.ScoreFormat) (Manga, error) {
	if mediaList.Media == nil {
		return Manga{}, errors.New("media is nil")
	}

	if mediaList.Status == nil {
		return Manga{}, errors.New("status is nil")
	}

	if mediaList.Media.Title == nil {
		return Manga{}, errors.New("title is nil")
	}

	var score int
	if mediaList.Score != nil {
		// Normalize AniList score to MAL format (0-10)
		score = normalizeMangaScoreForMAL(*mediaList.Score, scoreFormat)
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
		IDAnilist:       mediaList.Media.ID,
		IDMal:           idMal,
		Progress:        progress,
		ProgressVolumes: progressVolumes,
		Score:           score,
		Status:          mapAnilistMangaStatustToStatus(*mediaList.Status),
		TitleEN:         titleEN,
		TitleJP:         titleJP,
		TitleRomaji:     romajiTitle,
		Chapters:        chapters,
		Volumes:         volumes,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
	}, nil
}

func newMangaFromMalManga(manga mal.Manga) (Manga, error) {
	if manga.ID == 0 {
		return Manga{}, errors.New("ID is nil")
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
	if *reverseDirection {
		anilistID = 0 // This will trigger name-based search in reverse sync
	}

	return Manga{
		IDAnilist:       anilistID,
		IDMal:           manga.ID,
		Progress:        manga.MyListStatus.NumChaptersRead,
		ProgressVolumes: manga.MyListStatus.NumVolumesRead,
		Score:           manga.MyListStatus.Score, // MAL score is already 0-10 int
		Status:          mapMalMangaStatusToStatus(manga.MyListStatus.Status),
		TitleEN:         titleEN,
		TitleJP:         titleJP,
		TitleRomaji:     "",
		Chapters:        manga.NumChapters,
		Volumes:         manga.NumVolumes,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
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

func mapAnilistMangaStatustToStatus(s verniy.MediaListStatus) MangaStatus {
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
		return MangaStatusReading // TODO: handle repeating correctly
	default:
		return MangaStatusUnknown
	}
}

func newMangasFromMediaListGroups(groups []verniy.MediaListGroup, scoreFormat verniy.ScoreFormat) []Manga {
	res := make([]Manga, 0, len(groups))
	for _, group := range groups {
		for _, mediaList := range group.Entries {
			r, err := newMangaFromMediaListEntry(mediaList, scoreFormat)
			if err != nil {
				log.Printf("Error creating manga from media list entry: %v", err)
				continue
			}

			res = append(res, r)
		}
	}
	return res
}

func newMangasFromMalUserMangas(mangas []mal.UserManga) []Manga {
	res := make([]Manga, 0, len(mangas))
	for _, manga := range mangas {
		r, err := newMangaFromMalManga(manga.Manga)
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

func newMangasFromMalMangas(mangas []mal.Manga) []Manga {
	res := make([]Manga, 0, len(mangas))
	for _, manga := range mangas {
		r, err := newMangaFromMalManga(manga)
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

func newMangasFromVerniyMedias(medias []verniy.Media) []Manga {
	res := make([]Manga, 0, len(medias))
	for _, media := range medias {
		m, err := newMangaFromVerniyMedia(media)
		if err != nil {
			log.Printf("failed to convert verniy media to manga: %v", err)
			continue
		}
		res = append(res, m)
	}
	return res
}

func newMangaFromVerniyMedia(media verniy.Media) (Manga, error) {
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
		IDAnilist:       media.ID,
		IDMal:           idMal,
		Progress:        0,                  // Will be set from MAL source
		ProgressVolumes: 0,                  // Will be set from MAL source
		Score:           0,                  // Will be set from MAL source
		Status:          MangaStatusUnknown, // Will be set from MAL source
		TitleEN:         titleEN,
		TitleJP:         titleJP,
		TitleRomaji:     romajiTitle,
		Chapters:        chapters,
		Volumes:         volumes,
		StartedAt:       nil, // Will be set from MAL source
		FinishedAt:      nil, // Will be set from MAL source
	}, nil
}
