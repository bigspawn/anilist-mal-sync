package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type OAuthConfig struct {
	Port        string `yaml:"port"`
	RedirectURI string `yaml:"redirect_uri"`
}

type SiteConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	AuthURL      string `yaml:"auth_url"`
	TokenURL     string `yaml:"token_url"`
	Username     string `yaml:"username"`
}

type Config struct {
	OAuth         OAuthConfig `yaml:"oauth"`
	Anilist       SiteConfig  `yaml:"anilist"`
	MyAnimeList   SiteConfig  `yaml:"myanimelist"`
	TokenFilePath string      `yaml:"token_file_path"`
}

func loadConfigFromFile(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.OAuth.Port = port
	}

	if clientSecret := os.Getenv("CLIENT_SECRET_ANILIST"); clientSecret != "" {
		cfg.Anilist.ClientSecret = clientSecret
	}

	if clientSecret := os.Getenv("CLIENT_SECRET_MYANIMELIST"); clientSecret != "" {
		cfg.MyAnimeList.ClientSecret = clientSecret
	}

	if cfg.TokenFilePath == "" {
		cfg.TokenFilePath = os.ExpandEnv("$HOME/.config/anilist-mal-sync/token.json")
	}

	return cfg, nil
}
