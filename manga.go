package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/rl404/verniy"
)

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
	default:
		return "", errors.New("unknown status")
	}
}

type Manga struct {
	IDAnilist       int
	IDMal           int
	Progress        int
	ProgressVolumes int
	Score           float64
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
		DPrintf("Status: %s != %s", m.Status, b.Status)
		return false
	}
	if m.Score != b.Score {
		DPrintf("Score: %f != %f", m.Score, b.Score)
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

	if m.IDMal == b.IDMal || m.IDAnilist == b.IDAnilist {
		return true
	}

	if m.TitleEN != "" && b.TitleEN != "" && strings.EqualFold(m.TitleEN, b.TitleEN) {
		return true
	}

	if m.TitleJP != "" && b.TitleJP != "" && strings.EqualFold(m.TitleJP, b.TitleJP) {
		return true
	}

	if m.TitleRomaji != "" && b.TitleRomaji != "" && strings.EqualFold(m.TitleRomaji, b.TitleRomaji) {
		return true
	}

	if m.Chapters == b.Chapters && m.Volumes == b.Volumes {
		// NOTE: some mangas are joined in MAL in the same entry in Volumes, but it is separated in Anilist.
		// Skip it for now.
		return true
	}

	return false
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
	sb.WriteString(fmt.Sprintf("Score: %f, ", m.Score))
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

func newMangaFromMediaListEntry(mediaList verniy.MediaList) (Manga, error) {
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

	return Manga{
		IDAnilist:       -1,
		IDMal:           manga.ID,
		Progress:        manga.MyListStatus.NumChaptersRead,
		ProgressVolumes: manga.MyListStatus.NumVolumesRead,
		Score:           float64(manga.MyListStatus.Score),
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

func newMangasFromMediaListGroups(groups []verniy.MediaListGroup) []Manga {
	res := make([]Manga, 0, len(groups))
	for _, group := range groups {
		for _, mediaList := range group.Entries {
			r, err := newMangaFromMediaListEntry(mediaList)
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
