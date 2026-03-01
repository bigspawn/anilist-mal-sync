package main

import (
	"context"
	"errors"
	"fmt"
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

func NewMyAnimeListClient(ctx context.Context, oauth *OAuth, username string, httpTimeout time.Duration, verbose bool) *MyAnimeListClient {
	httpClient := oauth2.NewClient(ctx, oauth.TokenSource(ctx))
	httpClient.Transport = NewRetryableTransport(httpClient, 3)
	httpClient.Transport = newLoggingRoundTripper(httpClient.Transport, verbose)

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
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	list, _, err := c.c.Anime.List(ctx, name, animeFields, mal.Limit(3))
	return list, err
}

func (c *MyAnimeListClient) GetAnimeByID(ctx context.Context, id int) (*mal.Anime, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	anime, _, err := c.c.Anime.Details(ctx, id, animeFields)
	return anime, err
}

func (c *MyAnimeListClient) UpdateAnimeByIDAndOptions(ctx context.Context, id int, opts []mal.UpdateMyAnimeListStatusOption) error {
	if len(opts) == 0 {
		return nil
	}

	// Log update details for debugging
	LogDebug(ctx, "[MAL] Updating MAL ID %d with opts: %+v", id, opts)

	_, _, err := c.c.Anime.UpdateMyListStatus(ctx, id, opts...)
	if err != nil {
		return fmt.Errorf("failed to update anime %d: %w", id, err)
	}
	return nil
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
	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	list, _, err := c.c.Manga.List(ctx, name, mangaFields, mal.Limit(10))
	return list, err
}

func (c *MyAnimeListClient) GetMangaByID(ctx context.Context, id int) (*mal.Manga, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	ctx, cancel := withTimeout(ctx, c.httpTimeout)
	defer cancel()
	manga, _, err := c.c.Manga.Details(ctx, id, mangaFields)
	return manga, err
}

func (c *MyAnimeListClient) UpdateMangaByIDAndOptions(ctx context.Context, id int, opts []mal.UpdateMyMangaListStatusOption) error {
	if len(opts) == 0 {
		return nil
	}

	_, _, err := c.c.Manga.UpdateMyListStatus(ctx, id, opts...)
	if err != nil {
		return fmt.Errorf("failed to update manga %d: %w", id, err)
	}
	return nil
}

// newMyAnimeListOAuth creates MAL OAuth with optional initialization
func newMyAnimeListOAuth(ctx context.Context, config Config, initWithToken bool) (*OAuth, error) {
	verifier := oauth2.GenerateVerifier()

	oauthMAL, err := NewOAuth(
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
	if err != nil {
		return nil, err
	}

	return initOAuthIfNeeded(ctx, oauthMAL, config.OAuth.Port, initWithToken)
}

func NewMyAnimeListOAuth(ctx context.Context, config Config) (*OAuth, error) {
	return newMyAnimeListOAuth(ctx, config, true)
}

// NewMyAnimeListOAuthWithoutInit creates MAL OAuth without starting auth flow.
// Use InitToken() to manually trigger authentication when needed.
func NewMyAnimeListOAuthWithoutInit(ctx context.Context, config Config) (*OAuth, error) {
	return newMyAnimeListOAuth(ctx, config, false)
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

		ctx, cancel := withTimeout(ctx, timeout)
		var err error
		items, resp, err = fetch(ctx, offset)
		cancel()
		if err != nil {
			return nil, err
		}

		LogDebug(ctx, "[MAL] %s: fetched page %d with %d items (next offset: %d)", operationName, pageNum, len(items), resp.NextOffset)

		result = append(result, items...)

		if resp.NextOffset == 0 {
			LogDebug(ctx, "[MAL] %s: finished pagination, total %d items across %d page(s)", operationName, len(result), pageNum)
			break
		}

		offset = resp.NextOffset
		pageNum++
	}

	return result, nil
}
