package jikan

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type SearchService struct{ c *Client }

func (s *SearchService) Anime(ctx context.Context, query string, opts struct {
	Type    string
	Status  string
	Rating  string
	Genres  []int
	OrderBy string
	Sort    string
	Page    int
	Limit   int
},
) ([]Anime, *Pagination, error) {
	q := url.Values{}
	if query != "" {
		q.Set("q", query)
	}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Rating != "" {
		q.Set("rating", opts.Rating)
	}
	if len(opts.Genres) > 0 {
		q.Set("genres", joinInts(opts.Genres, ","))
	}
	if opts.OrderBy != "" {
		q.Set("order_by", opts.OrderBy)
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}

	var r struct {
		Data       []Anime    `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/anime", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
