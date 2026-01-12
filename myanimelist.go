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
	c *mal.Client

	username string
}

func NewMyAnimeListClient(ctx context.Context, oauth *OAuth, username string) *MyAnimeListClient {
	httpClient := oauth2.NewClient(ctx, oauth.TokenSource(ctx))
	httpClient.Timeout = 10 * time.Minute

	client := mal.NewClient(httpClient)

	return &MyAnimeListClient{c: client, username: username}
}

func (c *MyAnimeListClient) GetUserAnimeList(ctx context.Context) ([]mal.UserAnime, error) {
	var userAnimeList []mal.UserAnime
	var offset int
	for {
		list, resp, err := c.c.User.AnimeList(ctx, c.username, animeFields, mal.Offset(offset), mal.Limit(100))
		if err != nil {
			return nil, err
		}

		userAnimeList = append(userAnimeList, list...)

		if resp.NextOffset == 0 {
			break
		}

		offset = resp.NextOffset
	}
	return userAnimeList, nil
}

func (c *MyAnimeListClient) GetAnimesByName(ctx context.Context, name string) ([]mal.Anime, error) {
	animeList, _, err := c.c.Anime.List(ctx, name, animeFields, mal.Limit(3))
	if err != nil {
		return nil, err
	}

	return animeList, nil
}

func (c *MyAnimeListClient) GetAnimeByID(ctx context.Context, id int) (*mal.Anime, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	anime, _, err := c.c.Anime.Details(ctx, id, animeFields)
	if err != nil {
		return nil, err
	}

	return anime, nil
}

func (c *MyAnimeListClient) UpdateAnimeByIDAndOptions(ctx context.Context, id int, opts []mal.UpdateMyAnimeListStatusOption) error {
	if len(opts) == 0 {
		return nil
	}

	// Log update details for debugging
	if *verbose {
		log.Printf("[DEBUG] Updating MAL ID %d with opts: %+v", id, opts)
	}

	return retryWithBackoff(ctx, func() error {
		_, _, err := c.c.Anime.UpdateMyListStatus(ctx, id, opts...)
		if err != nil {
			return fmt.Errorf("failed to update anime %d: %w", id, err)
		}
		return nil
	}, fmt.Sprintf("MAL update anime: %d", id), "AniList to MAL Anime")
}

func (c *MyAnimeListClient) GetUserMangaList(ctx context.Context) ([]mal.UserManga, error) {
	var userMangaList []mal.UserManga
	var offset int
	for {
		list, resp, err := c.c.User.MangaList(ctx, c.username, mangaFields, mal.Offset(offset), mal.Limit(100))
		if err != nil {
			return nil, err
		}

		userMangaList = append(userMangaList, list...)

		if resp.NextOffset == 0 {
			break
		}

		offset = resp.NextOffset
	}
	return userMangaList, nil
}

func (c *MyAnimeListClient) GetMangasByName(ctx context.Context, name string) ([]mal.Manga, error) {
	l, _, err := c.c.Manga.List(ctx, name, mangaFields, mal.Limit(10))
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (c *MyAnimeListClient) GetMangaByID(ctx context.Context, id int) (*mal.Manga, error) {
	if id <= 0 {
		return nil, errEmptyMalID
	}

	m, _, err := c.c.Manga.Details(ctx, id, mangaFields)
	if err != nil {
		return nil, err
	}

	return m, nil
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
