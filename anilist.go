package main

import (
	"context"
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
