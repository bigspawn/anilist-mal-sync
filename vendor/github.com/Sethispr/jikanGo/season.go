package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type SeasonService struct{ c *Client }

func (s *SeasonService) Now(ctx context.Context, page int) ([]Anime, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Anime    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/seasons/now", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *SeasonService) Archive(ctx context.Context, year int, season string, page int) ([]Anime, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Anime    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	path := fmt.Sprintf("/seasons/%d/%s", year, season)
	if err := s.c.Do(ctx, http.MethodGet, path, q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *SeasonService) Upcoming(ctx context.Context, page int) ([]Anime, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Anime    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/seasons/upcoming", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
