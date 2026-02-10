// magazine.go
package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type MagazineService struct {
	c *Client
}

type Magazine struct {
	MalID ID     `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

func (s *MagazineService) ByID(ctx context.Context, id ID) (*Magazine, error) {
	var r struct{ Data Magazine }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/magazines/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *MagazineService) Search(ctx context.Context, query string, page int) ([]Magazine, *Pagination, error) {
	q := url.Values{"q": {query}, "page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Magazine `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/magazines", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
