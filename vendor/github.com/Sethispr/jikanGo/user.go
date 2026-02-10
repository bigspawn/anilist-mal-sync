package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type UserService struct {
	c *Client
}

type User struct {
	MalID    ID     `json:"mal_id"`
	Username string `json:"username"`
	URL      string `json:"url"`
	Images   struct {
		JPG struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
		WebP struct {
			ImageURL string `json:"image_url"`
		} `json:"webp"`
	} `json:"images"`
	LastOnline string `json:"last_online"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday"`
	Location   string `json:"location"`
	Joined     string `json:"joined"`
}

type UserStatistics struct {
	Anime struct {
		DaysWatched     float64 `json:"days_watched"`
		MeanScore       float64 `json:"mean_score"`
		Watching        int     `json:"watching"`
		Completed       int     `json:"completed"`
		OnHold          int     `json:"on_hold"`
		Dropped         int     `json:"dropped"`
		PlanToWatch     int     `json:"plan_to_watch"`
		TotalEntries    int     `json:"total_entries"`
		Rewatched       int     `json:"rewatched"`
		EpisodesWatched int     `json:"episodes_watched"`
	} `json:"anime"`
	Manga struct {
		DaysRead     float64 `json:"days_read"`
		MeanScore    float64 `json:"mean_score"`
		Reading      int     `json:"reading"`
		Completed    int     `json:"completed"`
		OnHold       int     `json:"on_hold"`
		Dropped      int     `json:"dropped"`
		PlanToRead   int     `json:"plan_to_read"`
		TotalEntries int     `json:"total_entries"`
		Reread       int     `json:"reread"`
		ChaptersRead int     `json:"chapters_read"`
		VolumesRead  int     `json:"volumes_read"`
	} `json:"manga"`
}

type UserFavoriteEntry struct {
	MalID     ID       `json:"mal_id"`
	URL       string   `json:"url"`
	Images    ImageSet `json:"images"`
	Title     string   `json:"title"`
	Type      string   `json:"type"`
	StartYear int      `json:"start_year"`
}

type UserFavoriteChar struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Name   string   `json:"name"`
}

type UserFavoritePerson struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Name   string   `json:"name"`
}

func (s *UserService) ByID(ctx context.Context, username string) (*User, error) {
	var r struct{ Data User }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/full", username), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *UserService) Statistics(ctx context.Context, username string) (*UserStatistics, error) {
	var r struct{ Data UserStatistics }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/statistics", username), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *UserService) About(ctx context.Context, username string) (string, error) {
	var r struct {
		Data struct {
			About string `json:"about"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/about", username), nil, &r); err != nil {
		return "", err
	}
	return r.Data.About, nil
}

func (s *UserService) History(ctx context.Context, username string, filter string, page int) ([]struct {
	Entry     Resource `json:"entry"`
	Increment int      `json:"increment"`
	Date      string   `json:"date"`
	Score     *float64 `json:"score"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	if filter != "" {
		q.Set("filter", filter)
	}

	var r struct {
		Data []struct {
			Entry     Resource `json:"entry"`
			Increment int      `json:"increment"`
			Date      string   `json:"date"`
			Score     *float64 `json:"score"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/history", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *UserService) Friends(ctx context.Context, username string, page int) ([]struct {
	User         Resource `json:"user"`
	LastOnline   string   `json:"last_online"`
	FriendsSince string   `json:"friends_since"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			User         Resource `json:"user"`
			LastOnline   string   `json:"last_online"`
			FriendsSince string   `json:"friends_since"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/friends", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *UserService) Favorites(ctx context.Context, username string) (*struct {
	Anime      []UserFavoriteEntry  `json:"anime"`
	Manga      []UserFavoriteEntry  `json:"manga"`
	Characters []UserFavoriteChar   `json:"characters"`
	People     []UserFavoritePerson `json:"people"`
}, error,
) {
	var r struct {
		Data struct {
			Anime      []UserFavoriteEntry  `json:"anime"`
			Manga      []UserFavoriteEntry  `json:"manga"`
			Characters []UserFavoriteChar   `json:"characters"`
			People     []UserFavoritePerson `json:"people"`
		} `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/favorites", username), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *UserService) Reviews(ctx context.Context, username string, page int) ([]struct {
	MalID  int      `json:"mal_id"`
	Entry  Resource `json:"entry"`
	Score  int      `json:"score"`
	Review string   `json:"review"`
	Date   string   `json:"date"`
	Votes  int      `json:"votes"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID  int      `json:"mal_id"`
			Entry  Resource `json:"entry"`
			Score  int      `json:"score"`
			Review string   `json:"review"`
			Date   string   `json:"date"`
			Votes  int      `json:"votes"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/reviews", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *UserService) Recommendations(ctx context.Context, username string, page int) ([]struct {
	MalID   string     `json:"mal_id"`
	Entry   []Resource `json:"entry"`
	Content string     `json:"content"`
	Date    string     `json:"date"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID   string     `json:"mal_id"`
			Entry   []Resource `json:"entry"`
			Content string     `json:"content"`
			Date    string     `json:"date"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/recommendations", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *UserService) Clubs(ctx context.Context, username string, page int) ([]struct {
	MalID ID     `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}, *Pagination, error,
) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID ID     `json:"mal_id"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/clubs", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *UserService) External(ctx context.Context, username string) ([]struct {
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
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/external", username), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
