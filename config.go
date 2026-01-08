package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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

type WatchConfig struct {
	Interval string `yaml:"interval"`
}

// GetInterval parses the interval string into a duration.
// Returns 0 if interval is empty (not specified).
func (w *WatchConfig) GetInterval() (time.Duration, error) {
	if w.Interval == "" {
		return 0, nil
	}
	return time.ParseDuration(w.Interval)
}

type Config struct {
	OAuth         OAuthConfig `yaml:"oauth"`
	Anilist       SiteConfig  `yaml:"anilist"`
	MyAnimeList   SiteConfig  `yaml:"myanimelist"`
	TokenFilePath string      `yaml:"token_file_path"`
	Watch         WatchConfig `yaml:"watch"`
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() Config {
	cfg := Config{
		OAuth: OAuthConfig{
			Port:        getEnvOrDefault("OAUTH_PORT", getEnvOrDefault("PORT", "18080")),
			RedirectURI: getEnvOrDefault("OAUTH_REDIRECT_URI", "http://localhost:18080/callback"),
		},
		Anilist: SiteConfig{
			ClientID:     os.Getenv("ANILIST_CLIENT_ID"),
			ClientSecret: os.Getenv("ANILIST_CLIENT_SECRET"),
			Username:     os.Getenv("ANILIST_USERNAME"),
		},
		MyAnimeList: SiteConfig{
			ClientID:     os.Getenv("MAL_CLIENT_ID"),
			ClientSecret: os.Getenv("MAL_CLIENT_SECRET"),
			Username:     os.Getenv("MAL_USERNAME"),
		},
		TokenFilePath: getEnvOrDefault("TOKEN_FILE_PATH", os.ExpandEnv("$HOME/.config/anilist-mal-sync/token.json")),
		Watch: WatchConfig{
			Interval: os.Getenv("WATCH_INTERVAL"),
		},
	}
	return cfg
}

// getEnvOrDefault returns environment variable value or default if empty
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// overrideConfigFromEnv applies environment variable overrides to a config
func overrideConfigFromEnv(cfg *Config) {
	if port := os.Getenv("OAUTH_PORT"); port != "" {
		cfg.OAuth.Port = port
	} else if port := os.Getenv("PORT"); port != "" {
		cfg.OAuth.Port = port
	}

	if redirectURI := os.Getenv("OAUTH_REDIRECT_URI"); redirectURI != "" {
		cfg.OAuth.RedirectURI = redirectURI
	}

	if clientID := os.Getenv("ANILIST_CLIENT_ID"); clientID != "" {
		cfg.Anilist.ClientID = clientID
	}

	// Support both new and old env var names for backward compatibility
	if clientSecret := os.Getenv("ANILIST_CLIENT_SECRET"); clientSecret != "" {
		cfg.Anilist.ClientSecret = clientSecret
	} else if clientSecret := os.Getenv("CLIENT_SECRET_ANILIST"); clientSecret != "" {
		cfg.Anilist.ClientSecret = clientSecret
	}

	if username := os.Getenv("ANILIST_USERNAME"); username != "" {
		cfg.Anilist.Username = username
	}

	if clientID := os.Getenv("MAL_CLIENT_ID"); clientID != "" {
		cfg.MyAnimeList.ClientID = clientID
	}

	if username := os.Getenv("MAL_USERNAME"); username != "" {
		cfg.MyAnimeList.Username = username
	}

	// Support both new and old env var names for backward compatibility
	if clientSecret := os.Getenv("MAL_CLIENT_SECRET"); clientSecret != "" {
		cfg.MyAnimeList.ClientSecret = clientSecret
	} else if clientSecret := os.Getenv("CLIENT_SECRET_MYANIMELIST"); clientSecret != "" {
		cfg.MyAnimeList.ClientSecret = clientSecret
	}

	if interval := os.Getenv("WATCH_INTERVAL"); interval != "" {
		cfg.Watch.Interval = interval
	}

	if tokenFilePath := os.Getenv("TOKEN_FILE_PATH"); tokenFilePath != "" {
		cfg.TokenFilePath = tokenFilePath
	}

	if cfg.TokenFilePath == "" {
		cfg.TokenFilePath = os.ExpandEnv("$HOME/.config/anilist-mal-sync/token.json")
	}
}

func validateConfig(cfg Config) error {
	if cfg.Anilist.ClientID == "" || cfg.Anilist.Username == "" ||
		cfg.MyAnimeList.ClientID == "" || cfg.MyAnimeList.Username == "" {
		return fmt.Errorf("required fields not set")
	}
	return nil
}

func loadConfigFromFile(filename string) (Config, error) {
	// If no config file specified, load from environment variables only
	if filename == "" {
		cfg := loadConfigFromEnv()
		if err := validateConfig(cfg); err != nil {
			return Config{}, fmt.Errorf("required environment variables not set (ANILIST_CLIENT_ID, ANILIST_USERNAME, MAL_CLIENT_ID, MAL_USERNAME)")
		}
		return cfg, nil
	}

	// Try to load from file
	// #nosec G304 - Config file path is provided by user via command line flag
	data, err := os.ReadFile(filename)
	if err != nil {
		// If file not found, try loading from env vars
		if os.IsNotExist(err) {
			cfg := loadConfigFromEnv()
			// Validate that required fields are set
			if err := validateConfig(cfg); err != nil {
				// Print help message to stderr
				fmt.Fprintln(os.Stderr, getConfigHelp(filename))
				return Config{}, fmt.Errorf("config file not found and required environment variables not set: %w", err)
			}
			return cfg, nil
		}
		// Other read errors
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("config file not found: %w", err)
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Environment variables override file values
	overrideConfigFromEnv(&cfg)

	// Validate required fields
	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("required fields not set (anilist.client_id, anilist.username, myanimelist.client_id, myanimelist.username)")
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
