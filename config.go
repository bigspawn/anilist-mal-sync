package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return Config{}, fmt.Errorf("invalid config path %q: %w", filename, err)
	}

	// Validate that the config file is within the working directory or the user's
	// config directory to avoid unexpected file inclusion.
	cwd, _ := os.Getwd()
	uconf, _ := os.UserConfigDir()
	// Ensure the absolute filename is under either the working directory or the
	// user's config directory. Use HasPrefix on cleaned absolute paths so this
	// is easier for static analysis to reason about.
	cleanedAbs := filepath.Clean(absFilename)
	cwdClean := filepath.Clean(cwd)
	uconfClean := filepath.Clean(uconf)
	sep := string(os.PathSeparator)
	underCwd := cleanedAbs == cwdClean || strings.HasPrefix(cleanedAbs, cwdClean+sep)
	underUconf := cleanedAbs == uconfClean || strings.HasPrefix(cleanedAbs, uconfClean+sep)
	if !underCwd && !underUconf {
		return Config{}, fmt.Errorf("config path %q is outside the working or user config directories", absFilename)
	}

	data, err := os.ReadFile(absFilename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %q: %w", absFilename, err)
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

	// Resolve token file path to a safe location under the user's config directory.
	// If the config provides a path, expand environment variables and validate it.
	resolved, err := resolveConfigPath(cfg.TokenFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("invalid token file path: %w", err)
	}
	cfg.TokenFilePath = resolved

	if cfg.ScoreNormalization == nil {
		v := DefaultScoreNormalization
		cfg.ScoreNormalization = &v
	}

	return cfg, nil
}

// resolveConfigPath returns a safe, absolute path for token/config files.
// If `provided` is empty the default path under the user's config directory
// is returned (e.g. $XDG_CONFIG_HOME/anilist-mal-sync/token.json).
// If `provided` is set it will be expanded and must reside under the
// user's config directory to be accepted.
func resolveConfigPath(provided string) (string, error) {
	if provided != "" {
		p := os.ExpandEnv(provided)
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", fmt.Errorf("invalid path: %w", err)
		}

		uconf, err := os.UserConfigDir()
		if err != nil {
			// If we can't get user config dir, still return the absolute path
			return abs, nil
		}
		rel, err := filepath.Rel(uconf, abs)
		if err != nil {
			return "", fmt.Errorf("failed to validate path: %w", err)
		}
		if strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("path %s is outside the user config directory %s", abs, uconf)
		}
		return abs, nil
	}

	uconf, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine user config dir: %w", err)
	}
	dir := filepath.Join(uconf, "anilist-mal-sync")
	return filepath.Join(dir, "token.json"), nil
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
