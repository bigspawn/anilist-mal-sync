// Package main implements a synchronization tool for AniList and MyAnimeList accounts
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rl404/verniy"
	"golang.org/x/oauth2"
)

type AnilistClient struct {
	c *verniy.Client

	username string
}

func NewAnilistClient(ctx context.Context, oauth *OAuth, username string) *AnilistClient {
	httpClient := oauth2.NewClient(ctx, oauth.TokenSource())
	httpClient.Timeout = 10 * time.Minute

	v := verniy.New()
	v.Http = *httpClient

	return &AnilistClient{c: v, username: username}
}

func (c *AnilistClient) GetUserAnimeList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	var result []verniy.MediaListGroup

	err := retryWithBackoff(ctx, func() error {
		mediaListGroups, e := c.c.GetUserAnimeListWithContext(ctx, c.username,
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
				),
			),
		)
		if e != nil {
			return fmt.Errorf("failed to get user anime list: %w", e)
		}
		result = mediaListGroups
		return nil
	}, fmt.Sprintf("AniList get user anime list: %s", c.username))

	return result, err
}

func (c *AnilistClient) GetUserMangaList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	var result []verniy.MediaListGroup

	err := retryWithBackoff(ctx, func() error {
		mediaListGroups, e := c.c.GetUserMangaListWithContext(ctx, c.username,
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
				),
			),
		)
		if e != nil {
			return fmt.Errorf("failed to get user manga list: %w", e)
		}
		result = mediaListGroups
		return nil
	}, fmt.Sprintf("AniList get user manga list: %s", c.username))

	return result, err
}

func NewAnilistOAuth(ctx context.Context, config Config) (*OAuth, error) {
	// Generate PKCE code verifier using oauth2 package
	verifier := oauth2.GenerateVerifier()

	oauthAnilist, err := NewOAuth(
		config.Anilist,
		config.OAuth.RedirectURI,
		"anilist",
		[]oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
			oauth2.S256ChallengeOption(verifier), // S256 challenge for auth URL
			oauth2.VerifierOption(verifier),      // Verifier for token exchange
		},
		config.TokenFilePath,
	)
	if err != nil {
		return nil, err
	}

	if oauthAnilist.NeedInit() {
		getToken(ctx, oauthAnilist, config.OAuth.Port)
	} else {
		log.Println("Token already set, no need to start server")
	}

	return oauthAnilist, nil
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

// UpdateAnimeEntry updates an anime entry in AniList using GraphQL mutation with retry logic
func (c *AnilistClient) UpdateAnimeEntry(
	ctx context.Context, mediaID int, status string, progress int, score int, prefix string,
) error {
	return retryWithBackoff(ctx, func() error {
		mutation := `
			mutation ($mediaId: Int, $status: MediaListStatus, $progress: Int, $score: Float) {
				SaveMediaListEntry(mediaId: $mediaId, status: $status, progress: $progress, score: $score) {
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
	}, fmt.Sprintf("AniList update anime entry: %d", mediaID), prefix)
}

// UpdateMangaEntry updates a manga entry in AniList using GraphQL mutation with retry logic
func (c *AnilistClient) UpdateMangaEntry(
	ctx context.Context,
	mediaID int,
	status string,
	progress int,
	progressVolumes int,
	score int,
	prefix string,
) error {
	return retryWithBackoff(ctx, func() error {
		mutation := `
			mutation ($mediaId: Int, $status: MediaListStatus, $progress: Int, $progressVolumes: Int, $score: Float) {
				SaveMediaListEntry(mediaId: $mediaId, status: $status, progress: $progress, progressVolumes: $progressVolumes, score: $score) {
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
	}, fmt.Sprintf("AniList update manga entry: %d", mediaID), prefix)
}

// GetAnimeByID gets an anime from AniList by ID with retry logic
func (c *AnilistClient) GetAnimeByID(ctx context.Context, id int, prefix string) (*verniy.Media, error) {
	var result *verniy.Media

	err := retryWithBackoff(ctx, func() error {
		media, e := c.c.GetAnimeWithContext(ctx, id,
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
		if e != nil {
			return fmt.Errorf("failed to get anime by ID %d: %w", id, e)
		}
		result = media
		return nil
	}, fmt.Sprintf("AniList get anime by ID: %d", id), prefix)

	return result, err
}

// GetAnimesByName searches for anime on AniList by name with retry logic
func (c *AnilistClient) GetAnimesByName(ctx context.Context, name string, prefix string) ([]verniy.Media, error) {
	var result []verniy.Media

	err := retryWithBackoff(ctx, func() error {
		page, e := c.c.SearchAnimeWithContext(ctx, verniy.PageParamMedia{Search: name}, 1, 10,
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
		if e != nil {
			return fmt.Errorf("failed to search anime by name: %w", e)
		}
		result = page.Media
		return nil
	}, fmt.Sprintf("AniList search anime by name: %s", name), prefix)

	return result, err
}

// GetMangaByID gets a manga from AniList by ID with retry logic
func (c *AnilistClient) GetMangaByID(ctx context.Context, id int, prefix string) (*verniy.Media, error) {
	var result *verniy.Media

	err := retryWithBackoff(ctx, func() error {
		media, e := c.c.GetMangaWithContext(ctx, id,
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
		if e != nil {
			return fmt.Errorf("failed to get manga by ID %d: %w", id, e)
		}
		result = media
		return nil
	}, fmt.Sprintf("AniList get manga by ID: %d", id), prefix)
	return result, err
}

// GetMangasByName searches for manga on AniList by name with retry logic
func (c *AnilistClient) GetMangasByName(ctx context.Context, name string, prefix string) ([]verniy.Media, error) {
	var result []verniy.Media

	err := retryWithBackoff(ctx, func() error {
		page, e := c.c.SearchMangaWithContext(ctx, verniy.PageParamMedia{Search: name}, 1, 10,
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
		if e != nil {
			return fmt.Errorf("failed to search manga by name: %w", e)
		}
		result = page.Media
		return nil
	}, fmt.Sprintf("AniList search manga by name: %s", name), prefix)

	return result, err
}

// createBackoffPolicy creates a configured exponential backoff policy for rate limiting
func createBackoffPolicy() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 2 * time.Minute
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.1 // Add jitter
	return b
}

// isRateLimitError checks if the error is a rate limiting error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429")
}

// retryWithBackoff wraps an operation with exponential backoff for rate limit handling
func retryWithBackoff(ctx context.Context, operation func() error, operationName string, prefix ...string) error {
	b := createBackoffPolicy()

	retryableOperation := func() error {
		err := operation()
		if err != nil && !isRateLimitError(err) {
			// Don't retry non-rate-limit errors
			return backoff.Permanent(err)
		}
		return err
	}

	return backoff.RetryNotify(
		retryableOperation,
		backoff.WithContext(b, ctx),
		func(err error, duration time.Duration) {
			if isRateLimitError(err) {
				if len(prefix) > 0 {
					log.Printf("[%s] Rate limit hit for %s, retrying in %v: %v", prefix[0], operationName, duration, err)
				} else {
					log.Printf("Rate limit hit for %s, retrying in %v: %v", operationName, duration, err)
				}
			}
		},
	)
}

// GetUserScoreFormat retrieves the user's score format preference from AniList
func (c *AnilistClient) GetUserScoreFormat(ctx context.Context) (verniy.ScoreFormat, error) {
	var result *verniy.User

	err := retryWithBackoff(ctx, func() error {
		user, e := c.c.GetUserWithContext(ctx, c.username,
			verniy.UserFieldMediaListOptions(
				verniy.MediaListOptionsFieldScoreFormat,
			),
		)
		if e != nil {
			return fmt.Errorf("failed to get user score format: %w", e)
		}
		result = user
		return nil
	}, fmt.Sprintf("AniList get user score format: %s", c.username))
	if err != nil {
		return "", err
	}

	if result.MediaListOptions == nil {
		return "", fmt.Errorf("user media list options is nil")
	}

	if result.MediaListOptions.ScoreFormat == nil {
		return "", fmt.Errorf("user score format is nil")
	}

	return *result.MediaListOptions.ScoreFormat, nil
}
