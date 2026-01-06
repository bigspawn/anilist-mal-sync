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
	OAuth         OAuthConfig `yaml:"oauth"`
	Anilist       SiteConfig  `yaml:"anilist"`
	MyAnimeList   SiteConfig  `yaml:"myanimelist"`
	TokenFilePath string      `yaml:"token_file_path"`
}

func loadConfigFromFile(filename string) (Config, error) {
	// #nosec G304 - Config file path is provided by user via command line flag
	data, err := os.ReadFile(filename)
	if err != nil {
		// Print help message to stderr
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("config file not found: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		// Print help message to stderr
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
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

// getConfigHelp returns a helpful message for creating config file
func getConfigHelp(configPath string) string {
	const (
		colorReset  = "\033[0m"
		colorBold   = "\033[1m"
		colorRed    = "\033[31m"
		colorYellow = "\033[33m"
		colorCyan   = "\033[36m"
	)

	examplePath := "config.example.yaml"

	return fmt.Sprintf(`
%sConfiguration file not found or invalid!%s

%sTo fix this:%s

1. Copy the example config:
   %scp %s %s%s

2. Edit the config file with your credentials:
   %snano config.yaml%s

3. Then run the command again.

%sAlternatively, you can specify a custom config path:%s
   %sanilist-mal-sync -c /path/to/config.yaml%s

`, colorBold+colorRed, colorReset,
		colorBold+colorYellow, colorReset,
		colorCyan, examplePath, configPath, colorReset,
		colorCyan, colorReset,
		colorBold+colorYellow, colorReset,
		colorCyan, colorReset)
}

// IsConfigNotFoundError checks if error is related to config file
func IsConfigNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "config file not found") ||
		strings.Contains(errMsg, "no such file or directory") ||
		strings.Contains(errMsg, "failed to parse config")
}
