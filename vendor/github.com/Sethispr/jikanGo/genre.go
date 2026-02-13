package jikan

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type GenreService struct {
	c *Client
}

type Genre struct {
	MalID ID     `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

type GenreFilter string

const (
	GenreGenres       GenreFilter = "genres"
	GenreExplicit     GenreFilter = "explicit_genres"
	GenreThemes       GenreFilter = "themes"
	GenreDemographics GenreFilter = "demographics"
)

func (s *GenreService) Anime(ctx context.Context, filter GenreFilter, page, limit int) ([]*Genre, *Pagination, error) {
	return s.list(ctx, "/genres/anime", filter, page, limit)
}

func (s *GenreService) Manga(ctx context.Context, filter GenreFilter, page, limit int) ([]*Genre, *Pagination, error) {
	return s.list(ctx, "/genres/manga", filter, page, limit)
}

func (s *GenreService) list(ctx context.Context, endpoint string, filter GenreFilter, page, limit int) ([]*Genre, *Pagination, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", string(filter))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	var r struct {
		Data       []*Genre   `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, endpoint, q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
