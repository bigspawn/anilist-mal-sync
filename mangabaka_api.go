package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// mangaBakaMaxFailureStreak matches the Hato breaker threshold (PR #94):
// an unreachable third party must not cost retries on every remaining entry.
const mangaBakaMaxFailureStreak = 5

// Cache/query directions. The value is also the undocumented `q=` prefix
// MangaBaka expects (`mal:`, not `my_anime_list:` — verified 2026-09-12).
const (
	mangaBakaDirectionAniList = "anilist"
	mangaBakaDirectionMAL     = "mal"

	mangaBakaSeriesStateActive = "active"
)

// MangaBakaClient is an HTTP client for the MangaBaka API (https://mangabaka.org).
// Supports bidirectional AniList<->MAL manga ID mapping with persistent JSON caching.
type MangaBakaClient struct {
	apiSource
	cache *MangaBakaCache // Persistent cache (can be nil)
}

// MangaBakaSearchResponse represents the response from GET /v1/series/search.
type MangaBakaSearchResponse struct {
	Data []MangaBakaSeries `json:"data"`
}

// MangaBakaSeries is one search hit.
type MangaBakaSeries struct {
	ID     int             `json:"id"`
	Title  string          `json:"title"`
	Type   string          `json:"type"`
	State  string          `json:"state"`
	Source MangaBakaSource `json:"source"`
}

// MangaBakaSource carries the cross-provider ids MangaBaka aggregates.
type MangaBakaSource struct {
	AniList     MangaBakaSourceID `json:"anilist"`
	MyAnimeList MangaBakaSourceID `json:"my_anime_list"`
}

// MangaBakaSourceID is a single provider id, nil when the series has none.
type MangaBakaSourceID struct {
	ID *int `json:"id"`
}

// NewMangaBakaClient creates a new MangaBaka API client with optional caching.
// If cacheDir is empty, caching is disabled (cache = nil).
func NewMangaBakaClient(
	ctx context.Context,
	baseURL string,
	timeout time.Duration,
	cacheDir string,
	cacheMaxAgeStr string,
) *MangaBakaClient {
	if baseURL == "" {
		baseURL = defaultMangaBakaBaseURL
	}

	maxAge := defaultHatoCacheMaxAge
	parsed, err := time.ParseDuration(cacheMaxAgeStr)
	if err == nil {
		maxAge = parsed
	}

	var cache *MangaBakaCache
	if cacheDir != "" {
		cache = NewMangaBakaCache(cacheDir, maxAge) //nolint:contextcheck // Cache init doesn't need context
		LogInfoSuccess(ctx, "MangaBaka cache loaded (%d entries)", cache.Size())
	}

	return &MangaBakaClient{
		apiSource: newAPISource("MangaBaka", baseURL, timeout, mangaBakaMaxFailureStreak),
		cache:     cache,
	}
}

// SaveCache persists the cache to disk if there are unsaved changes.
func (c *MangaBakaClient) SaveCache(ctx context.Context) error {
	if c.cache == nil {
		return nil
	}
	return c.cache.Save(ctx)
}

// getCachedResult attempts to retrieve a cached lookup result for direction+id.
func (c *MangaBakaClient) getCachedResult(direction string, id int) (*MangaBakaLookupResult, bool) {
	if c.cache == nil {
		return nil, false
	}
	return c.cache.Get(direction, id)
}

// setCachedResult stores a lookup result, positive or negative, so a cold
// sync does not re-spend the 30 req/min search-tier budget on repeat runs.
func (c *MangaBakaClient) setCachedResult(direction string, id int, result MangaBakaLookupResult) {
	if c.cache != nil {
		c.cache.Set(direction, id, result)
	}
}

// GetMALID returns the MAL ID for a given AniList manga ID.
// Checks cache first, then queries the API if needed.
func (c *MangaBakaClient) GetMALID(ctx context.Context, anilistID int) (int, bool, error) {
	return c.lookup(ctx, mangaBakaDirectionAniList, anilistID)
}

// GetAniListID returns the AniList ID for a given MAL manga ID.
// Checks cache first, then queries the API if needed.
func (c *MangaBakaClient) GetAniListID(ctx context.Context, malID int) (int, bool, error) {
	return c.lookup(ctx, mangaBakaDirectionMAL, malID)
}

// lookup resolves id (a value of the given direction) to its opposite-side id.
func (c *MangaBakaClient) lookup(ctx context.Context, direction string, id int) (int, bool, error) {
	if cached, found := c.getCachedResult(direction, id); found {
		LogDebug(ctx, "[MANGABAKA CACHE] HIT: %s %d -> found=%v id=%d", direction, id, cached.Found, cached.ID)
		return cached.ID, cached.Found, nil
	}

	// A negative result is never cached here: MangaBaka being down is not an answer.
	if c.gaveUp() {
		return 0, false, nil
	}

	url := fmt.Sprintf("%s/series/search?q=%s:%d&limit=5", c.baseURL, direction, id)
	LogDebug(ctx, "[MANGABAKA API] GET %s", url)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		LogDebug(ctx, "[MANGABAKA API] Error: %v", err)
		return 0, false, err
	}

	var targetID int
	var found bool
	if resp != nil {
		targetID, found = selectMangaBakaHit(resp.Data, direction)
	}

	c.setCachedResult(direction, id, MangaBakaLookupResult{ID: targetID, Found: found})
	LogDebug(ctx, "[MANGABAKA API] Response: %s %d -> found=%v id=%d", direction, id, found, targetID)

	return targetID, found, nil
}

// selectMangaBakaHit applies the disambiguation rule: keep only active series,
// then require every surviving hit to agree on the opposite-side id.
func selectMangaBakaHit(hits []MangaBakaSeries, direction string) (int, bool) {
	var targetID int
	seen := false

	for _, hit := range hits {
		if hit.State != mangaBakaSeriesStateActive {
			continue
		}

		id := oppositeSourceID(hit.Source, direction)
		switch {
		case !seen:
			targetID, seen = id, true
		case id != targetID:
			return 0, false // ambiguous — never guess
		}
	}

	if !seen || targetID <= 0 {
		return 0, false
	}
	return targetID, true
}

// oppositeSourceID reads the id on the side opposite to direction, treating a
// nil provider id (series known to MangaBaka but unmapped on that side) as 0.
func oppositeSourceID(src MangaBakaSource, direction string) int {
	var ptr *int
	switch direction {
	case mangaBakaDirectionAniList:
		ptr = src.MyAnimeList.ID
	case mangaBakaDirectionMAL:
		ptr = src.AniList.ID
	}
	if ptr == nil {
		return 0
	}
	return *ptr
}

// doRequest issues a GET and returns the decoded search response.
// A nil response with a nil error means "not found" — HTTP 404 or an empty
// data array — never an error, per the MangaBaka undocumented-syntax contract.
func (c *MangaBakaClient) doRequest(ctx context.Context, url string) (*MangaBakaSearchResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "anilist-mal-sync")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request: %w", err)
		c.noteFailure(ctx, err)
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	LogDebug(ctx, "[MANGABAKA API] Status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		c.noteSuccess()
		return nil, nil //nolint:nilnil // nil means "not found", not an error
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status: %s", resp.Status)
		c.noteFailure(ctx, err)
		return nil, err
	}

	var searchResp MangaBakaSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		err = fmt.Errorf("decode response: %w", err)
		c.noteFailure(ctx, err)
		return nil, err
	}

	c.noteSuccess()
	if len(searchResp.Data) == 0 {
		return nil, nil //nolint:nilnil // nil means "not found", not an error
	}

	return &searchResp, nil
}
