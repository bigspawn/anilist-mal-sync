package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rl404/verniy"
	"golang.org/x/oauth2"
)

type AnilistClient struct {
	c *verniy.Client

	username string
}

func NewAnilistClient(ctx context.Context, oauth *OAuth, username string) (*AnilistClient, error) {
	if oauth.GetToken() == nil {
		return nil, errors.New("token is nil")
	}

	httpClient := oauth.Config.Client(ctx, oauth.GetToken())
	httpClient.Timeout = 10 * time.Minute

	v := verniy.New()
	v.Http = *httpClient

	return &AnilistClient{c: v, username: username}, nil
}

func (c *AnilistClient) GetUserAnimeList(ctx context.Context) ([]verniy.MediaListGroup, error) {
	lists, err := c.c.GetUserAnimeListWithContext(ctx, c.username,
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
	if err != nil {
		return nil, err
	}

	return lists, nil
}

func NewAnilistOAuth(ctx context.Context, config Config) (*OAuth, error) {
	oauthAnilist, err := NewOAuth(
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

	if oauthAnilist.GetToken() != nil {
		log.Println("Token already set, no need to start server")
		return oauthAnilist, nil
	}

	getToken(ctx, oauthAnilist, config.OAuth.Port)

	return oauthAnilist, nil
}
