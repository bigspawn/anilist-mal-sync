package jikan

import (
	"context"
	"fmt"
	"net/http"
)

type PeopleService struct{ c *Client }

type Person struct {
	MalID          ID       `json:"mal_id"`
	URL            string   `json:"url"`
	Images         ImageSet `json:"images"`
	Name           string   `json:"name"`
	GivenName      string   `json:"given_name"`
	FamilyName     string   `json:"family_name"`
	AlternateNames []string `json:"alternate_names"`
	Birthday       string   `json:"birthday"`
	Favorites      int      `json:"favorites"`
	About          string   `json:"about"`
	VoiceRoles     []struct {
		Role      string   `json:"role"`
		Anime     Resource `json:"anime"`
		Character Resource `json:"character"`
	} `json:"voices"`
}

func (s *PeopleService) ByID(ctx context.Context, id ID) (*Person, error) {
	var r struct{ Data Person }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/people/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}
