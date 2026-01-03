package main

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
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
	OAuth              OAuthConfig `yaml:"oauth"`
	Anilist            SiteConfig  `yaml:"anilist"`
	MyAnimeList        SiteConfig  `yaml:"myanimelist"`
	TokenFilePath      string      `yaml:"token_file_path"`
	ScoreNormalization *bool       `yaml:"score_normalization"`
	IgnoreAnimeTitles  []string    `yaml:"ignore_anime_titles"` // Anime titles to skip (case-insensitive)
	IgnoreMangaTitles  []string    `yaml:"ignore_manga_titles"` // Manga titles to skip (case-insensitive)
}

// Load reads configuration from a YAML file with environment variable overrides
func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %q: %w", filename, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file %q: %w", filename, err)
	}

	if port := os.Getenv(EnvVarPort); port != "" {
		cfg.OAuth.Port = port
	}

	if clientSecret := os.Getenv(EnvVarClientSecretAnilist); clientSecret != "" {
		cfg.Anilist.ClientSecret = clientSecret
	}

	if clientSecret := os.Getenv(EnvVarClientSecretMyAnimeList); clientSecret != "" {
		cfg.MyAnimeList.ClientSecret = clientSecret
	}

	if cfg.TokenFilePath == "" {
		cfg.TokenFilePath = os.ExpandEnv(DefaultTokenFilePath)
	}

	if cfg.ScoreNormalization == nil {
		v := DefaultScoreNormalization
		cfg.ScoreNormalization = &v
	}

	return cfg, nil
}

// GetIgnoreAnimeTitlesMap converts anime ignore titles to a map for O(1) lookup
func (c Config) GetIgnoreAnimeTitlesMap() map[string]struct{} {
	ignoreMap := make(map[string]struct{}, len(c.IgnoreAnimeTitles))
	for _, title := range c.IgnoreAnimeTitles {
		ignoreMap[strings.ToLower(title)] = struct{}{}
	}
	return ignoreMap
}

// GetIgnoreMangaTitlesMap converts manga ignore titles to a map for O(1) lookup
func (c Config) GetIgnoreMangaTitlesMap() map[string]struct{} {
	ignoreMap := make(map[string]struct{}, len(c.IgnoreMangaTitles))
	for _, title := range c.IgnoreMangaTitles {
		ignoreMap[strings.ToLower(title)] = struct{}{}
	}
	return ignoreMap
}
