package main

import (
	"context"
	"errors"
	"log"
	"net/url"
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

type MyAnimeListClient struct {
	c *mal.Client

	username string
}

func NewMyAnimeListClient(ctx context.Context, oauth *OAuth, username string) (*MyAnimeListClient, error) {
	token := oauth.GetToken()
	if token == nil {
		return nil, errors.New("token is nil")
	}

	httpClient := oauth.Config.Client(ctx, token)
	httpClient.Timeout = 10 * time.Minute

	client := mal.NewClient(httpClient)

	return &MyAnimeListClient{c: client, username: username}, nil
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

func (c *MyAnimeListClient) UpdateAnime(ctx context.Context, anime Anime) error {
	if anime.IDMal <= 0 {
		return errEmptyMalID
	}

	st, err := anime.Status.GetMalStatus()
	if err != nil {
		return err
	}

	opts := []mal.UpdateMyAnimeListStatusOption{
		st,
		mal.Score(anime.Score),
		mal.NumEpisodesWatched(anime.Progress),
	}

	if anime.StartedAt != nil {
		opts = append(opts, mal.StartDate(*anime.StartedAt))
	} else {
		opts = append(opts, mal.StartDate(time.Time{}))
	}

	if anime.Status == StatusCompleted && anime.FinishedAt != nil {
		opts = append(opts, mal.FinishDate(*anime.FinishedAt))
	} else {
		opts = append(opts, mal.FinishDate(time.Time{}))
	}

	_, _, err = c.c.Anime.UpdateMyListStatus(
		ctx,
		anime.IDMal,
		opts...,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewMyAnimeListOAuth(ctx context.Context, config Config) (*OAuth, error) {
	code := url.QueryEscape(randHttpParamString(43))

	oauthMAL, err := NewOAuth(
		config.MyAnimeList,
		config.OAuth.RedirectURI,
		"myanimelist",
		[]oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", code),
			oauth2.SetAuthURLParam("code_verifier", code),
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
		},
		config.TokenFilePath,
	)
	if err != nil {
		return nil, err
	}

	if oauthMAL.GetToken() != nil {
		log.Println("Token already set, no need to start server")
		return oauthMAL, nil
	}

	getToken(ctx, oauthMAL, config.OAuth.Port)

	return oauthMAL, nil
}
