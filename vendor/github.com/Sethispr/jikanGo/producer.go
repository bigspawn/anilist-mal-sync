// producer.go
package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ProducerService struct {
	c *Client
}

type Producer struct {
	MalID       ID       `json:"mal_id"`
	URL         string   `json:"url"`
	Images      ImageSet `json:"images"`
	Title       string   `json:"title"`
	Established string   `json:"established"`
	About       string   `json:"about"`
	Count       int      `json:"count"`
}

func (s *ProducerService) ByID(ctx context.Context, id ID) (*Producer, error) {
	var r struct{ Data Producer }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/producers/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *ProducerService) Search(ctx context.Context, query string, page int) ([]Producer, *Pagination, error) {
	q := url.Values{"q": {query}, "page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Producer `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/producers", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
