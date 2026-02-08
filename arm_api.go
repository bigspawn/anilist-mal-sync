package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultARMBaseURL = "https://arm.haglund.dev"

// ARMClient is an HTTP client for the ARM API (https://arm.haglund.dev).
type ARMClient struct {
	baseURL    string
	httpClient HTTPClient
}

// ARMResponse represents the response from /api/v2/ids.
type ARMResponse struct {
	AniList     *int `json:"anilist"`
	MyAnimeList *int `json:"myanimelist"`
	AniDB       *int `json:"anidb"`
	Kitsu       *int `json:"kitsu"`
}

// NewARMClient creates a new ARM API client.
func NewARMClient(baseURL string, timeout time.Duration) *ARMClient {
	if baseURL == "" {
		baseURL = defaultARMBaseURL
	}
	return &ARMClient{
		baseURL: baseURL,
		httpClient: NewRetryableClient(&http.Client{
			Timeout: timeout,
		}, 3),
	}
}

// GetAniListID returns the AniList ID for a given MAL ID.
func (c *ARMClient) GetAniListID(ctx context.Context, malID int) (int, bool, error) {
	url := fmt.Sprintf("%s/api/v2/ids?source=myanimelist&id=%d&include=anilist", c.baseURL, malID)
	LogDebug(ctx, "[ARM API] GET %s", url)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		LogDebug(ctx, "[ARM API] Error: %v", err)
		return 0, false, err
	}
	if resp == nil || resp.AniList == nil {
		LogDebug(ctx, "[ARM API] Response: not found (404 or null)")
		return 0, false, nil
	}
	LogDebug(ctx, "[ARM API] Response: AniList ID = %d", *resp.AniList)
	return *resp.AniList, true, nil
}

// GetMALID returns the MAL ID for a given AniList ID.
func (c *ARMClient) GetMALID(ctx context.Context, anilistID int) (int, bool, error) {
	url := fmt.Sprintf("%s/api/v2/ids?source=anilist&id=%d&include=myanimelist", c.baseURL, anilistID)
	LogDebug(ctx, "[ARM API] GET %s", url)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		LogDebug(ctx, "[ARM API] Error: %v", err)
		return 0, false, err
	}
	if resp == nil || resp.MyAnimeList == nil {
		LogDebug(ctx, "[ARM API] Response: not found (404 or null)")
		return 0, false, nil
	}
	LogDebug(ctx, "[ARM API] Response: MAL ID = %d", *resp.MyAnimeList)
	return *resp.MyAnimeList, true, nil
}

func (c *ARMClient) doRequest(ctx context.Context, url string) (*ARMResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	LogDebug(ctx, "[ARM API] Status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil //nolint:nilnil // nil means "not found", not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var armResp ARMResponse
	if err := json.NewDecoder(resp.Body).Decode(&armResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &armResp, nil
}
