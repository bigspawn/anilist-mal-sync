// review.go
package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ReviewService struct {
	c *Client
}

type Review struct {
	MalID     int `json:"mal_id"`
	Score     int `json:"score"`
	Reactions struct {
		Overall int `json:"overall"`
	} `json:"reactions"`
	Date      string   `json:"date"`
	Review    string   `json:"review"`
	IsSpoiler bool     `json:"is_spoiler"`
	Entry     Resource `json:"entry"`
	User      struct {
		Username string `json:"username"`
	} `json:"user"`
}

func (s *ReviewService) Recent(ctx context.Context, page int) ([]Review, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Review   `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/reviews/recent", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *ReviewService) ForAnime(ctx context.Context, id ID, page int) ([]Review, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Review   `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/reviews", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *ReviewService) ForManga(ctx context.Context, id ID, page int) ([]Review, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Review   `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/manga/%d/reviews", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
