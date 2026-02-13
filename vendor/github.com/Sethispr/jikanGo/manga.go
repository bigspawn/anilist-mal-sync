package jikan

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type MangaService struct {
	c *Client
}

func (s *MangaService) validateID(id ID) error {
	if id < 1 {
		return fmt.Errorf("invalid manga id: %d", id)
	}
	return nil
}

type Manga struct {
	MalID          ID         `json:"mal_id"`
	URL            string     `json:"url"`
	Images         ImageSet   `json:"images"`
	Title          string     `json:"title"`
	TitleEnglish   string     `json:"title_english"`
	TitleJapanese  string     `json:"title_japanese"`
	TitleSynonyms  []string   `json:"title_synonyms"`
	Type           string     `json:"type"`
	Chapters       *int       `json:"chapters"`
	Volumes        *int       `json:"volumes"`
	Status         string     `json:"status"`
	Publishing     bool       `json:"publishing"`
	Published      DateRange  `json:"published"`
	Score          *float64   `json:"score"`
	ScoredBy       *int       `json:"scored_by"`
	Rank           *int       `json:"rank"`
	Popularity     *int       `json:"popularity"`
	Members        *int       `json:"members"`
	Favorites      *int       `json:"favorites"`
	Synopsis       string     `json:"synopsis"`
	Background     string     `json:"background"`
	Authors        []Resource `json:"authors"`
	Serializations []Resource `json:"serializations"`
	Genres         []Resource `json:"genres"`
	ExplicitGenres []Resource `json:"explicit_genres"`
	Themes         []Resource `json:"themes"`
	Demographics   []Resource `json:"demographics"`
}

type MangaFull struct {
	Manga
	Relations []MangaRelation `json:"relations"`
	External  []ExternalLink  `json:"external"`
}

type MangaCharacter struct {
	Character struct {
		MalID  ID       `json:"mal_id"`
		URL    string   `json:"url"`
		Images ImageSet `json:"images"`
		Name   string   `json:"name"`
	} `json:"character"`
	Role string `json:"role"`
}

type MangaNews struct {
	MalID          int       `json:"mal_id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Date           time.Time `json:"date"`
	AuthorUsername string    `json:"author_username"`
	AuthorURL      string    `json:"author_url"`
	ForumURL       string    `json:"forum_url"`
	Images         ImageSet  `json:"images"`
	Comments       int       `json:"comments"`
	Excerpt        string    `json:"excerpt"`
}

type MangaTopic struct {
	MalID          int       `json:"mal_id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Date           time.Time `json:"date"`
	AuthorUsername string    `json:"author_username"`
	AuthorURL      string    `json:"author_url"`
	Comments       int       `json:"comments"`
	LastComment    struct {
		URL            string    `json:"url"`
		AuthorUsername string    `json:"author_username"`
		AuthorURL      string    `json:"author_url"`
		Date           time.Time `json:"date"`
	} `json:"last_comment"`
}

type MangaStats struct {
	Reading    int          `json:"reading"`
	Completed  int          `json:"completed"`
	OnHold     int          `json:"on_hold"`
	Dropped    int          `json:"dropped"`
	PlanToRead int          `json:"plan_to_read"`
	Total      int          `json:"total"`
	Scores     []ScoreStats `json:"scores"`
}

type ScoreStats struct {
	Score      int     `json:"score"`
	Votes      int     `json:"votes"`
	Percentage float64 `json:"percentage"`
}

type MangaRecommendation struct {
	Entry struct {
		MalID  ID       `json:"mal_id"`
		URL    string   `json:"url"`
		Images ImageSet `json:"images"`
		Title  string   `json:"title"`
	} `json:"entry"`
	URL   string `json:"url"`
	Votes int    `json:"votes"`
}

type MangaUserUpdate struct {
	User struct {
		Username string   `json:"username"`
		URL      string   `json:"url"`
		Images   ImageSet `json:"images"`
	} `json:"user"`
	Score         *int      `json:"score"`
	Status        string    `json:"status"`
	ChaptersRead  int       `json:"chapters_read"`
	ChaptersTotal *int      `json:"chapters_total"`
	VolumesRead   int       `json:"volumes_read"`
	VolumesTotal  *int      `json:"volumes_total"`
	Date          time.Time `json:"date"`
}

type MangaReview struct {
	MalID         int             `json:"mal_id"`
	URL           string          `json:"url"`
	Type          string          `json:"type"`
	Reactions     ReviewReactions `json:"reactions"`
	Date          time.Time       `json:"date"`
	Review        string          `json:"review"`
	Score         int             `json:"score"`
	Tags          []string        `json:"tags"`
	IsSpoiler     bool            `json:"is_spoiler"`
	IsPreliminary bool            `json:"is_preliminary"`
	ChaptersRead  int             `json:"chapters_read"`
	User          struct {
		Username string   `json:"username"`
		URL      string   `json:"url"`
		Images   ImageSet `json:"images"`
	} `json:"user"`
}

type ReviewReactions struct {
	Overall     int `json:"overall"`
	Nice        int `json:"nice"`
	LoveIt      int `json:"love_it"`
	Funny       int `json:"funny"`
	Confusing   int `json:"confusing"`
	Informative int `json:"informative"`
	WellWritten int `json:"well_written"`
	Creative    int `json:"creative"`
}

type MangaRelation struct {
	Relation string     `json:"relation"`
	Entry    []Resource `json:"entry"`
}

type ExternalLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MangaSearchOptions struct {
	Query   string
	Type    string
	Status  string
	OrderBy string
	Sort    string
	Page    int
}

func (o MangaSearchOptions) ToValues() url.Values {
	v := url.Values{}
	if o.Query != "" {
		v.Set("q", o.Query)
	}
	if o.Type != "" {
		v.Set("type", o.Type)
	}
	if o.Status != "" {
		v.Set("status", o.Status)
	}
	if o.OrderBy != "" {
		v.Set("order_by", o.OrderBy)
	}
	if o.Sort != "" {
		v.Set("sort", o.Sort)
	}
	if o.Page > 0 {
		v.Set("page", strconv.Itoa(o.Page))
	}
	return v
}

func (s *MangaService) ByID(ctx context.Context, id ID) (*Manga, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	r, err := fetch[Manga](ctx, s.c, fmt.Sprintf("/manga/%d", id), nil)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *MangaService) Full(ctx context.Context, id ID) (*MangaFull, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	r, err := fetch[MangaFull](ctx, s.c, fmt.Sprintf("/manga/%d/full", id), nil)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *MangaService) Characters(ctx context.Context, id ID) ([]MangaCharacter, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	return fetch[[]MangaCharacter](ctx, s.c, fmt.Sprintf("/manga/%d/characters", id), nil)
}

func (s *MangaService) News(ctx context.Context, id ID, page int) ([]MangaNews, *Pagination, error) {
	if err := s.validateID(id); err != nil {
		return nil, nil, err
	}
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	return fetchPaged[[]MangaNews](ctx, s.c, fmt.Sprintf("/manga/%d/news", id), q)
}

func (s *MangaService) Forum(ctx context.Context, id ID, filter string) ([]MangaTopic, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	return fetch[[]MangaTopic](ctx, s.c, fmt.Sprintf("/manga/%d/forum", id), q)
}

func (s *MangaService) Pictures(ctx context.Context, id ID) ([]ImageSet, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	return fetch[[]ImageSet](ctx, s.c, fmt.Sprintf("/manga/%d/pictures", id), nil)
}

func (s *MangaService) Statistics(ctx context.Context, id ID) (*MangaStats, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	r, err := fetch[MangaStats](ctx, s.c, fmt.Sprintf("/manga/%d/statistics", id), nil)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *MangaService) MoreInfo(ctx context.Context, id ID) (string, error) {
	if err := s.validateID(id); err != nil {
		return "", err
	}
	type moreInfo struct {
		MoreInfo string `json:"moreinfo"`
	}
	r, err := fetch[moreInfo](ctx, s.c, fmt.Sprintf("/manga/%d/moreinfo", id), nil)
	return r.MoreInfo, err
}

func (s *MangaService) Recommendations(ctx context.Context, id ID) ([]MangaRecommendation, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	return fetch[[]MangaRecommendation](ctx, s.c, fmt.Sprintf("/manga/%d/recommendations", id), nil)
}

func (s *MangaService) UserUpdates(ctx context.Context, id ID, page int) ([]MangaUserUpdate, *Pagination, error) {
	if err := s.validateID(id); err != nil {
		return nil, nil, err
	}
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	return fetchPaged[[]MangaUserUpdate](ctx, s.c, fmt.Sprintf("/manga/%d/userupdates", id), q)
}

func (s *MangaService) Reviews(ctx context.Context, id ID, page int, preliminary, spoiler bool) ([]MangaReview, *Pagination, error) {
	if err := s.validateID(id); err != nil {
		return nil, nil, err
	}
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if preliminary {
		q.Set("preliminary", "true")
	}
	if spoiler {
		q.Set("spoiler", "true")
	}
	return fetchPaged[[]MangaReview](ctx, s.c, fmt.Sprintf("/manga/%d/reviews", id), q)
}

func (s *MangaService) Relations(ctx context.Context, id ID) ([]MangaRelation, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	return fetch[[]MangaRelation](ctx, s.c, fmt.Sprintf("/manga/%d/relations", id), nil)
}

func (s *MangaService) External(ctx context.Context, id ID) ([]ExternalLink, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}
	return fetch[[]ExternalLink](ctx, s.c, fmt.Sprintf("/manga/%d/external", id), nil)
}

func (s *MangaService) Search(ctx context.Context, opts MangaSearchOptions) ([]*Manga, *Pagination, error) {
	return fetchPaged[[]*Manga](ctx, s.c, "/manga", opts.ToValues())
}
