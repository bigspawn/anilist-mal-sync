// Package main implements a synchronization tool for AniList and MyAnimeList accounts
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rl404/verniy"
	"golang.org/x/oauth2"
)

type AnilistClient struct {
	c           *verniy.Client
	username    string
	httpTimeout time.Duration
}

func NewAnilistClient(ctx context.Context, oauth *OAuth, username string, httpTimeout time.Duration, verbose bool) *AnilistClient {
	httpClient := oauth2.NewClient(ctx, oauth.TokenSource(ctx))
	httpClient.Transport = NewRetryableTransport(httpClient, 3)
	httpClient.Transport = newLoggingRoundTripper(httpClient.Transport, verbose)

	v := verniy.New()
	v.Http = *httpClient

	return &AnilistClient{c: v, username: username, httpTimeout: httpTimeout}
}

func (c *AnilistClient) GetUserAnimeList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	mediaListGroups, err := c.c.GetUserAnimeListWithContext(ctx, c.username,
		verniy.MediaListGroupFieldStatus,
		verniy.MediaListGroupFieldEntries(
			verniy.MediaListFieldID,
			verniy.MediaListFieldStatus,
			verniy.MediaListFieldScore,
			verniy.MediaListFieldProgress,
			verniy.MediaListFieldStartedAt,
			verniy.MediaListFieldCompletedAt,
			verniy.MediaListFieldMedia(
				verniy.MediaFieldID,
				verniy.MediaFieldIDMAL,
				verniy.MediaFieldTitle(
					verniy.MediaTitleFieldRomaji,
					verniy.MediaTitleFieldEnglish,
					verniy.MediaTitleFieldNative,
				),
				verniy.MediaFieldStatusV2,
				verniy.MediaFieldEpisodes,
				verniy.MediaFieldSeasonYear,
				verniy.MediaFieldIsFavourite,
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user anime list: %w", err)
	}
	return mediaListGroups, nil
}

func (c *AnilistClient) GetUserMangaList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	mediaListGroups, err := c.c.GetUserMangaListWithContext(ctx, c.username,
		verniy.MediaListGroupFieldName,
		verniy.MediaListGroupFieldStatus,
		verniy.MediaListGroupFieldEntries(
			verniy.MediaListFieldID,
			verniy.MediaListFieldStatus,
			verniy.MediaListFieldScore,
			verniy.MediaListFieldProgress,
			verniy.MediaListFieldProgressVolumes,
			verniy.MediaListFieldStartedAt,
			verniy.MediaListFieldCompletedAt,
			verniy.MediaListFieldMedia(
				verniy.MediaFieldID,
				verniy.MediaFieldIDMAL,
				verniy.MediaFieldTitle(
					verniy.MediaTitleFieldRomaji,
					verniy.MediaTitleFieldEnglish,
					verniy.MediaTitleFieldNative),
				verniy.MediaFieldType,
				verniy.MediaFieldFormat,
				verniy.MediaFieldStatusV2,
				verniy.MediaFieldChapters,
				verniy.MediaFieldVolumes,
				verniy.MediaFieldIsFavourite,
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user manga list: %w", err)
	}
	return mediaListGroups, nil
}

// newAnilistOAuth creates AniList OAuth with optional initialization
func newAnilistOAuth(ctx context.Context, config Config, initWithToken bool) (*OAuth, error) {
	verifier := oauth2.GenerateVerifier()

	oauthAnilist, err := NewOAuth(
		config.Anilist,
		config.OAuth.RedirectURI,
		"anilist",
		[]oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
			oauth2.S256ChallengeOption(verifier),
			oauth2.VerifierOption(verifier),
		},
		config.TokenFilePath,
	)
	if err != nil {
		return nil, err
	}

	return initOAuthIfNeeded(ctx, oauthAnilist, config.OAuth.Port, initWithToken)
}

func NewAnilistOAuth(ctx context.Context, config Config) (*OAuth, error) {
	return newAnilistOAuth(ctx, config, true)
}

// NewAnilistOAuthWithoutInit creates AniList OAuth without starting auth flow.
// Use InitToken() to manually trigger authentication when needed.
func NewAnilistOAuthWithoutInit(ctx context.Context, config Config) (*OAuth, error) {
	return newAnilistOAuth(ctx, config, false)
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message   string `json:"message"`
	Status    int    `json:"status"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
}

// GraphQLResponse represents a GraphQL response with potential errors
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// SaveMediaListEntry represents the response from AniList SaveMediaListEntry mutation
type SaveMediaListEntry struct {
	Data struct {
		SaveMediaListEntry struct {
			ID       int     `json:"id"`
			Status   string  `json:"status"`
			Progress int     `json:"progress"`
			Score    float64 `json:"score"`
		} `json:"SaveMediaListEntry"`
	} `json:"data"`
}

// UpdateAnimeEntry updates an anime entry in AniList using GraphQL mutation.
//
// Date handling:
//
//	startedAt   | completedAt | Mutation behavior
//	------------|-------------|------------------------------------------
//	nil         | nil         | dates omitted, AniList keeps existing dates
//	set         | nil         | only startedAt sent, completedAt unchanged
//	nil         | set         | only completedAt sent, startedAt unchanged
//	set         | set         | both dates sent and updated on AniList
//
// Note: nil dates are omitted (not sent as null) to avoid clearing
// manually-set dates on AniList.
func (c *AnilistClient) UpdateAnimeEntry(
	ctx context.Context, mediaID int, status string, progress int, score int,
	startedAt *time.Time, completedAt *time.Time,
) error {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	mutation := `
		mutation (
			$mediaId: Int,
			$status: MediaListStatus,
			$progress: Int,
			$score: Float,
			$startedAt: FuzzyDateInput,
			$completedAt: FuzzyDateInput
		) {
			SaveMediaListEntry(
				mediaId: $mediaId,
				status: $status,
				progress: $progress,
				score: $score,
				startedAt: $startedAt,
				completedAt: $completedAt
			) {
				id
				status
				progress
				score
			}
		}
	`

	variables := map[string]interface{}{
		"mediaId":  mediaID,
		"status":   status,
		"progress": progress,
		"score":    float64(score),
	}

	if fd := timeToFuzzyDateInput(startedAt); fd != nil {
		variables["startedAt"] = fd
	}
	if fd := timeToFuzzyDateInput(completedAt); fd != nil {
		variables["completedAt"] = fd
	}

	requestBody := map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, code, err := c.c.MakeRequest(ctx, jsonBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	if code != 200 {
		return fmt.Errorf(
			"AniList API returned status code %d: %s",
			code, string(responseBody),
		)
	}

	// Check for GraphQL errors first
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(responseBody, &graphqlResp); err != nil {
		return fmt.Errorf(
			"failed to unmarshal GraphQL response: %w", err,
		)
	}

	if len(graphqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %+v", graphqlResp.Errors)
	}

	var response SaveMediaListEntry
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// UpdateMangaEntry updates a manga entry in AniList using GraphQL mutation.
//
// Date handling follows the same rules as UpdateAnimeEntry — see its doc comment.
func (c *AnilistClient) UpdateMangaEntry(
	ctx context.Context,
	mediaID int,
	status string,
	progress int,
	progressVolumes int,
	score int,
	startedAt *time.Time,
	completedAt *time.Time,
) error {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	mutation := `
		mutation (
			$mediaId: Int,
			$status: MediaListStatus,
			$progress: Int,
			$progressVolumes: Int,
			$score: Float,
			$startedAt: FuzzyDateInput,
			$completedAt: FuzzyDateInput
		) {
			SaveMediaListEntry(
				mediaId: $mediaId,
				status: $status,
				progress: $progress,
				progressVolumes: $progressVolumes,
				score: $score,
				startedAt: $startedAt,
				completedAt: $completedAt
			) {
				id
				status
				progress
				progressVolumes
				score
			}
		}
	`

	variables := map[string]interface{}{
		"mediaId":         mediaID,
		"status":          status,
		"progress":        progress,
		"progressVolumes": progressVolumes,
		"score":           float64(score),
	}

	if fd := timeToFuzzyDateInput(startedAt); fd != nil {
		variables["startedAt"] = fd
	}
	if fd := timeToFuzzyDateInput(completedAt); fd != nil {
		variables["completedAt"] = fd
	}

	requestBody := map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, code, err := c.c.MakeRequest(ctx, jsonBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	if code != 200 {
		return fmt.Errorf("AniList API returned status code %d: %s", code, string(responseBody))
	}

	// Check for GraphQL errors first
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(responseBody, &graphqlResp); err != nil {
		return fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %+v", graphqlResp.Errors)
	}

	var response SaveMediaListEntry
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// ToggleFavouriteResponse represents the response from AniList ToggleFavourite mutation.
// The response contains nested "nodes" arrays matching the GraphQL query structure:
// anime { nodes { id } }, manga { nodes { id } }.
type ToggleFavouriteResponse struct {
	Data struct {
		ToggleFavourite struct {
			Anime struct {
				Nodes []struct {
					ID int `json:"id"`
				} `json:"nodes"`
			} `json:"anime"`
			Manga struct {
				Nodes []struct {
					ID int `json:"id"`
				} `json:"nodes"`
			} `json:"manga"`
		} `json:"ToggleFavourite"`
	} `json:"data"`
}

// ToggleFavourite toggles the favorite status of an anime or manga on AniList.
// Only one of animeID or mangaID should be provided (non-zero) per call.
// The ToggleFavourite mutation is idempotent: calling it on an already-favorited
// item removes it from favorites, calling it on a non-favorited item adds it.
//
// Parameters:
//   - animeID: AniList anime ID (use 0 if toggling manga)
//   - mangaID: AniList manga ID (use 0 if toggling anime)
func (c *AnilistClient) ToggleFavourite(ctx context.Context, animeID, mangaID int) error {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()

	if animeID <= 0 && mangaID <= 0 {
		return fmt.Errorf("at least one of animeID or mangaID must be positive")
	}
	if animeID > 0 && mangaID > 0 {
		return fmt.Errorf("only one of animeID or mangaID can be specified per call")
	}

	mutation := `
		mutation ($animeId: Int, $mangaId: Int) {
			ToggleFavourite(animeId: $animeId, mangaId: $mangaId) {
				anime { nodes { id } }
				manga { nodes { id } }
			}
		}
	`

	variables := map[string]any{}
	if animeID > 0 {
		variables["animeId"] = animeID
	}
	if mangaID > 0 {
		variables["mangaId"] = mangaID
	}

	requestBody := map[string]any{
		"query":     mutation,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, code, err := c.c.MakeRequest(ctx, jsonBody)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	if code != 200 {
		return fmt.Errorf("AniList API returned status code %d: %s", code, string(responseBody))
	}

	// Check for GraphQL errors
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(responseBody, &graphqlResp); err != nil {
		return fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %+v", graphqlResp.Errors)
	}

	var response ToggleFavouriteResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetAnimeByID gets an anime from AniList by ID
func (c *AnilistClient) GetAnimeByID(ctx context.Context, id int) (*verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	media, err := c.c.GetAnimeWithContext(ctx, id,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldEpisodes,
		verniy.MediaFieldSeasonYear,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime by ID %d: %w", id, err)
	}
	return media, nil
}

// GetAnimesByName searches for anime on AniList by name
func (c *AnilistClient) GetAnimesByName(ctx context.Context, name string) ([]verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	page, err := c.c.SearchAnimeWithContext(ctx, verniy.PageParamMedia{Search: name}, 1, 10,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldEpisodes,
		verniy.MediaFieldSeasonYear,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime by name: %w", err)
	}
	return page.Media, nil
}

// GetAnimeByMALID gets an anime from AniList by MAL ID
func (c *AnilistClient) GetAnimeByMALID(ctx context.Context, malID int) (*verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	page, err := c.c.SearchAnimeWithContext(ctx, verniy.PageParamMedia{IDMAL: malID}, 1, 1,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldEpisodes,
		verniy.MediaFieldSeasonYear,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime by MAL ID %d: %w", malID, err)
	}
	if len(page.Media) == 0 {
		return nil, fmt.Errorf("no anime found with MAL ID %d", malID)
	}
	return &page.Media[0], nil
}

// GetMangaByID gets a manga from AniList by ID
func (c *AnilistClient) GetMangaByID(ctx context.Context, id int) (*verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	media, err := c.c.GetMangaWithContext(ctx, id,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldType,
		verniy.MediaFieldFormat,
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldChapters,
		verniy.MediaFieldVolumes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get manga by ID %d: %w", id, err)
	}
	return media, nil
}

// GetMangasByName searches for manga on AniList by name
func (c *AnilistClient) GetMangasByName(ctx context.Context, name string) ([]verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	page, err := c.c.SearchMangaWithContext(ctx, verniy.PageParamMedia{Search: name}, 1, 10,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldType,
		verniy.MediaFieldFormat,
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldChapters,
		verniy.MediaFieldVolumes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search manga by name: %w", err)
	}
	return page.Media, nil
}

// GetMangaByMALID gets a manga from AniList by MAL ID
func (c *AnilistClient) GetMangaByMALID(ctx context.Context, malID int) (*verniy.Media, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	page, err := c.c.SearchMangaWithContext(ctx, verniy.PageParamMedia{IDMAL: malID}, 1, 1,
		verniy.MediaFieldID,
		verniy.MediaFieldIDMAL,
		verniy.MediaFieldTitle(
			verniy.MediaTitleFieldRomaji,
			verniy.MediaTitleFieldEnglish,
			verniy.MediaTitleFieldNative,
		),
		verniy.MediaFieldType,
		verniy.MediaFieldFormat,
		verniy.MediaFieldStatusV2,
		verniy.MediaFieldChapters,
		verniy.MediaFieldVolumes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search manga by MAL ID %d: %w", malID, err)
	}
	if len(page.Media) == 0 {
		return nil, fmt.Errorf("no manga found with MAL ID %d", malID)
	}
	return &page.Media[0], nil
}

// timeToFuzzyDateInput converts a *time.Time to an AniList FuzzyDateInput map.
//
// Behavior:
//
//	Input   | Output
//	--------|----------------------------------
//	nil     | nil (date omitted from mutation)
//	set     | map[year, month, day]
func timeToFuzzyDateInput(t *time.Time) map[string]int {
	if t == nil {
		return nil
	}
	return map[string]int{
		"year":  t.Year(),
		"month": int(t.Month()),
		"day":   t.Day(),
	}
}

// GetUserScoreFormat retrieves the user's score format preference from AniList
func (c *AnilistClient) GetUserScoreFormat(ctx context.Context) (verniy.ScoreFormat, error) {
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	user, err := c.c.GetUserWithContext(ctx, c.username,
		verniy.UserFieldMediaListOptions(
			verniy.MediaListOptionsFieldScoreFormat,
		),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get user score format: %w", err)
	}

	if user.MediaListOptions == nil {
		return "", fmt.Errorf("user media list options is nil")
	}

	if user.MediaListOptions.ScoreFormat == nil {
		return "", fmt.Errorf("user score format is nil")
	}

	return *user.MediaListOptions.ScoreFormat, nil
}
