package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	yaml "gopkg.in/yaml.v2"
)

const (
	minInterval = 1 * time.Hour
	maxInterval = 168 * time.Hour // 7 days
)

var ErrWatchConfigMissing = errors.New("watch mode requires either interval or schedule")

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
	Schedule string `yaml:"schedule"`
}

// GetInterval parses the interval string into a duration.
// Returns 0 if interval is empty (not specified).
func (w *WatchConfig) GetInterval() (time.Duration, error) {
	if w.Interval == "" {
		return 0, nil
	}
	return time.ParseDuration(w.Interval)
}

// Validate checks that exactly one of Interval or Schedule is set and valid.
func (w *WatchConfig) Validate() error {
	if w.Interval != "" && w.Schedule != "" {
		return fmt.Errorf(
			"watch mode accepts either interval or schedule, not both (interval=%s, schedule=%q)",
			w.Interval, w.Schedule,
		)
	}
	if w.Interval != "" {
		d, err := time.ParseDuration(w.Interval)
		if err != nil {
			return fmt.Errorf("invalid watch interval %q: %w", w.Interval, err)
		}
		if d < minInterval {
			return fmt.Errorf("interval must be at least 1h (got %v)", d)
		}
		if d > maxInterval {
			return fmt.Errorf("interval must be at most 168h/7days (got %v)", d)
		}
		return nil
	}
	if w.Schedule != "" {
		_, err := cron.ParseStandard(w.Schedule)
		if err != nil {
			return fmt.Errorf("invalid watch schedule %q: %w", w.Schedule, err)
		}
		return nil
	}
	return ErrWatchConfigMissing
}

// ParseSchedule parses the schedule string and returns a cron.Schedule.
func (w *WatchConfig) ParseSchedule() (cron.Schedule, error) {
	if w.Schedule == "" {
		return nil, errors.New("schedule is empty")
	}
	return cron.ParseStandard(w.Schedule)
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

// defaultMangaBakaBaseURL is the MangaBaka API base URL, verified 2026-09-12 (see plan docs).
const defaultMangaBakaBaseURL = "https://api.mangabaka.org/v1"

// defaultLongCacheMaxAge is the shared default for ID-mapping caches whose
// upstream data rarely changes (Hato, MangaBaka).
const defaultLongCacheMaxAge = "720h"

// MappingSourceConfig is the shared config shape for ID-mapping sources
// (ARM, Hato, Jikan, MangaBaka). ARM and Jikan simply leave the fields
// they never use unset.
type MappingSourceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BaseURL     string `yaml:"base_url"`
	CacheDir    string `yaml:"cache_dir"`
	CacheMaxAge string `yaml:"cache_max_age"`
}

type FavoritesConfig struct {
	Enabled bool `yaml:"enabled"`
}

type Config struct {
	OAuth            OAuthConfig           `yaml:"oauth"`
	Anilist          SiteConfig            `yaml:"anilist"`
	MyAnimeList      SiteConfig            `yaml:"myanimelist"`
	TokenFilePath    string                `yaml:"token_file_path"`
	Watch            WatchConfig           `yaml:"watch"`
	HTTPTimeout      string                `yaml:"http_timeout"`
	OfflineDatabase  OfflineDatabaseConfig `yaml:"offline_database"`
	ARMAPI           MappingSourceConfig   `yaml:"arm_api"`
	HatoAPI          MappingSourceConfig   `yaml:"hato_api"`
	JikanAPI         MappingSourceConfig   `yaml:"jikan_api"`
	MangaBakaAPI     MappingSourceConfig   `yaml:"mangabaka_api"`
	Favorites        FavoritesConfig       `yaml:"favorites"`
	MappingsFilePath string                `yaml:"mappings_file_path"`
}

// loadConfigFromEnv builds a Config by layering environment variables on
// top of defaultConfig() — the entry point used when no config file is
// given. TokenFilePath is resolved here rather than in defaultConfig()
// because getDefaultTokenPath can fail (a fallible OS lookup), and this is
// the one entry point that has always surfaced that failure as an error.
func loadConfigFromEnv() (Config, error) {
	tokenPath, err := getDefaultTokenPath()
	if err != nil {
		return Config{}, err
	}

	cfg := defaultConfig()
	cfg.TokenFilePath = getEnvOrDefault("TOKEN_FILE_PATH", tokenPath)
	overrideConfigFromEnv(&cfg)
	return cfg, nil
}

// parseBoolString parses a string as a boolean value.
func parseBoolString(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}

// getEnvOrDefault returns environment variable value or default if empty.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultTokenPath returns the default token file path for the current platform.
func getDefaultTokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "anilist-mal-sync", "token.json"), nil
}

// overrideConfigFromEnv applies environment variable overrides to a config.
func overrideConfigFromEnv(cfg *Config) {
	overrideOAuthFromEnv(&cfg.OAuth)
	overrideAnilistFromEnv(&cfg.Anilist)
	overrideMyAnimeListFromEnv(&cfg.MyAnimeList)
	overrideWatchFromEnv(&cfg.Watch)
	overrideHTTPTimeoutFromEnv(cfg)
	overrideTokenPathFromEnv(cfg)
	overrideOfflineDatabaseFromEnv(&cfg.OfflineDatabase)
	for _, src := range mappingSources() {
		overrideMappingSourceFromEnv(src.configField(cfg), src)
	}
	overrideFavoritesFromEnv(&cfg.Favorites)
	overrideStringFromEnv(&cfg.MappingsFilePath, "MAPPINGS_FILE_PATH")
}

func overrideOAuthFromEnv(oauth *OAuthConfig) {
	if port := os.Getenv("OAUTH_PORT"); port != "" {
		oauth.Port = port
	} else if port := os.Getenv("PORT"); port != "" {
		oauth.Port = port
	}

	overrideStringFromEnv(&oauth.RedirectURI, "OAUTH_REDIRECT_URI")
}

// overrideAnilistFromEnv applies the AniList env vars shared by both entry
// points. The legacy CLIENT_SECRET_ANILIST fallback is not here — see
// overrideLegacySecretsFromEnv — because loadConfigFromEnv (no config file)
// has never supported it, only the YAML config path has.
func overrideAnilistFromEnv(anilist *SiteConfig) {
	overrideStringFromEnv(&anilist.ClientID, "ANILIST_CLIENT_ID")
	overrideStringFromEnv(&anilist.ClientSecret, "ANILIST_CLIENT_SECRET")
	overrideStringFromEnv(&anilist.Username, "ANILIST_USERNAME")
}

// overrideMyAnimeListFromEnv applies the MAL env vars shared by both entry
// points; see overrideAnilistFromEnv's comment on the legacy secret.
func overrideMyAnimeListFromEnv(mal *SiteConfig) {
	overrideStringFromEnv(&mal.ClientID, "MAL_CLIENT_ID")
	overrideStringFromEnv(&mal.ClientSecret, "MAL_CLIENT_SECRET")
	overrideStringFromEnv(&mal.Username, "MAL_USERNAME")
}

// overrideLegacySecretsFromEnv falls back to the deprecated
// CLIENT_SECRET_ANILIST / CLIENT_SECRET_MYANIMELIST env vars. Only the YAML
// config path (loadConfigFromFile) calls this — loadConfigFromEnv
// intentionally does not, and TestLoadConfigFromEnv_Legacy*SecretFallback
// pin that.
func overrideLegacySecretsFromEnv(cfg *Config) {
	// The legacy key only applies when the new-style key wasn't set — same
	// first-non-empty-wins precedence overrideStringFromEnv gives a single
	// field with multiple keys, just split across two entry-point-scoped calls.
	if os.Getenv("ANILIST_CLIENT_SECRET") == "" {
		overrideStringFromEnv(&cfg.Anilist.ClientSecret, "CLIENT_SECRET_ANILIST")
	}
	if os.Getenv("MAL_CLIENT_SECRET") == "" {
		overrideStringFromEnv(&cfg.MyAnimeList.ClientSecret, "CLIENT_SECRET_MYANIMELIST")
	}
}

func overrideWatchFromEnv(watch *WatchConfig) {
	overrideStringFromEnv(&watch.Interval, "WATCH_INTERVAL")
	overrideStringFromEnv(&watch.Schedule, "WATCH_SCHEDULE")
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

// overrideMappingSourceFromEnv applies env-var overrides for one ID-mapping
// source, described by src, so all four sources share one override path.
func overrideMappingSourceFromEnv(sc *MappingSourceConfig, src mappingSource) {
	overrideBoolFromEnv(&sc.Enabled, src.envPrefix+"ENABLED")
	if src.urlFlag != "" {
		overrideStringFromEnv(&sc.BaseURL, src.envPrefix+"URL")
	}
	if src.defaultCacheDir != nil {
		overrideStringFromEnv(&sc.CacheDir, src.envPrefix+"CACHE_DIR")
		overrideStringFromEnv(&sc.CacheMaxAge, src.envPrefix+"CACHE_MAX_AGE")
	}
}

func overrideFavoritesFromEnv(fc *FavoritesConfig) {
	overrideBoolFromEnv(&fc.Enabled, "FAVORITES_SYNC_ENABLED")
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

// getDefaultTokenPathOrEmpty returns the default token path or empty string on error.
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
		return errors.New("required fields not set")
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
	overrideLegacySecretsFromEnv(&cfg)

	// Validate required fields
	err = validateConfig(cfg)
	if err != nil {
		return Config{}, errors.New("required fields not set (anilist.client_id, anilist.username, myanimelist.client_id, myanimelist.username)")
	}

	return cfg, nil
}

func loadConfigFromEnvWithValidation() (Config, error) {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return Config{}, err
	}
	err = validateConfig(cfg)
	if err != nil {
		return Config{}, errors.New("required environment variables not set (ANILIST_CLIENT_ID, ANILIST_USERNAME, MAL_CLIENT_ID, MAL_USERNAME)")
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
	err := validateConfig(cfg)
	if err != nil {
		// Print help message to stderr
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("config file not found and required environment variables not set: %w", err)
	}
	return cfg, nil
}

func parseConfigFile(data []byte, filename string) (Config, error) {
	cfg := defaultConfig()
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, getConfigHelp(filename))
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

// AniList and MyAnimeList OAuth2 endpoints — public, well-known URLs, not
// secrets — used as the SiteConfig defaults below.
const (
	defaultAnilistAuthURL  = "https://anilist.co/api/v2/oauth/authorize"
	defaultAnilistTokenURL = "https://anilist.co/api/v2/oauth/token" // #nosec G101 -- public endpoint, not a secret
	defaultMALAuthURL      = "https://myanimelist.net/v1/oauth2/authorize"
	defaultMALTokenURL     = "https://myanimelist.net/v1/oauth2/token" // #nosec G101 -- public endpoint, not a secret
)

// defaultConfig is the one place every built-in default lives. Both
// loadConfigFromEnv (no config file) and parseConfigFile (a config.yaml,
// pre-filled here before unmarshalling) build their Config by layering on
// top of this, so a default is never declared twice — the way MangaBaka's
// Enabled default once was (hardcoded true here, false everywhere else).
// TokenFilePath is not set here: see loadConfigFromEnv's doc comment.
func defaultConfig() Config {
	cfg := Config{
		OAuth: OAuthConfig{
			Port:        "18080",
			RedirectURI: "http://localhost:18080/callback",
		},
		Anilist: SiteConfig{
			AuthURL:  defaultAnilistAuthURL,
			TokenURL: defaultAnilistTokenURL,
		},
		MyAnimeList: SiteConfig{
			AuthURL:  defaultMALAuthURL,
			TokenURL: defaultMALTokenURL,
		},
		HTTPTimeout: "30s",
		OfflineDatabase: OfflineDatabaseConfig{
			Enabled:    true,
			CacheDir:   getDefaultCacheDir(),
			AutoUpdate: true,
		},
		MappingsFilePath: getDefaultMappingsPath(),
	}
	for _, src := range mappingSources() {
		src.setConfig(&cfg, mappingSourceDefaultConfig(src))
	}
	return cfg
}

// mappingSourceDefaultConfig builds one ID-mapping source's built-in-default
// config, with no env vars involved.
func mappingSourceDefaultConfig(src mappingSource) MappingSourceConfig {
	sc := MappingSourceConfig{Enabled: src.defaultEnabled}
	if src.urlFlag != "" {
		sc.BaseURL = src.defaultBaseURL
	}
	if src.defaultCacheDir != nil {
		sc.CacheDir = src.defaultCacheDir()
		sc.CacheMaxAge = src.defaultCacheMaxAge
	}
	return sc
}

// getConfigHelp returns a helpful message for creating config file.
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

// IsConfigNotFoundError checks if error is related to config file.
func IsConfigNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "config file not found") ||
		strings.Contains(errMsg, "no such file or directory") ||
		strings.Contains(errMsg, "failed to parse config")
}
