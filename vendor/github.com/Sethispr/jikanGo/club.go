package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ClubMember struct {
	Username string `json:"username"`
	URL      string `json:"url"`
	Images   struct {
		JPG struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
	} `json:"images"`
	LastOnline string `json:"last_online"`
}

type ClubStaff struct {
	URL      string `json:"url"`
	Username string `json:"username"`
}

type ClubRelations struct {
	Anime      []Resource `json:"anime"`
	Manga      []Resource `json:"manga"`
	Characters []Resource `json:"characters"`
}

type ClubService struct {
	c *Client
}

type Club struct {
	MalID  int    `json:"mal_id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Images struct {
		JPG struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
		WebP struct {
			ImageURL string `json:"image_url"`
		} `json:"webp"`
	} `json:"images"`
	Members  int    `json:"members"`
	Category string `json:"category"`
	Access   string `json:"access"`
	Created  string `json:"created"`
}

// ByID gets a club by their MAL ID
func (s *ClubService) ByID(ctx context.Context, id ID) (*Club, error) {
	var r struct {
		Data Club `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *ClubService) Search(ctx context.Context, query string, page int) ([]Club, *Pagination, error) {
	q := url.Values{
		"q":    {query},
		"page": {strconv.Itoa(page)},
	}
	var r struct {
		Data       []Club     `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/clubs", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *ClubService) Members(ctx context.Context, id ID, page int) ([]ClubMember, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []ClubMember `json:"data"`
		Pagination Pagination   `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d/members", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *ClubService) Staff(ctx context.Context, id ID) ([]ClubStaff, error) {
	var r struct {
		Data []ClubStaff `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d/staff", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *ClubService) Relations(ctx context.Context, id ID) (*ClubRelations, error) {
	var r struct {
		Data ClubRelations `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d/relations", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}
