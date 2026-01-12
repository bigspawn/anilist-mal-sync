package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"golang.org/x/oauth2"
)

var errEmptyMalID = errors.New("mal id is empty")

var animeFields = mal.Fields{
	"alternative_titles",
	"num_episodes",
	"my_list_status",
	"start_season",
}

var mangaFields = mal.Fields{
	"alternative_titles",
	"num_volumes",
	"num_chapters",
	"my_list_status",
	"start_date",
}

type MyAnimeListClient struct {
	c           *mal.Client
	username    string
	httpTimeout time.Duration
}

func NewMyAnimeListClient(ctx context.Context, oauth *OAuth, username string, httpTimeout time.Duration) *MyAnimeListClient {
	httpClient := oauth2.NewClient(ctx, oauth.TokenSource(ctx))
	httpClient.Transport = newLoggingRoundTripper(httpClient.Transport)

	client := mal.NewClient(httpClient)

	return &MyAnimeListClient{c: client, username: username, httpTimeout: httpTimeout}
}

func (c *MyAnimeListClient) GetUserAnimeList(ctx context.Context) ([]mal.UserAnime, error) {
	return fetchAllPages(
		ctx,
		"MAL get user anime list",
		c.httpTimeout,
		func(ctx context.Context, offset int) ([]mal.UserAnime, *mal.Response, error) {
			return c.c.User.AnimeList(ctx, c.username, animeFields, mal.Offset(offset), mal.Limit(100))
		})
}

func (c *MyAnimeListClient) GetAnimesByName(ctx context.Context, name string) ([]mal.Anime, error) {
	var result []mal.Anime
	err := retryWithBackoff(ctx, func() error {
		ctx, cancel := withTimeout(ctx, c.httpTimeout)
		defer cancel()
		list, _, e := c.c.Anime.List(ctx, name, animeFields, mal.Limit(3))
		if e != nil {
			return e
		}
		result = list
		return nil
	}, fmt.Sprintf("MAL search anime by name: %s", name))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *MyAnimeListClient) GetAnimeByID(ctx context.Context, id int) (*mal.Anime, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	var result *mal.Anime
	err := retryWithBackoff(ctx, func() error {
		ctx, cancel := withTimeout(ctx, c.httpTimeout)
		defer cancel()
		anime, _, e := c.c.Anime.Details(ctx, id, animeFields)
		if e != nil {
			return e
		}
		result = anime
		return nil
	}, fmt.Sprintf("MAL get anime by ID: %d", id))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *MyAnimeListClient) UpdateAnimeByIDAndOptions(ctx context.Context, id int, opts []mal.UpdateMyAnimeListStatusOption) error {
	if len(opts) == 0 {
		return nil
	}

	// Log update details for debugging
	DPrintf("[DEBUG] Updating MAL ID %d with opts: %+v", id, opts)

	return retryWithBackoff(ctx, func() error {
		_, _, err := c.c.Anime.UpdateMyListStatus(ctx, id, opts...)
		if err != nil {
			return fmt.Errorf("failed to update anime %d: %w", id, err)
		}
		return nil
	}, fmt.Sprintf("MAL update anime: %d", id), "AniList to MAL Anime")
}

func (c *MyAnimeListClient) GetUserMangaList(ctx context.Context) ([]mal.UserManga, error) {
	return fetchAllPages(
		ctx,
		"MAL get user manga list",
		c.httpTimeout,
		func(ctx context.Context, offset int) ([]mal.UserManga, *mal.Response, error) {
			return c.c.User.MangaList(ctx, c.username, mangaFields, mal.Offset(offset), mal.Limit(100))
		})
}

func (c *MyAnimeListClient) GetMangasByName(ctx context.Context, name string) ([]mal.Manga, error) {
	var result []mal.Manga
	err := retryWithBackoff(ctx, func() error {
		ctx, cancel := withTimeout(ctx, c.httpTimeout)
		defer cancel()
		l, _, e := c.c.Manga.List(ctx, name, mangaFields, mal.Limit(10))
		if e != nil {
			return e
		}
		result = l
		return nil
	}, fmt.Sprintf("MAL search manga by name: %s", name))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *MyAnimeListClient) GetMangaByID(ctx context.Context, id int) (*mal.Manga, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	var result *mal.Manga
	err := retryWithBackoff(ctx, func() error {
		ctx, cancel := withTimeout(ctx, c.httpTimeout)
		defer cancel()
		m, _, e := c.c.Manga.Details(ctx, id, mangaFields)
		if e != nil {
			return e
		}
		result = m
		return nil
	}, fmt.Sprintf("MAL get manga by ID: %d", id))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *MyAnimeListClient) UpdateMangaByIDAndOptions(ctx context.Context, id int, opts []mal.UpdateMyMangaListStatusOption) error {
	if len(opts) == 0 {
		return nil
	}

	return retryWithBackoff(ctx, func() error {
		_, _, err := c.c.Manga.UpdateMyListStatus(ctx, id, opts...)
		if err != nil {
			return fmt.Errorf("failed to update manga %d: %w", id, err)
		}
		return nil
	}, fmt.Sprintf("MAL update manga: %d", id), "AniList to MAL Manga")
}

func NewMyAnimeListOAuth(ctx context.Context, config Config) (*OAuth, error) {
	// Generate PKCE code verifier using oauth2 package
	verifier := oauth2.GenerateVerifier()

	oauthMAL, err := NewOAuth(
		config.MyAnimeList,
		config.OAuth.RedirectURI,
		"myanimelist",
		[]oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", verifier),       // Plain challenge (same as verifier)
			oauth2.SetAuthURLParam("code_challenge_method", "plain"), // Explicit plain method
			oauth2.VerifierOption(verifier),                          // Verifier for token exchange
		},
		config.TokenFilePath,
	)
	if err != nil {
		return nil, err
	}

	if oauthMAL.NeedInit() {
		getToken(ctx, oauthMAL, config.OAuth.Port)
		// Check if context was cancelled during OAuth flow
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	} else {
		log.Println("Token already set, no need to start server")
	}

	return oauthMAL, nil
}

// NewMyAnimeListOAuthWithoutInit creates MAL OAuth without starting auth flow.
// Use InitToken() to manually trigger authentication when needed.
func NewMyAnimeListOAuthWithoutInit(config Config) (*OAuth, error) {
	verifier := oauth2.GenerateVerifier()

	return NewOAuth(
		config.MyAnimeList,
		config.OAuth.RedirectURI,
		"myanimelist",
		[]oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", verifier),
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
			oauth2.VerifierOption(verifier),
		},
		config.TokenFilePath,
	)
}

// fetchAllPages fetches all pages from a paginated MAL API endpoint using retry logic with timeout.
// The fetch function should return a page of items along with the pagination response.
// Verbose logs are printed to show pagination progress.
func fetchAllPages[T any](
	ctx context.Context,
	operationName string,
	timeout time.Duration,
	fetch func(ctx context.Context, offset int) ([]T, *mal.Response, error),
) ([]T, error) {
	var result []T
	offset := 0
	pageNum := 1

	for {
		var items []T
		var resp *mal.Response

		err := retryWithBackoff(ctx, func() error {
			ctx, cancel := withTimeout(ctx, timeout)
			defer cancel()
			var e error
			items, resp, e = fetch(ctx, offset)
			return e
		}, fmt.Sprintf("%s (page %d, offset: %d)", operationName, pageNum, offset))
		if err != nil {
			return nil, err
		}

		DPrintf("[DEBUG] %s: fetched page %d with %d items (next offset: %d)", operationName, pageNum, len(items), resp.NextOffset)

		result = append(result, items...)

		if resp.NextOffset == 0 {
			DPrintf("[DEBUG] %s: finished pagination, total %d items across %d page(s)", operationName, len(result), pageNum)
			break
		}

		offset = resp.NextOffset
		pageNum++
	}

	return result, nil
}
