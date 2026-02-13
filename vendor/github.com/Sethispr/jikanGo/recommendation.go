// recommendation.go
package jikan

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type RecommendationService struct {
	c *Client
}

type Recommendation struct {
	MalID   string     `json:"mal_id"`
	Entry   []Resource `json:"entry"`
	Content string     `json:"content"`
	Date    string     `json:"date"`
	User    struct {
		Username string `json:"username"`
	} `json:"user"`
}

func (s *RecommendationService) Anime(ctx context.Context, page int) ([]Recommendation, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Recommendation `json:"data"`
		Pagination Pagination       `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/recommendations/anime", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *RecommendationService) Manga(ctx context.Context, page int) ([]Recommendation, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Recommendation `json:"data"`
		Pagination Pagination       `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/recommendations/manga", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
