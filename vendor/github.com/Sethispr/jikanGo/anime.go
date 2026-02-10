package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type AnimeService struct {
	c *Client
}

type Anime struct {
	MalID   ID       `json:"mal_id"`
	URL     string   `json:"url"`
	Images  ImageSet `json:"images"`
	Trailer struct {
		YoutubeID string `json:"youtube_id"`
		URL       string `json:"url"`
		EmbedURL  string `json:"embed_url"`
		Images    struct {
			ImageURL        string `json:"image_url"`
			SmallImageURL   string `json:"small_image_url"`
			MediumImageURL  string `json:"medium_image_url"`
			LargeImageURL   string `json:"large_image_url"`
			MaximumImageURL string `json:"maximum_image_url"`
		} `json:"images"`
	} `json:"trailer"`
	Title         string    `json:"title"`
	TitleEnglish  string    `json:"title_english"`
	TitleJapanese string    `json:"title_japanese"`
	TitleSynonyms []string  `json:"title_synonyms"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
	Episodes      int       `json:"episodes"`
	Status        string    `json:"status"`
	Airing        bool      `json:"airing"`
	Aired         DateRange `json:"aired"`
	Duration      string    `json:"duration"`
	Rating        string    `json:"rating"`
	Score         float64   `json:"score"`
	ScoredBy      int       `json:"scored_by"`
	Rank          int       `json:"rank"`
	Popularity    int       `json:"popularity"`
	Members       int       `json:"members"`
	Favorites     int       `json:"favorites"`
	Synopsis      string    `json:"synopsis"`
	Background    string    `json:"background"`
	Season        string    `json:"season"`
	Year          int       `json:"year"`
	Broadcast     struct {
		Day      string `json:"day"`
		Time     string `json:"time"`
		Timezone string `json:"timezone"`
		String   string `json:"string"`
	} `json:"broadcast"`
	Producers      []Resource `json:"producers"`
	Licensors      []Resource `json:"licensors"`
	Studios        []Resource `json:"studios"`
	Genres         []Resource `json:"genres"`
	ExplicitGenres []Resource `json:"explicit_genres"`
	Themes         []Resource `json:"themes"`
	Demographics   []Resource `json:"demographics"`
}

func (s *AnimeService) ByID(ctx context.Context, id ID) (*Anime, error) {
	var r struct{ Data Anime }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) Characters(ctx context.Context, id ID) ([]struct {
	Character   Resource `json:"character"`
	Role        string   `json:"role"`
	VoiceActors []struct {
		Person   Resource `json:"person"`
		Language string   `json:"language"`
	} `json:"voice_actors"`
}, error,
) {
	var r struct {
		Data []struct {
			Character   Resource `json:"character"`
			Role        string   `json:"role"`
			VoiceActors []struct {
				Person   Resource `json:"person"`
				Language string   `json:"language"`
			} `json:"voice_actors"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/characters", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Staff(ctx context.Context, id ID) ([]struct {
	Person    Resource `json:"person"`
	Positions []string `json:"positions"`
}, error,
) {
	var r struct {
		Data []struct {
			Person    Resource `json:"person"`
			Positions []string `json:"positions"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/staff", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Episodes(ctx context.Context, id ID, page int) ([]struct {
	MalID         ID      `json:"mal_id"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	TitleJapanese string  `json:"title_japanese"`
	TitleRomanji  string  `json:"title_romanji"`
	Aired         string  `json:"aired"`
	Score         float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID         ID      `json:"mal_id"`
			URL           string  `json:"url"`
			Title         string  `json:"title"`
			TitleJapanese string  `json:"title_japanese"`
			TitleRomanji  string  `json:"title_romanji"`
			Aired         string  `json:"aired"`
			Score         float64 `json:"score"`
			Filler        bool    `json:"filler"`
			Recap         bool    `json:"recap"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/episodes", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *AnimeService) EpisodeByID(ctx context.Context, animeID ID, episode int) (*struct {
	MalID         ID      `json:"mal_id"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	TitleJapanese string  `json:"title_japanese"`
	TitleRomanji  string  `json:"title_romanji"`
	Synopsis      string  `json:"synopsis"`
	Aired         string  `json:"aired"`
	Score         float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
}, error,
) {
	var r struct {
		Data struct {
			MalID         ID      `json:"mal_id"`
			URL           string  `json:"url"`
			Title         string  `json:"title"`
			TitleJapanese string  `json:"title_japanese"`
			TitleRomanji  string  `json:"title_romanji"`
			Synopsis      string  `json:"synopsis"`
			Aired         string  `json:"aired"`
			Score         float64 `json:"score"`
			Filler        bool    `json:"filler"`
			Recap         bool    `json:"recap"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/episodes/%d", animeID, episode), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) News(ctx context.Context, id ID, page int) ([]struct {
	MalID    ID     `json:"mal_id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Author   string `json:"author_username"`
	Comments int    `json:"comments"`
	Excerpt  string `json:"excerpt"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID    ID     `json:"mal_id"`
			Title    string `json:"title"`
			Date     string `json:"date"`
			Author   string `json:"author_username"`
			Comments int    `json:"comments"`
			Excerpt  string `json:"excerpt"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/news", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

type ForumFilter string

const (
	ForumFilterAll     ForumFilter = "all"
	ForumFilterEpisode ForumFilter = "episode"
	ForumFilterOther   ForumFilter = "other"
)

func (s *AnimeService) Forum(ctx context.Context, id ID, filter ForumFilter) ([]struct {
	MalID    ID     `json:"mal_id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Author   string `json:"author_username"`
	Comments int    `json:"comments"`
}, error,
) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", string(filter))
	}
	var r struct {
		Data []struct {
			MalID    ID     `json:"mal_id"`
			Title    string `json:"title"`
			Date     string `json:"date"`
			Author   string `json:"author_username"`
			Comments int    `json:"comments"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/forum", id), q, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Videos(ctx context.Context, id ID) (*struct {
	Promos []struct {
		Title   string `json:"title"`
		Trailer struct {
			YoutubeID string `json:"youtube_id"`
			URL       string `json:"url"`
			EmbedURL  string `json:"embed_url"`
			Images    struct {
				DefaultImageURL string `json:"default_image_url"`
				SmallImageURL   string `json:"small_image_url"`
				MediumImageURL  string `json:"medium_image_url"`
				LargeImageURL   string `json:"large_image_url"`
				MaximumImageURL string `json:"maximum_image_url"`
			} `json:"images"`
		} `json:"trailer"`
	} `json:"promos"`
	Episodes []struct {
		MalID   ID     `json:"mal_id"`
		URL     string `json:"url"`
		Title   string `json:"title"`
		Episode string `json:"episode"`
		Images  struct {
			JPG struct {
				ImageURL string `json:"image_url"`
			} `json:"jpg"`
		} `json:"images"`
	} `json:"episodes"`
}, error,
) {
	var r struct {
		Data struct {
			Promos []struct {
				Title   string `json:"title"`
				Trailer struct {
					YoutubeID string `json:"youtube_id"`
					URL       string `json:"url"`
					EmbedURL  string `json:"embed_url"`
					Images    struct {
						DefaultImageURL string `json:"default_image_url"`
						SmallImageURL   string `json:"small_image_url"`
						MediumImageURL  string `json:"medium_image_url"`
						LargeImageURL   string `json:"large_image_url"`
						MaximumImageURL string `json:"maximum_image_url"`
					} `json:"images"`
				} `json:"trailer"`
			} `json:"promos"`
			Episodes []struct {
				MalID   ID     `json:"mal_id"`
				URL     string `json:"url"`
				Title   string `json:"title"`
				Episode string `json:"episode"`
				Images  struct {
					JPG struct {
						ImageURL string `json:"image_url"`
					} `json:"jpg"`
				} `json:"images"`
			} `json:"episodes"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/videos", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) Pictures(ctx context.Context, id ID) ([]ImageSet, error) {
	var r struct {
		Data []struct {
			Images ImageSet `json:"images"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/pictures", id), nil, &r); err != nil {
		return nil, err
	}
	sets := make([]ImageSet, len(r.Data))
	for i, d := range r.Data {
		sets[i] = d.Images
	}
	return sets, nil
}

// Statistics returns full stats with score breakdown
func (s *AnimeService) Statistics(ctx context.Context, id ID) (*struct {
	Watching    int `json:"watching"`
	Completed   int `json:"completed"`
	OnHold      int `json:"on_hold"`
	Dropped     int `json:"dropped"`
	PlanToWatch int `json:"plan_to_watch"`
	Total       int `json:"total"`
	Scores      []struct {
		Score      int     `json:"score"`
		Votes      int     `json:"votes"`
		Percentage float64 `json:"percentage"`
	} `json:"scores"`
}, error,
) {
	var r struct {
		Data struct {
			Watching    int `json:"watching"`
			Completed   int `json:"completed"`
			OnHold      int `json:"on_hold"`
			Dropped     int `json:"dropped"`
			PlanToWatch int `json:"plan_to_watch"`
			Total       int `json:"total"`
			Scores      []struct {
				Score      int     `json:"score"`
				Votes      int     `json:"votes"`
				Percentage float64 `json:"percentage"`
			} `json:"scores"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/statistics", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) MoreInfo(ctx context.Context, id ID) (string, error) {
	var r struct {
		Data struct {
			MoreInfo string `json:"moreinfo"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/moreinfo", id), nil, &r); err != nil {
		return "", err
	}
	return r.Data.MoreInfo, nil
}

func (s *AnimeService) Recommendations(ctx context.Context, id ID) ([]struct {
	Entry Resource `json:"entry"`
	URL   string   `json:"url"`
	Votes int      `json:"votes"`
}, error,
) {
	var r struct {
		Data []struct {
			Entry Resource `json:"entry"`
			URL   string   `json:"url"`
			Votes int      `json:"votes"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/recommendations", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) UserUpdates(ctx context.Context, id ID, page int) ([]struct {
	User          Resource `json:"user"`
	Score         float64  `json:"score"`
	Status        string   `json:"status"`
	EpisodesSeen  int      `json:"episodes_seen"`
	EpisodesTotal int      `json:"episodes_total"`
	Date          string   `json:"date"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			User          Resource `json:"user"`
			Score         float64  `json:"score"`
			Status        string   `json:"status"`
			EpisodesSeen  int      `json:"episodes_seen"`
			EpisodesTotal int      `json:"episodes_total"`
			Date          string   `json:"date"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/userupdates", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *AnimeService) Reviews(ctx context.Context, id ID, page int) ([]struct {
	User            Resource `json:"user"`
	MalID           int      `json:"mal_id"`
	Score           int      `json:"score"`
	Review          string   `json:"review"`
	EpisodesWatched int      `json:"episodes_watched"`
	Date            string   `json:"date"`
	Votes           int      `json:"votes"`
	Scores          struct {
		Overall   int `json:"overall"`
		Story     int `json:"story"`
		Animation int `json:"animation"`
		Sound     int `json:"sound"`
		Character int `json:"character"`
		Enjoyment int `json:"enjoyment"`
	} `json:"scores"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			User            Resource `json:"user"`
			MalID           int      `json:"mal_id"`
			Score           int      `json:"score"`
			Review          string   `json:"review"`
			EpisodesWatched int      `json:"episodes_watched"`
			Date            string   `json:"date"`
			Votes           int      `json:"votes"`
			Scores          struct {
				Overall   int `json:"overall"`
				Story     int `json:"story"`
				Animation int `json:"animation"`
				Sound     int `json:"sound"`
				Character int `json:"character"`
				Enjoyment int `json:"enjoyment"`
			} `json:"scores"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/reviews", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *AnimeService) Relations(ctx context.Context, id ID) ([]struct {
	Relation string     `json:"relation"`
	Entry    []Resource `json:"entry"`
}, error,
) {
	var r struct {
		Data []struct {
			Relation string     `json:"relation"`
			Entry    []Resource `json:"entry"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/relations", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Themes(ctx context.Context, id ID) (*struct {
	Openings []string `json:"openings"`
	Endings  []string `json:"endings"`
}, error,
) {
	var r struct {
		Data struct {
			Openings []string `json:"openings"`
			Endings  []string `json:"endings"`
		}
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/themes", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) External(ctx context.Context, id ID) ([]struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}, error,
) {
	var r struct {
		Data []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/external", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
