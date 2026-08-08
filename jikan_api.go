package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	defaultJikanBaseURL     = "https://api.jikan.moe/v4"
	defaultJikanCacheMaxAge = 168 * time.Hour // 7 days

	// Documented Jikan v4 limits: 3 requests/second AND 60 requests/minute.
	// The per-minute cap binds first — honouring only the per-second one
	// exceeds the quota after ~30 seconds of steady traffic.
	jikanMinRequestInterval = 340 * time.Millisecond
	jikanRequestsPerMinute  = 60
	jikanMinuteWindow       = time.Minute

	// Rate-limit windows reset in tens of seconds, so a 429 needs more
	// than the two attempts the other clients get.
	jikanMaxRetries            = 4
	jikanResponseHeaderTimeout = 20 * time.Second
)

// JikanMangaData contains the manga data we extract from Jikan API responses.
type JikanMangaData struct {
	MalID         int      `json:"mal_id"`
	Title         string   `json:"title"`
	TitleEnglish  string   `json:"title_english"`
	TitleJapanese string   `json:"title_japanese"`
	TitleSynonyms []string `json:"title_synonyms"`
	Type          string   `json:"type"`
	Chapters      int      `json:"chapters"`
	Volumes       int      `json:"volumes"`
	Status        string   `json:"status"`
}

// jikanResponse wraps the Jikan API v4 single-item response.
type jikanResponse struct {
	Data JikanMangaData `json:"data"`
}

// jikanSearchResponse wraps the Jikan API v4 search response.
type jikanSearchResponse struct {
	Data []JikanMangaData `json:"data"`
}

// JikanFavoriteEntry represents a favorite anime or manga from Jikan API.
type JikanFavoriteEntry struct {
	MalID int    `json:"mal_id"`
	Title string `json:"title"`
}

// jikanFavoritesResponse wraps the Jikan API v4 user favorites response.
type jikanFavoritesResponse struct {
	Data struct {
		Anime []JikanFavoriteEntry `json:"anime"`
		Manga []JikanFavoriteEntry `json:"manga"`
	} `json:"data"`
}

// JikanClient is an API client for Jikan (unofficial MAL REST API).
// Implements rate limiting and file-based caching.
type JikanClient struct {
	baseURL    string
	httpClient HTTPClient
	cache      *JikanCache
}

// NewJikanClient creates a new Jikan API client with caching.
func NewJikanClient(ctx context.Context, cacheDir string, cacheMaxAgeStr string) *JikanClient {
	maxAge := defaultJikanCacheMaxAge
	if parsed, err := time.ParseDuration(cacheMaxAgeStr); err == nil {
		maxAge = parsed
	}

	cache := NewJikanCache(cacheDir, maxAge) //nolint:contextcheck // Cache init doesn't need context
	LogInfoSuccess(ctx, "Jikan cache loaded (%d entries)", cache.Size())

	return &JikanClient{
		baseURL:    defaultJikanBaseURL,
		httpClient: newJikanHTTPClient(),
		cache:      cache,
	}
}

// newJikanHTTPClient stacks the retry logic ON TOP of the rate limiter, so
// every attempt — retries included — consumes a quota slot. A client-wide
// timeout is deliberately absent: it would cover the whole retry sequence,
// including backoff sleeps. Each attempt is bounded by the transport instead.
func newJikanHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // stdlib guarantees the concrete type
	transport.ResponseHeaderTimeout = jikanResponseHeaderTimeout

	limited := &http.Client{Transport: &jikanRateLimiter{underlying: transport}}

	return &http.Client{Transport: NewRetryableTransport(limited, jikanMaxRetries)}
}

// jikanRateLimiter paces outgoing requests to stay inside both documented
// Jikan quotas at once: a minimum gap between requests and a rolling minute.
type jikanRateLimiter struct {
	underlying http.RoundTripper

	mu   sync.Mutex
	sent []time.Time // send times inside the trailing minute, oldest first
}

func (l *jikanRateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	wait := l.reserve(time.Now())
	if wait <= 0 {
		return l.underlying.RoundTrip(req)
	}

	LogDebug(req.Context(), "[JIKAN RATE] waiting %v before %s", wait, req.URL)

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}

	return l.underlying.RoundTrip(req)
}

// reserve claims the next send slot and reports how long to wait for it.
func (l *jikanRateLimiter) reserve(now time.Time) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	sendAt := now
	if len(l.sent) > 0 {
		if next := l.sent[len(l.sent)-1].Add(jikanMinRequestInterval); next.After(sendAt) {
			sendAt = next
		}
	}
	if len(l.sent) >= jikanRequestsPerMinute {
		// Wait for the oldest of the last N sends to leave the window.
		if freed := l.sent[len(l.sent)-jikanRequestsPerMinute].Add(jikanMinuteWindow); freed.After(sendAt) {
			sendAt = freed
		}
	}

	l.sent = append(l.sent, sendAt)
	l.sent = droppedBefore(l.sent, sendAt.Add(-jikanMinuteWindow))

	return sendAt.Sub(now)
}

// droppedBefore returns the tail of times that is newer than cutoff.
func droppedBefore(times []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(times) && !times[i].After(cutoff) {
		i++
	}
	return times[i:]
}

// GetMangaByMALID retrieves manga data by MAL ID, checking cache first.
// Returns (data, found). Errors are non-fatal and logged internally.
func (c *JikanClient) GetMangaByMALID(ctx context.Context, malID int) (*JikanMangaData, bool) {
	if malID <= 0 {
		return nil, false
	}

	// Check cache first
	if cached, found := c.cache.Get(malID); found {
		if string(cached) == "null" {
			LogDebug(ctx, "[JIKAN CACHE] HIT (negative): manga %d", malID)
			return nil, false
		}
		var data JikanMangaData
		err := json.Unmarshal(cached, &data)
		if err == nil {
			LogDebug(ctx, "[JIKAN CACHE] HIT: manga %d -> %s", malID, data.Title)
			return &data, true
		}
	}

	// Cache miss — call API
	apiURL := fmt.Sprintf("%s/manga/%d", c.baseURL, malID)
	LogDebug(ctx, "[JIKAN API] GET %s", apiURL)

	resp, err := c.doRequest(ctx, apiURL)
	if err != nil {
		// A failed lookup is indistinguishable from "no such manga" downstream,
		// so surface it — silent rate limiting degrades matching quality.
		LogWarn(ctx, "[JIKAN API] manga %d: %v", malID, err)
		return nil, false
	}
	if resp == nil {
		LogDebug(ctx, "[JIKAN API] manga %d: not found (404)", malID)
		c.cache.Set(malID, json.RawMessage("null"))
		return nil, false
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	var jResp jikanResponse
	if err := json.NewDecoder(resp.Body).Decode(&jResp); err != nil {
		LogDebug(ctx, "[JIKAN API] manga %d: decode error: %v", malID, err)
		return nil, false
	}

	data := &jResp.Data
	encoded, err := json.Marshal(data)
	if err == nil {
		c.cache.Set(malID, encoded)
	}

	LogDebug(ctx, "[JIKAN API] manga %d: found -> %s", malID, data.Title)
	return data, true
}

// SearchManga searches for manga by title, checking cache first.
// Errors are non-fatal and logged internally.
func (c *JikanClient) SearchManga(ctx context.Context, query string) []JikanMangaData {
	if query == "" {
		return nil
	}

	// Check cache first
	if cached, found := c.cache.GetSearch(query); found {
		var results []JikanMangaData
		err := json.Unmarshal(cached, &results)
		if err == nil {
			LogDebug(ctx, "[JIKAN CACHE] HIT: search %q -> %d results", query, len(results))
			return results
		}
	}

	// Cache miss — call API
	params := url.Values{}
	params.Set("q", query)
	apiURL := fmt.Sprintf("%s/manga?%s", c.baseURL, params.Encode())
	LogDebug(ctx, "[JIKAN API] GET %s", apiURL)

	resp, err := c.doRequest(ctx, apiURL)
	if err != nil {
		LogWarn(ctx, "[JIKAN API] search %q: %v", query, err)
		return nil
	}
	if resp == nil {
		return nil
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	var jResp jikanSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&jResp); err != nil {
		LogDebug(ctx, "[JIKAN API] search %q: decode error: %v", query, err)
		return nil
	}

	encoded, err := json.Marshal(jResp.Data)
	if err == nil {
		c.cache.SetSearch(query, encoded)
	}

	LogDebug(ctx, "[JIKAN API] search %q: found %d results", query, len(jResp.Data))
	return jResp.Data
}

// GetUserFavorites fetches a MAL user's favorite anime and manga via Jikan API.
// Returns sets of MAL IDs for anime and manga favorites.
// The Jikan API endpoint returns public user data and does not require authentication.
func (c *JikanClient) GetUserFavorites(ctx context.Context, username string) (
	animeIDs map[int]struct{}, mangaIDs map[int]struct{}, err error,
) {
	if username == "" {
		return nil, nil, errors.New("username cannot be empty")
	}

	apiURL := fmt.Sprintf("%s/users/%s/favorites", c.baseURL, url.PathEscape(username))
	LogDebug(ctx, "[JIKAN API] GET %s", apiURL)

	resp, err := c.doRequest(ctx, apiURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user favorites: %w", err)
	}
	if resp == nil {
		// Jikan answers 404 both for an unknown user and for a MyAnimeList error.
		return nil, nil, fmt.Errorf("user %s not found, profile is private, or MyAnimeList rejected the lookup", username)
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	var jResp jikanFavoritesResponse
	if err := json.NewDecoder(resp.Body).Decode(&jResp); err != nil {
		LogDebug(ctx, "[JIKAN API] user %s favorites: decode error: %v", username, err)
		return nil, nil, fmt.Errorf("failed to decode favorites response: %w", err)
	}

	animeIDs = make(map[int]struct{}, len(jResp.Data.Anime))
	for _, fav := range jResp.Data.Anime {
		animeIDs[fav.MalID] = struct{}{}
	}

	mangaIDs = make(map[int]struct{}, len(jResp.Data.Manga))
	for _, fav := range jResp.Data.Manga {
		mangaIDs[fav.MalID] = struct{}{}
	}

	LogDebug(ctx, "[JIKAN API] user %s favorites: %d anime, %d manga", username, len(animeIDs), len(mangaIDs))
	return animeIDs, mangaIDs, nil
}

// SaveCache persists the cache to disk if there are unsaved changes.
func (c *JikanClient) SaveCache(ctx context.Context) error {
	return c.cache.Save(ctx)
}

// doRequest makes an HTTP GET request and returns the response.
// Returns (nil, nil) for 404 (not found).
func (c *JikanClient) doRequest(ctx context.Context, apiURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close() //nolint:errcheck,gosec // best effort close
		return nil, nil   //nolint:nilnil // nil means "not found"
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close() //nolint:errcheck,gosec // best effort close
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return resp, nil
}

// matchJikanMangaToSource checks if a Jikan manga result matches a source manga
// by comparing titles using the existing multi-level title matching.
func matchJikanMangaToSource(ctx context.Context, jikanData *JikanMangaData, srcTitleEN, srcTitleJP, srcTitleRomaji string) bool {
	// Build all possible Jikan titles to compare against
	jikanTitles := make([]struct{ en, jp, romaji string }, 0, 1+len(jikanData.TitleSynonyms))
	jikanTitles = append(jikanTitles, struct{ en, jp, romaji string }{jikanData.TitleEnglish, jikanData.TitleJapanese, jikanData.Title})

	// Also try matching with synonyms as English title
	for _, synonym := range jikanData.TitleSynonyms {
		jikanTitles = append(jikanTitles, struct{ en, jp, romaji string }{synonym, jikanData.TitleJapanese, jikanData.Title})
	}

	for _, jt := range jikanTitles {
		if titleMatchingLevels(ctx, srcTitleEN, srcTitleJP, srcTitleRomaji, jt.en, jt.jp, jt.romaji) {
			return true
		}
	}

	// Also try cross-matching: source English vs Jikan romaji, source Romaji vs Jikan English
	if srcTitleEN != "" && jikanData.Title != "" {
		if normalizeTitle(srcTitleEN) == normalizeTitle(jikanData.Title) {
			return true
		}
	}
	if srcTitleRomaji != "" && jikanData.TitleEnglish != "" {
		if normalizeTitle(srcTitleRomaji) == normalizeTitle(jikanData.TitleEnglish) {
			return true
		}
	}

	return false
}

// findBestJikanMatch finds the best matching manga from Jikan search results.
// Returns the MAL ID of the best match, or 0 if no match found.
func findBestJikanMatch(ctx context.Context, results []JikanMangaData, srcTitleEN, srcTitleJP, srcTitleRomaji string) int {
	for _, result := range results {
		if matchJikanMangaToSource(ctx, &result, srcTitleEN, srcTitleJP, srcTitleRomaji) {
			return result.MalID
		}
	}
	return 0
}

// searchTitlesForJikan returns a list of search queries to try against Jikan API
// for a given manga source, in order of preference.
func searchTitlesForJikan(titleEN, _, titleRomaji string) []string {
	titles := make([]string, 0, 2)
	seen := make(map[string]struct{})

	for _, t := range []string{titleRomaji, titleEN} {
		if t == "" {
			continue
		}
		normalized := normalizeTitle(t)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		titles = append(titles, t)
	}

	return titles
}
