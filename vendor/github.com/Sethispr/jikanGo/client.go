package jikan

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const (
	_baseURL     = "https://api.jikan.moe/v4"
	_version     = "0.1.0"
	defaultRPS   = 3
	defaultBurst = 3
)

type ctxKey int

const ctxNoCache ctxKey = 1

func NoCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxNoCache, true)
}

type Client struct {
	client     *http.Client
	baseURL    *url.URL
	agent      string
	maxRetries int
	cache      Cache
	cacheTTL   time.Duration
	limiter    *rate.Limiter

	Anime          *AnimeService
	Manga          *MangaService
	Character      *CharacterService
	People         *PeopleService
	User           *UserService
	Season         *SeasonService
	Top            *TopService
	Producer       *ProducerService
	Magazine       *MagazineService
	Genre          *GenreService
	Search         *SearchService
	Review         *ReviewService
	Recommendation *RecommendationService
	Watch          *WatchService
	Club           *ClubService
	Random         *RandomService
}

type Option func(*Client)

func New(opts ...Option) *Client {
	u, _ := url.Parse(_baseURL)
	c := &Client{
		client:     &http.Client{Timeout: 30 * time.Second},
		baseURL:    u,
		agent:      "jikan-go/" + _version,
		maxRetries: 0,
		cacheTTL:   5 * time.Minute,
	}
	for _, o := range opts {
		o(c)
	}
	c.initServices()
	return c
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.client = hc
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = d
	}
}

func WithRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

func WithCache(cache Cache, ttl time.Duration) Option {
	return func(c *Client) {
		c.cache = cache
		c.cacheTTL = ttl
	}
}

// WithRateLimit enables request rate limiting at rps requests per second.
// Pass 0 or negative to use default (3 rps).
func WithRateLimit(rps int) Option {
	return func(c *Client) {
		if rps <= 0 {
			rps = defaultRPS
		}
		c.limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), defaultBurst)
	}
}

// WithRateLimiter accepts a pre configured rate limiter for more customizable control.
func WithRateLimiter(l *rate.Limiter) Option {
	return func(c *Client) {
		c.limiter = l
	}
}

func (c *Client) initServices() {
	c.Anime = &AnimeService{c}
	c.Manga = &MangaService{c}
	c.Character = &CharacterService{c}
	c.People = &PeopleService{c}
	c.User = &UserService{c}
	c.Season = &SeasonService{c}
	c.Top = &TopService{c}
	c.Producer = &ProducerService{c}
	c.Magazine = &MagazineService{c}
	c.Genre = &GenreService{c}
	c.Search = &SearchService{c}
	c.Review = &ReviewService{c}
	c.Recommendation = &RecommendationService{c}
	c.Watch = &WatchService{c}
	c.Club = &ClubService{c}
	c.Random = &RandomService{c}
}

func (c *Client) Do(ctx context.Context, method, path string, q url.Values, v interface{}) error {
	// Apply rate limiting if you configured it
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}

	if c.cache != nil && method == http.MethodGet && v != nil {
		if skip, _ := ctx.Value(ctxNoCache).(bool); !skip {
			key := c.cacheKey(method, path, q)
			if err := c.cache.Get(ctx, key, v); err == nil {
				return nil
			}
		}
	}

	var lastErr error
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.url(path, q), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", c.agent)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
			resp.Body.Close()
			lastErr = &Error{Status: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			return parseError(resp)
		}

		if v != nil {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				resp.Body.Close()
				return err
			}
		}
		resp.Body.Close()

		if c.cache != nil && method == http.MethodGet && v != nil {
			if skip, _ := ctx.Value(ctxNoCache).(bool); !skip {
				key := c.cacheKey(method, path, q)
				_ = c.cache.Set(ctx, key, v, c.cacheTTL)
			}
		}
		return nil
	}
	return lastErr
}

func (c *Client) url(path string, q url.Values) string {
	if len(q) == 0 {
		return c.baseURL.String() + path
	}
	return c.baseURL.String() + path + "?" + q.Encode()
}

func (c *Client) cacheKey(method, path string, q url.Values) string {
	h := fnv.New64a()
	h.Write([]byte(method))
	h.Write([]byte(path))
	if q != nil {
		h.Write([]byte(q.Encode()))
	}
	return fmt.Sprintf("%x", h.Sum64())
}

func fetch[T any](ctx context.Context, c *Client, path string, query url.Values) (T, error) {
	var r struct {
		Data T `json:"data"`
	}
	err := c.Do(ctx, http.MethodGet, path, query, &r)
	return r.Data, err
}

func fetchPaged[T any](ctx context.Context, c *Client, path string, query url.Values) (T, *Pagination, error) {
	var r struct {
		Data       T          `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	err := c.Do(ctx, http.MethodGet, path, query, &r)
	return r.Data, &r.Pagination, err
}
