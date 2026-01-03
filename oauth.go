package main

import (
	"context"

	oauthpkg "github.com/bigspawn/anilist-mal-sync/internal/oauth"
	"golang.org/x/oauth2"
)

// Re-export OAuth type and constructors from internal/oauth to keep existing
// callers unchanged.
type OAuth = oauthpkg.OAuth

func NewOAuth(
	config SiteConfig,
	redirectURI string,
	siteName string,
	authCodeOptions []oauth2.AuthCodeOption,
	tokenFilePath string,
) (*OAuth, error) {
	return oauthpkg.NewOAuth(config, redirectURI, siteName, authCodeOptions, tokenFilePath)
}

func getToken(ctx context.Context, oauth *OAuth, port string) {
	oauthpkg.GetToken(ctx, oauth, port)
}
