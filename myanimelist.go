package main

import (
	"context"
	"log"
	"net/url"

	malpkg "github.com/bigspawn/anilist-mal-sync/internal/mal"
	"golang.org/x/oauth2"
)

// Re-export MyAnimeList client type and constructor
type MyAnimeListClient = malpkg.MyAnimeListClient

func NewMyAnimeListClient(ctx context.Context, oauth *OAuth, username string) *MyAnimeListClient {
	return malpkg.NewMyAnimeListClient(ctx, oauth, username)
}

const randNumb = 43

func NewMyAnimeListOAuth(ctx context.Context, config Config) (*OAuth, error) {
	code := url.QueryEscape(randHTTPParamString(randNumb))

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

	if oauthMAL.NeedInit() {
		getToken(ctx, oauthMAL, config.OAuth.Port)
	} else {
		log.Println("Token already set, no need to start server")
	}

	return oauthMAL, nil
}
