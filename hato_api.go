package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultHatoBaseURL = "https://hato.malupdaterosx.moe"

// HatoClient is an HTTP client for the Hato API (https://hato.malupdaterosx.moe).
// Supports both anime and manga ID mapping with persistent JSON caching.
type HatoClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *HatoCache // Persistent cache (can be nil)
}

// HatoResponse represents the response from /api/mappings/{service}/{media_type}/{id}.
type HatoResponse struct {
	Data HatoResponseData `json:"data"`
}

// HatoResponseData contains the actual mapping data.
// This structure is shared between API responses and cache entries.
type HatoResponseData struct {
	AniDBID   *int    `json:"anidb_id,omitempty"`
	AniListID *int    `json:"anilist_id,omitempty"`
	KitsuID   *int    `json:"kitsu_id,omitempty"`
	MalID     *int    `json:"mal_id,omitempty"`
	NotifyID  *string `json:"notify_id,omitempty"`
	Type      *int    `json:"type,omitempty"`     // 0 = anime, 1 = manga
	TypeStr   *string `json:"type_str,omitempty"` // "anime" or "manga"
}

// NewHatoClient creates a new Hato API client with optional caching.
// If cacheDir is empty, caching is disabled (cache = nil).
func NewHatoClient(ctx context.Context, baseURL string, timeout time.Duration, cacheDir string) *HatoClient {
	if baseURL == "" {
		baseURL = defaultHatoBaseURL
	}

	var cache *HatoCache
	if cacheDir != "" {
		var err error
		cache, err = NewHatoCache(cacheDir) //nolint:contextcheck // Cache init doesn't need context
		if err != nil {
			LogWarn(ctx, "Failed to initialize Hato cache: %v (caching disabled)", err)
		} else {
			LogInfoSuccess(ctx, "Hato cache loaded (%d entries)", cache.Size())
		}
	}

	return &HatoClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cache: cache,
	}
}

// GetAniListID returns the AniList ID for a given MAL ID and media type.
// mediaType should be "anime" or "manga".
// Checks cache first, then makes API request if needed.
//
//nolint:dupl // Similar to GetMALID but with different field access
func (c *HatoClient) GetAniListID(ctx context.Context, malID int, mediaType string) (int, bool, error) {
	// Check cache first
	if c.cache != nil {
		if data, found := c.cache.Get("mal", mediaType, malID); found {
			if data.AniListID != nil && *data.AniListID > 0 {
				LogDebug(ctx, "[HATO CACHE] HIT: MAL %d -> AniList %d (%s)", malID, *data.AniListID, mediaType)
				return *data.AniListID, true, nil
			}
			// Cached negative result
			LogDebug(ctx, "[HATO CACHE] HIT: MAL %d -> not found (cached) (%s)", malID, mediaType)
			return 0, false, nil
		}
	}

	// Cache miss - make API request
	url := fmt.Sprintf("%s/api/mappings/mal/%s/%d", c.baseURL, mediaType, malID)
	LogDebug(ctx, "[HATO API] GET %s", url)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		LogDebug(ctx, "[HATO API] Error: %v", err)
		return 0, false, err
	}
	if resp == nil || resp.Data.AniListID == nil {
		LogDebug(ctx, "[HATO API] Response: not found (404 or null)")
		// Cache negative result to avoid repeated API calls
		if c.cache != nil {
			c.cache.Set("mal", mediaType, malID, HatoResponseData{}) // Empty data = not found
		}
		return 0, false, nil
	}

	anilistID := *resp.Data.AniListID
	LogDebug(ctx, "[HATO API] Response: AniList ID = %d", anilistID)

	// Save complete response to cache
	if c.cache != nil {
		c.cache.Set("mal", mediaType, malID, resp.Data)
	}

	return anilistID, true, nil
}

// GetMALID returns the MAL ID for a given AniList ID and media type.
// mediaType should be "anime" or "manga".
// Checks cache first, then makes API request if needed.
//
//nolint:dupl // Similar to GetAniListID but with different field access
func (c *HatoClient) GetMALID(ctx context.Context, anilistID int, mediaType string) (int, bool, error) {
	// Check cache first
	if c.cache != nil {
		if data, found := c.cache.Get("anilist", mediaType, anilistID); found {
			if data.MalID != nil && *data.MalID > 0 {
				LogDebug(ctx, "[HATO CACHE] HIT: AniList %d -> MAL %d (%s)", anilistID, *data.MalID, mediaType)
				return *data.MalID, true, nil
			}
			// Cached negative result
			LogDebug(ctx, "[HATO CACHE] HIT: AniList %d -> not found (cached) (%s)", anilistID, mediaType)
			return 0, false, nil
		}
	}

	// Cache miss - make API request
	url := fmt.Sprintf("%s/api/mappings/anilist/%s/%d", c.baseURL, mediaType, anilistID)
	LogDebug(ctx, "[HATO API] GET %s", url)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		LogDebug(ctx, "[HATO API] Error: %v", err)
		return 0, false, err
	}
	if resp == nil || resp.Data.MalID == nil {
		LogDebug(ctx, "[HATO API] Response: not found (404 or null)")
		// Cache negative result
		if c.cache != nil {
			c.cache.Set("anilist", mediaType, anilistID, HatoResponseData{}) // Empty data = not found
		}
		return 0, false, nil
	}

	malID := *resp.Data.MalID
	LogDebug(ctx, "[HATO API] Response: MAL ID = %d", malID)

	// Save complete response to cache
	if c.cache != nil {
		c.cache.Set("anilist", mediaType, anilistID, resp.Data)
	}

	return malID, true, nil
}

// SaveCache persists the cache to disk if there are unsaved changes.
func (c *HatoClient) SaveCache(ctx context.Context) error {
	if c.cache == nil {
		return nil
	}
	return c.cache.Save(ctx)
}

func (c *HatoClient) doRequest(ctx context.Context, url string) (*HatoResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// CRITICAL: Hato API requires User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	LogDebug(ctx, "[HATO API] Status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil //nolint:nilnil // nil means "not found", not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var hatoResp HatoResponse
	if err := json.NewDecoder(resp.Body).Decode(&hatoResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &hatoResp, nil
}
