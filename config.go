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

type Config struct {
	OAuth           OAuthConfig           `yaml:"oauth"`
	Anilist         SiteConfig            `yaml:"anilist"`
	MyAnimeList     SiteConfig            `yaml:"myanimelist"`
	TokenFilePath   string                `yaml:"token_file_path"`
	Watch           WatchConfig           `yaml:"watch"`
	HTTPTimeout     string                `yaml:"http_timeout"`
	OfflineDatabase OfflineDatabaseConfig `yaml:"offline_database"`
	ARMAPI          ARMAPIConfig          `yaml:"arm_api"`
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
}

func overrideOAuthFromEnv(oauth *OAuthConfig) {
	if port := os.Getenv("OAUTH_PORT"); port != "" {
		oauth.Port = port
	} else if port := os.Getenv("PORT"); port != "" {
		oauth.Port = port
	}

	if redirectURI := os.Getenv("OAUTH_REDIRECT_URI"); redirectURI != "" {
		oauth.RedirectURI = redirectURI
	}
}

func overrideAnilistFromEnv(anilist *SiteConfig) {
	if clientID := os.Getenv("ANILIST_CLIENT_ID"); clientID != "" {
		anilist.ClientID = clientID
	}

	// Support both new and old env var names for backward compatibility
	if clientSecret := os.Getenv("ANILIST_CLIENT_SECRET"); clientSecret != "" {
		anilist.ClientSecret = clientSecret
	} else if clientSecret := os.Getenv("CLIENT_SECRET_ANILIST"); clientSecret != "" {
		anilist.ClientSecret = clientSecret
	}

	if username := os.Getenv("ANILIST_USERNAME"); username != "" {
		anilist.Username = username
	}
}

func overrideMyAnimeListFromEnv(mal *SiteConfig) {
	if clientID := os.Getenv("MAL_CLIENT_ID"); clientID != "" {
		mal.ClientID = clientID
	}

	// Support both new and old env var names for backward compatibility
	if clientSecret := os.Getenv("MAL_CLIENT_SECRET"); clientSecret != "" {
		mal.ClientSecret = clientSecret
	} else if clientSecret := os.Getenv("CLIENT_SECRET_MYANIMELIST"); clientSecret != "" {
		mal.ClientSecret = clientSecret
	}

	if username := os.Getenv("MAL_USERNAME"); username != "" {
		mal.Username = username
	}
}

func overrideWatchFromEnv(watch *WatchConfig) {
	if interval := os.Getenv("WATCH_INTERVAL"); interval != "" {
		watch.Interval = interval
	}
}

func overrideTokenPathFromEnv(cfg *Config) {
	if tokenFilePath := os.Getenv("TOKEN_FILE_PATH"); tokenFilePath != "" {
		cfg.TokenFilePath = tokenFilePath
	}

	if cfg.TokenFilePath == "" {
		cfg.TokenFilePath = getDefaultTokenPathOrEmpty()
	}
}

func overrideHTTPTimeoutFromEnv(cfg *Config) {
	if timeout := os.Getenv("HTTP_TIMEOUT"); timeout != "" {
		cfg.HTTPTimeout = timeout
	}
}

func overrideOfflineDatabaseFromEnv(odc *OfflineDatabaseConfig) {
	if v := os.Getenv("OFFLINE_DATABASE_ENABLED"); v != "" {
		odc.Enabled = parseBoolString(v)
	}
	if v := os.Getenv("OFFLINE_DATABASE_CACHE_DIR"); v != "" {
		odc.CacheDir = v
	}
	if v := os.Getenv("OFFLINE_DATABASE_AUTO_UPDATE"); v != "" {
		odc.AutoUpdate = parseBoolString(v)
	}
}

func overrideARMAPIFromEnv(ac *ARMAPIConfig) {
	if v := os.Getenv("ARM_API_ENABLED"); v != "" {
		ac.Enabled = parseBoolString(v)
	}
	if v := os.Getenv("ARM_API_URL"); v != "" {
		ac.BaseURL = v
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
