// watch.go
package jikan

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type WatchService struct {
	c *Client
}

type EpisodePreview struct {
	Entry    Resource `json:"entry"`
	Episodes []struct {
		MalID   ID     `json:"mal_id"`
		Title   string `json:"title"`
		Premium bool   `json:"premium"`
	} `json:"episodes"`
	RegionLocked bool `json:"region_locked"`
}

func (s *WatchService) Episodes(ctx context.Context, page int) ([]EpisodePreview, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []EpisodePreview `json:"data"`
		Pagination Pagination       `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/watch/episodes", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

type Promo struct {
	Title   string `json:"title"`
	Trailer struct {
		YoutubeID string   `json:"youtube_id"`
		EmbedURL  string   `json:"embed_url"`
		Images    ImageSet `json:"images"`
	} `json:"trailer"`
	Entry Resource `json:"entry"`
}

func (s *WatchService) Promos(ctx context.Context, page int) ([]Promo, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Promo    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/watch/promos", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
