package main

//go:generate mockgen -destination mock_service_test.go -package main -source=service.go

import (
	"context"
	"fmt"
	"time"

	"github.com/rl404/verniy"
)

// MediaService abstracts operations with a media service (MAL/AniList)
type MediaService interface {
	GetByID(ctx context.Context, id TargetID, prefix string) (Target, error)
	GetByName(ctx context.Context, name string, prefix string) ([]Target, error)
	Update(ctx context.Context, id TargetID, src Source, prefix string) error
}

// MediaServiceWithMALID extends MediaService with MAL ID lookup capability
type MediaServiceWithMALID interface {
	MediaService
	GetByMALID(ctx context.Context, malID int, prefix string) (Target, error)
}

// MALAnimeService implements MediaService for MAL anime operations
type MALAnimeService struct {
	client *MyAnimeListClient
}

// NewMALAnimeService creates a new MAL anime service
func NewMALAnimeService(client *MyAnimeListClient) *MALAnimeService {
	return &MALAnimeService{client: client}
}

func (s *MALAnimeService) GetByID(ctx context.Context, id TargetID, _ string) (Target, error) {
	resp, err := s.client.GetAnimeByID(ctx, int(id))
	if err != nil {
		return nil, fmt.Errorf("error getting anime by id: %w", err)
	}
	// false: MALAnimeService is only used in forward sync (AniList→MAL)
	ani, err := newAnimeFromMalAnime(*resp, false)
	if err != nil {
		return nil, fmt.Errorf("error creating anime from mal anime: %w", err)
	}
	return ani, nil
}

func (s *MALAnimeService) GetByName(ctx context.Context, name string, _ string) ([]Target, error) {
	resp, err := s.client.GetAnimesByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error getting anime by name: %w", err)
	}
	// false: search results are forward-sync targets, not reverse-sync sources
	return newTargetsFromAnimes(newAnimesFromMalAnimes(resp, false)), nil
}

func (s *MALAnimeService) Update(ctx context.Context, id TargetID, src Source, _ string) error {
	a, ok := src.(Anime)
	if !ok {
		return fmt.Errorf("source is not an anime")
	}
	if err := s.client.UpdateAnimeByIDAndOptions(ctx, int(id), a.GetUpdateOptions()); err != nil {
		return fmt.Errorf("error updating anime by id and options: %w", err)
	}
	return nil
}

// MALMangaService implements MediaService for MAL manga operations
type MALMangaService struct {
	client *MyAnimeListClient
}

// NewMALMangaService creates a new MAL manga service
func NewMALMangaService(client *MyAnimeListClient) *MALMangaService {
	return &MALMangaService{client: client}
}

func (s *MALMangaService) GetByID(ctx context.Context, id TargetID, _ string) (Target, error) {
	resp, err := s.client.GetMangaByID(ctx, int(id))
	if err != nil {
		return nil, fmt.Errorf("error getting manga by id: %w", err)
	}
	// false: MALMangaService is only used in forward sync (AniList→MAL)
	manga, err := newMangaFromMalManga(*resp, false)
	if err != nil {
		return nil, fmt.Errorf("error creating manga from mal manga: %w", err)
	}
	return manga, nil
}

func (s *MALMangaService) GetByName(ctx context.Context, name string, _ string) ([]Target, error) {
	resp, err := s.client.GetMangasByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error getting manga by name: %w", err)
	}
	// false: search results are forward-sync targets, not reverse-sync sources
	return newTargetsFromMangas(newMangasFromMalMangas(resp, false)), nil
}

func (s *MALMangaService) Update(ctx context.Context, id TargetID, src Source, _ string) error {
	m, ok := src.(Manga)
	if !ok {
		return fmt.Errorf("source is not a manga")
	}
	if err := s.client.UpdateMangaByIDAndOptions(ctx, int(id), m.GetUpdateOptions()); err != nil {
		return fmt.Errorf("error updating manga by id and options: %w", err)
	}
	return nil
}

// AniListAnimeService implements MediaServiceWithMALID for AniList anime operations.
// Reverse=true marks returned entries as reverse-sync targets (IDAnilist as target ID).
type AniListAnimeService struct {
	client      *AnilistClient
	scoreFormat verniy.ScoreFormat
	Reverse     bool
}

// NewAniListAnimeService creates a new AniList anime service.
// reverse=true for MAL→AniList direction.
func NewAniListAnimeService(client *AnilistClient, scoreFormat verniy.ScoreFormat, reverse bool) *AniListAnimeService {
	return &AniListAnimeService{client: client, scoreFormat: scoreFormat, Reverse: reverse}
}

func (s *AniListAnimeService) GetByID(ctx context.Context, id TargetID, _ string) (Target, error) {
	resp, err := s.client.GetAnimeByID(ctx, int(id))
	if err != nil {
		return nil, fmt.Errorf("error getting anilist anime by id: %w", err)
	}
	ani, err := newAnimeFromVerniyMedia(*resp, s.Reverse)
	if err != nil {
		return nil, fmt.Errorf("error creating anime from anilist media: %w", err)
	}
	return ani, nil
}

func (s *AniListAnimeService) GetByName(ctx context.Context, name string, _ string) ([]Target, error) {
	resp, err := s.client.GetAnimesByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error getting anilist anime by name: %w", err)
	}
	return newTargetsFromAnimes(newAnimesFromVerniyMedias(resp, s.Reverse)), nil
}

func (s *AniListAnimeService) GetByMALID(ctx context.Context, malID int, _ string) (Target, error) {
	resp, err := s.client.GetAnimeByMALID(ctx, malID)
	if err != nil {
		return nil, fmt.Errorf("error getting anilist anime by MAL ID: %w", err)
	}
	ani, err := newAnimeFromVerniyMedia(*resp, s.Reverse)
	if err != nil {
		return nil, fmt.Errorf("error creating anime from anilist media: %w", err)
	}
	return ani, nil
}

func (s *AniListAnimeService) Update(ctx context.Context, id TargetID, src Source, _ string) error {
	a, ok := src.(Anime)
	if !ok {
		return fmt.Errorf("source is not an anime")
	}
	anilistScore := denormalizeScoreForAniList(ctx, a.Score, s.scoreFormat)

	var completedAt *time.Time
	if a.Status == StatusCompleted {
		completedAt = a.FinishedAt
	}

	if err := s.client.UpdateAnimeEntry(
		ctx,
		int(id),
		a.Status.GetAnilistStatus(),
		a.Progress,
		anilistScore,
		a.StartedAt,
		completedAt,
	); err != nil {
		return fmt.Errorf("error updating anilist anime: %w", err)
	}
	return nil
}

// AniListMangaService implements MediaServiceWithMALID for AniList manga operations.
// Reverse=true marks returned entries as reverse-sync targets (IDAnilist as target ID).
type AniListMangaService struct {
	client      *AnilistClient
	scoreFormat verniy.ScoreFormat
	Reverse     bool
}

// NewAniListMangaService creates a new AniList manga service.
// reverse=true for MAL→AniList direction.
func NewAniListMangaService(client *AnilistClient, scoreFormat verniy.ScoreFormat, reverse bool) *AniListMangaService {
	return &AniListMangaService{client: client, scoreFormat: scoreFormat, Reverse: reverse}
}

func (s *AniListMangaService) GetByID(ctx context.Context, id TargetID, _ string) (Target, error) {
	resp, err := s.client.GetMangaByID(ctx, int(id))
	if err != nil {
		return nil, fmt.Errorf("error getting anilist manga by id: %w", err)
	}
	manga, err := newMangaFromVerniyMedia(*resp, s.Reverse)
	if err != nil {
		return nil, fmt.Errorf("error creating manga from anilist media: %w", err)
	}
	return manga, nil
}

func (s *AniListMangaService) GetByName(ctx context.Context, name string, _ string) ([]Target, error) {
	resp, err := s.client.GetMangasByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error getting anilist manga by name: %w", err)
	}
	return newTargetsFromMangas(newMangasFromVerniyMedias(resp, s.Reverse)), nil
}

func (s *AniListMangaService) GetByMALID(ctx context.Context, malID int, _ string) (Target, error) {
	resp, err := s.client.GetMangaByMALID(ctx, malID)
	if err != nil {
		return nil, fmt.Errorf("error getting anilist manga by MAL ID: %w", err)
	}
	manga, err := newMangaFromVerniyMedia(*resp, s.Reverse)
	if err != nil {
		return nil, fmt.Errorf("error creating manga from anilist media: %w", err)
	}
	return manga, nil
}

func (s *AniListMangaService) Update(ctx context.Context, id TargetID, src Source, _ string) error {
	m, ok := src.(Manga)
	if !ok {
		return fmt.Errorf("source is not a manga")
	}
	anilistScore := denormalizeMangaScoreForAniList(ctx, m.Score, s.scoreFormat)

	var completedAt *time.Time
	if m.Status == MangaStatusCompleted {
		completedAt = m.FinishedAt
	}

	if err := s.client.UpdateMangaEntry(
		ctx,
		int(id),
		m.Status.GetAnilistStatus(),
		m.Progress,
		m.ProgressVolumes,
		anilistScore,
		m.StartedAt,
		completedAt,
	); err != nil {
		return fmt.Errorf("error updating anilist manga: %w", err)
	}
	return nil
}
