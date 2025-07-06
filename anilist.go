package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
	return c.c.GetUserAnimeListWithContext(ctx, c.username,
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
}

func (c *AnilistClient) GetUserMangaList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	return c.c.GetUserMangaListWithContext(ctx, c.username,
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
}

func NewAnilistOAuth(ctx context.Context, config Config) (*OAuth, error) {
	oauthAnilist, err := NewOAuth(
		ctx,
		config.Anilist,
		config.OAuth.RedirectURI,
		"anilist",
		[]oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
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

// SaveMediaListEntry represents the response from AniList SaveMediaListEntry mutation
type SaveMediaListEntry struct {
	Data struct {
		SaveMediaListEntry struct {
			ID       int    `json:"id"`
			Status   string `json:"status"`
			Progress int    `json:"progress"`
			Score    int    `json:"score"`
		} `json:"SaveMediaListEntry"`
	} `json:"data"`
}

// UpdateAnimeEntry updates an anime entry in AniList using GraphQL mutation
func (c *AnilistClient) UpdateAnimeEntry(ctx context.Context, mediaID int, status string, progress int, score int) error {
	mutation := `
		mutation ($mediaId: Int, $status: MediaListStatus, $progress: Int, $score: Int) {
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
		"score":    score,
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

	var response SaveMediaListEntry
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// UpdateMangaEntry updates a manga entry in AniList using GraphQL mutation
func (c *AnilistClient) UpdateMangaEntry(ctx context.Context, mediaID int, status string, progress int, progressVolumes int, score int) error {
	mutation := `
		mutation ($mediaId: Int, $status: MediaListStatus, $progress: Int, $progressVolumes: Int, $score: Int) {
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
		"score":           score,
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

	var response SaveMediaListEntry
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetAnimeByID gets an anime from AniList by ID
func (c *AnilistClient) GetAnimeByID(ctx context.Context, id int) (*verniy.Media, error) {
	return c.c.GetAnimeWithContext(ctx, id,
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
}

// GetAnimesByName searches for anime on AniList by name
func (c *AnilistClient) GetAnimesByName(ctx context.Context, name string) ([]verniy.Media, error) {
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
		return nil, err
	}
	return page.Media, nil
}

// GetMangaByID gets a manga from AniList by ID
func (c *AnilistClient) GetMangaByID(ctx context.Context, id int) (*verniy.Media, error) {
	return c.c.GetMangaWithContext(ctx, id,
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
}

// GetMangasByName searches for manga on AniList by name
func (c *AnilistClient) GetMangasByName(ctx context.Context, name string) ([]verniy.Media, error) {
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
		return nil, err
	}
	return page.Media, nil
}
