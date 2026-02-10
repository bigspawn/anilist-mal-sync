package jikan

import (
	"context"
	"net/http"
)

type RandomService struct{ c *Client }

func (s *RandomService) Anime(ctx context.Context) (*Anime, error) {
	var r struct{ Data Anime }
	if err := s.c.Do(ctx, http.MethodGet, "/random/anime", nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *RandomService) Manga(ctx context.Context) (*Manga, error) {
	var r struct{ Data Manga }
	if err := s.c.Do(ctx, http.MethodGet, "/random/manga", nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *RandomService) Character(ctx context.Context) (*Character, error) {
	var r struct{ Data Character }
	if err := s.c.Do(ctx, http.MethodGet, "/random/characters", nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *RandomService) Person(ctx context.Context) (*Person, error) {
	var r struct{ Data Person }
	if err := s.c.Do(ctx, http.MethodGet, "/random/people", nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}
