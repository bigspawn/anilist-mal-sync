package main

import (
	"context"
	"log"

	anilistpkg "github.com/bigspawn/anilist-mal-sync/internal/anilist"
	"golang.org/x/oauth2"
)

// Re-export Anilist client type and constructor
type AnilistClient = anilistpkg.AnilistClient

func NewAnilistClient(ctx context.Context, oauth *OAuth, username string) *AnilistClient {
	return anilistpkg.NewAnilistClient(ctx, oauth, username)
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

	if oauthAnilist.NeedInit() {
		getToken(ctx, oauthAnilist, config.OAuth.Port)
	} else {
		log.Println("Token already set, no need to start server")
	}

	return oauthAnilist, nil
}
