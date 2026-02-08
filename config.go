package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// GetHTTPTimeout parses the http_timeout string into a duration.
// Returns 30s as default if not specified or invalid.
func (c *Config) GetHTTPTimeout() time.Duration {
	if c.HTTPTimeout == "" {
		return 30 * time.Second // default
	}
	dur, err := time.ParseDuration(c.HTTPTimeout)
	if err != nil {
		log.Printf("Invalid http_timeout format '%s', using default 30s: %v", c.HTTPTimeout, err)
		return 30 * time.Second
	}
	return dur
}

type OfflineDatabaseConfig struct {
	Enabled      bool   `yaml:"enabled"`
	CacheDir     string `yaml:"cache_dir"`
	AutoUpdate   bool   `yaml:"auto_update"`
	ForceRefresh bool   `yaml:"-"` // CLI flag only
}

type ARMAPIConfig struct {
	Enabled bool   `yaml:"enabled"`
	BaseURL string `yaml:"base_url"`
}

type HatoAPIConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BaseURL     string `yaml:"base_url"`
	CacheDir    string `yaml:"cache_dir"`
	CacheMaxAge string `yaml:"cache_max_age"`
}

type Config struct {
	OAuth           OAuthConfig           `yaml:"oauth"`
	Anilist         SiteConfig            `yaml:"anilist"`
	MyAnimeList     SiteConfig            `yaml:"myanimelist"`
	TokenFilePath   string                `yaml:"token_file_path"`
	Watch           WatchConfig           `yaml:"watch"`
	HTTPTimeout     string                `yaml:"http_timeout"`
	OfflineDatabase OfflineDatabaseConfig `yaml:"offline_database"`
	ARMAPI          ARMAPIConfig          `yaml:"arm_api"`
	HatoAPI         HatoAPIConfig         `yaml:"hato_api"`
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() (Config, error) {
	tokenPath, err := getDefaultTokenPath()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		OAuth: OAuthConfig{
			Port:        getEnvOrDefault("OAUTH_PORT", getEnvOrDefault("PORT", "18080")),
			RedirectURI: getEnvOrDefault("OAUTH_REDIRECT_URI", "http://localhost:18080/callback"),
		},
		Anilist: SiteConfig{
			ClientID:     os.Getenv("ANILIST_CLIENT_ID"),
			ClientSecret: os.Getenv("ANILIST_CLIENT_SECRET"),
			Username:     os.Getenv("ANILIST_USERNAME"),
			AuthURL:      "https://anilist.co/api/v2/oauth/authorize",
			TokenURL:     "https://anilist.co/api/v2/oauth/token",
		},
		MyAnimeList: SiteConfig{
			ClientID:     os.Getenv("MAL_CLIENT_ID"),
			ClientSecret: os.Getenv("MAL_CLIENT_SECRET"),
			Username:     os.Getenv("MAL_USERNAME"),
			AuthURL:      "https://myanimelist.net/v1/oauth2/authorize",
			TokenURL:     "https://myanimelist.net/v1/oauth2/token",
		},
		TokenFilePath: getEnvOrDefault("TOKEN_FILE_PATH", tokenPath),
		Watch: WatchConfig{
			Interval: os.Getenv("WATCH_INTERVAL"),
		},
		HTTPTimeout: getEnvOrDefault("HTTP_TIMEOUT", "30s"),
		OfflineDatabase: OfflineDatabaseConfig{
			Enabled:    getEnvBoolOrDefault("OFFLINE_DATABASE_ENABLED", true),
			CacheDir:   getEnvOrDefault("OFFLINE_DATABASE_CACHE_DIR", getDefaultCacheDir()),
			AutoUpdate: getEnvBoolOrDefault("OFFLINE_DATABASE_AUTO_UPDATE", true),
		},
		ARMAPI: ARMAPIConfig{
			Enabled: getEnvBoolOrDefault("ARM_API_ENABLED", false),
			BaseURL: getEnvOrDefault("ARM_API_URL", defaultARMBaseURL),
		},
		HatoAPI: HatoAPIConfig{
			Enabled:     getEnvBoolOrDefault("HATO_API_ENABLED", true),
			BaseURL:     getEnvOrDefault("HATO_API_URL", defaultHatoBaseURL),
			CacheDir:    getEnvOrDefault("HATO_API_CACHE_DIR", getDefaultHatoCacheDir()),
			CacheMaxAge: getEnvOrDefault("HATO_API_CACHE_MAX_AGE", "720h"),
		},
	}
	return cfg, nil
}

// parseBoolString parses a string as a boolean value.
func parseBoolString(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}

// getEnvBoolOrDefault returns environment variable as bool or default if empty.
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return parseBoolString(value)
}

// getEnvOrDefault returns environment variable value or default if empty
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultTokenPath returns the default token file path for the current platform
func getDefaultTokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "anilist-mal-sync", "token.json"), nil
}

// overrideConfigFromEnv applies environment variable overrides to a config
func overrideConfigFromEnv(cfg *Config) {
	overrideOAuthFromEnv(&cfg.OAuth)
	overrideAnilistFromEnv(&cfg.Anilist)
	overrideMyAnimeListFromEnv(&cfg.MyAnimeList)
	overrideWatchFromEnv(&cfg.Watch)
	overrideHTTPTimeoutFromEnv(cfg)
	overrideTokenPathFromEnv(cfg)
	overrideOfflineDatabaseFromEnv(&cfg.OfflineDatabase)
	overrideARMAPIFromEnv(&cfg.ARMAPI)
	overrideHatoAPIFromEnv(&cfg.HatoAPI)
}

func overrideOAuthFromEnv(oauth *OAuthConfig) {
	if port := os.Getenv("OAUTH_PORT"); port != "" {
		oauth.Port = port
	} else if port := os.Getenv("PORT"); port != "" {
		oauth.Port = port
	}

	overrideStringFromEnv(&oauth.RedirectURI, "OAUTH_REDIRECT_URI")
}

func overrideAnilistFromEnv(anilist *SiteConfig) {
	overrideStringFromEnv(&anilist.ClientID, "ANILIST_CLIENT_ID")
	overrideStringFromEnv(&anilist.ClientSecret, "ANILIST_CLIENT_SECRET", "CLIENT_SECRET_ANILIST")
	overrideStringFromEnv(&anilist.Username, "ANILIST_USERNAME")
}

func overrideMyAnimeListFromEnv(mal *SiteConfig) {
	overrideStringFromEnv(&mal.ClientID, "MAL_CLIENT_ID")
	overrideStringFromEnv(&mal.ClientSecret, "MAL_CLIENT_SECRET", "CLIENT_SECRET_MYANIMELIST")
	overrideStringFromEnv(&mal.Username, "MAL_USERNAME")
}

func overrideWatchFromEnv(watch *WatchConfig) {
	overrideStringFromEnv(&watch.Interval, "WATCH_INTERVAL")
}

func overrideTokenPathFromEnv(cfg *Config) {
	overrideStringFromEnv(&cfg.TokenFilePath, "TOKEN_FILE_PATH")

	if cfg.TokenFilePath == "" {
		cfg.TokenFilePath = getDefaultTokenPathOrEmpty()
	}
}

func overrideHTTPTimeoutFromEnv(cfg *Config) {
	overrideStringFromEnv(&cfg.HTTPTimeout, "HTTP_TIMEOUT")
}

func overrideOfflineDatabaseFromEnv(odc *OfflineDatabaseConfig) {
	overrideBoolFromEnv(&odc.Enabled, "OFFLINE_DATABASE_ENABLED")
	overrideStringFromEnv(&odc.CacheDir, "OFFLINE_DATABASE_CACHE_DIR")
	overrideBoolFromEnv(&odc.AutoUpdate, "OFFLINE_DATABASE_AUTO_UPDATE")
}

func overrideARMAPIFromEnv(ac *ARMAPIConfig) {
	overrideBoolFromEnv(&ac.Enabled, "ARM_API_ENABLED")
	overrideStringFromEnv(&ac.BaseURL, "ARM_API_URL")
}

func overrideHatoAPIFromEnv(hc *HatoAPIConfig) {
	overrideBoolFromEnv(&hc.Enabled, "HATO_API_ENABLED")
	overrideStringFromEnv(&hc.BaseURL, "HATO_API_URL")
	overrideStringFromEnv(&hc.CacheDir, "HATO_API_CACHE_DIR")
	overrideStringFromEnv(&hc.CacheMaxAge, "HATO_API_CACHE_MAX_AGE")
}

// overrideStringFromEnv overrides a string field from environment variables.
// Tries each key in order until a non-empty value is found.
func overrideStringFromEnv(field *string, keys ...string) {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			*field = value
			return
		}
	}
}

// overrideBoolFromEnv overrides a boolean field from an environment variable.
func overrideBoolFromEnv(field *bool, key string) {
	if value := os.Getenv(key); value != "" {
		*field = parseBoolString(value)
	}
}

// getDefaultTokenPathOrEmpty returns the default token path or empty string on error
func getDefaultTokenPathOrEmpty() string {
	path, err := getDefaultTokenPath()
	if err != nil {
		return ""
	}
	return path
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
		return loadConfigFromEnvWithValidation()
	}

	// Try to load from file
	// #nosec G304 - Config file path is provided by user via command line flag
	data, err := os.ReadFile(filename)
	if err != nil {
		return handleConfigFileReadError(err, filename)
	}

	cfg, err := parseConfigFile(data, filename)
	if err != nil {
		return Config{}, err
	}

	// Environment variables override file values
	overrideConfigFromEnv(&cfg)

	// Validate required fields
	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("required fields not set (anilist.client_id, anilist.username, myanimelist.client_id, myanimelist.username)")
	}

	return cfg, nil
}

func loadConfigFromEnvWithValidation() (Config, error) {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return Config{}, err
	}
	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("required environment variables not set (ANILIST_CLIENT_ID, ANILIST_USERNAME, MAL_CLIENT_ID, MAL_USERNAME)")
	}
	return cfg, nil
}

func handleConfigFileReadError(readErr error, filename string) (Config, error) {
	// If file not found, try loading from env vars
	if os.IsNotExist(readErr) {
		return tryLoadFromEnvWithHelp(filename)
	}
	// Other read errors
	fmt.Fprintln(os.Stderr, getConfigHelp(filename))
	return Config{}, fmt.Errorf("config file not found: %w", readErr)
}

func tryLoadFromEnvWithHelp(filename string) (Config, error) {
	cfg, envErr := loadConfigFromEnv()
	if envErr != nil {
		return Config{}, envErr
	}
	// Validate that required fields are set
	if err := validateConfig(cfg); err != nil {
		// Print help message to stderr
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("config file not found and required environment variables not set: %w", err)
	}
	return cfg, nil
}

func parseConfigFile(data []byte, filename string) (Config, error) {
	cfg := configWithDefaults()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

// configWithDefaults returns a Config with default values pre-filled
// so that YAML fields not specified keep their defaults.
func configWithDefaults() Config {
	return Config{
		OfflineDatabase: OfflineDatabaseConfig{
			Enabled:    true,
			CacheDir:   getDefaultCacheDir(),
			AutoUpdate: true,
		},
		ARMAPI: ARMAPIConfig{
			Enabled: false,
			BaseURL: defaultARMBaseURL,
		},
		HatoAPI: HatoAPIConfig{
			Enabled:     true,
			BaseURL:     defaultHatoBaseURL,
			CacheDir:    getDefaultHatoCacheDir(),
			CacheMaxAge: "720h",
		},
	}
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
