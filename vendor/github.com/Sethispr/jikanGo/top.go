// top.go
package jikan

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type TopService struct {
	c *Client
}

func (s *TopService) Anime(ctx context.Context, filter string, page int) ([]Anime, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	if filter != "" {
		q.Set("type", filter)
	}
	var r struct {
		Data       []Anime    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/top/anime", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *TopService) Manga(ctx context.Context, filter string, page int) ([]Manga, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	if filter != "" {
		q.Set("type", filter)
	}
	var r struct {
		Data       []Manga    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/top/manga", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *TopService) People(ctx context.Context, page int) ([]struct {
	MalID     ID       `json:"mal_id"`
	Name      string   `json:"name"`
	Images    ImageSet `json:"images"`
	Favorites int      `json:"favorites"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID     ID       `json:"mal_id"`
			Name      string   `json:"name"`
			Images    ImageSet `json:"images"`
			Favorites int      `json:"favorites"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/top/people", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *TopService) Characters(ctx context.Context, page int) ([]struct {
	MalID     ID       `json:"mal_id"`
	Name      string   `json:"name"`
	Images    ImageSet `json:"images"`
	Favorites int      `json:"favorites"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID     ID       `json:"mal_id"`
			Name      string   `json:"name"`
			Images    ImageSet `json:"images"`
			Favorites int      `json:"favorites"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/top/characters", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
