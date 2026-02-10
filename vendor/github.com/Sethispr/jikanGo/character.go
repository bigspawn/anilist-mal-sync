package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type CharacterService struct {
	c *Client
}

type Character struct {
	MalID     ID       `json:"mal_id"`
	URL       string   `json:"url"`
	Images    ImageSet `json:"images"`
	Name      string   `json:"name"`
	NameKanji string   `json:"name_kanji"`
	Nicknames []string `json:"nicknames"`
	About     string   `json:"about"`
	Favorites int      `json:"favorites"`
}

type Entry struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Title  string   `json:"title"`
}

type CharacterAnime struct {
	Role  string `json:"role"`
	Anime Entry  `json:"anime"`
}

type CharacterManga struct {
	Role  string `json:"role"`
	Manga Entry  `json:"manga"`
}

type VoiceActor struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Name   string   `json:"name"`
}

type CharacterVoice struct {
	Language string     `json:"language"`
	Person   VoiceActor `json:"person"`
}

type CharacterPicture struct {
	ImageURL      string `json:"image_url"`
	LargeImageURL string `json:"large_image_url"`
}

type CharacterFull struct {
	Character `json:",inline"`
	Anime     []CharacterAnime `json:"anime"`
	Manga     []CharacterManga `json:"manga"`
	Voices    []CharacterVoice `json:"voices"`
}

func (s *CharacterService) validateID(id ID) error {
	if id < 1 {
		return fmt.Errorf("invalid character id: %d", id)
	}
	return nil
}

func (s *CharacterService) ByID(ctx context.Context, id ID) (*Character, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data Character `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *CharacterService) Full(ctx context.Context, id ID) (*CharacterFull, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data CharacterFull `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/full", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *CharacterService) Anime(ctx context.Context, id ID) ([]CharacterAnime, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data []CharacterAnime `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/anime", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Manga(ctx context.Context, id ID) ([]CharacterManga, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data []CharacterManga `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/manga", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Voices(ctx context.Context, id ID) ([]CharacterVoice, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data []CharacterVoice `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/voices", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Pictures(ctx context.Context, id ID) ([]CharacterPicture, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	var r struct {
		Data []CharacterPicture `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/pictures", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Search(ctx context.Context, query string, page int) ([]*Character, *Pagination, error) {
	q := url.Values{}
	if query != "" {
		q.Set("q", query)
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var r struct {
		Data       []*Character `json:"data"`
		Pagination Pagination   `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/characters", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
