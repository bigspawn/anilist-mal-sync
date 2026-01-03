package main

import (
	cfg "github.com/bigspawn/anilist-mal-sync/internal/config"
)

// Re-export config types from internal/config so existing callers in package main
// can continue to use the same type names.
type OAuthConfig = cfg.OAuthConfig
type SiteConfig = cfg.SiteConfig
type Config = cfg.Config

func loadConfigFromFile(filename string) (Config, error) {
	return cfg.Load(filename)
}
