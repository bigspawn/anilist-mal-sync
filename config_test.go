package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFromFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a valid config file
	configContent := `
oauth:
  port: "18080"
  redirect_uri: "http://localhost:18080/callback"
anilist:
  client_id: "test_id"
  client_secret: "test_secret"
  auth_url: "https://anilist.co/api/v2/oauth/authorize"
  token_url: "https://anilist.co/api/v2/oauth/token"
  username: "test_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "mal_secret"
  auth_url: "https://myanimelist.net/v1/oauth2/authorize"
  token_url: "https://myanimelist.net/v1/oauth2/token"
  username: "test_user"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile() error = %v", err)
	}

	// Verify OAuth settings
	if config.OAuth.Port != "18080" {
		t.Errorf("OAuth.Port = %v, want 18080", config.OAuth.Port)
	}

	// Verify AniList settings
	if config.Anilist.ClientID != "test_id" {
		t.Errorf("Anilist.ClientID = %v, want test_id", config.Anilist.ClientID)
	}

	// Verify MyAnimeList settings
	if config.MyAnimeList.ClientID != "mal_id" {
		t.Errorf("MyAnimeList.ClientID = %v, want mal_id", config.MyAnimeList.ClientID)
	}

	// Verify default token file path
	if config.TokenFilePath == "" {
		t.Error("TokenFilePath should not be empty")
	}
}

func TestLoadConfigFromFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	_, err := loadConfigFromFile(configPath)
	if err == nil {
		t.Error("loadConfigFromFile() should return error when file not found")
	}

	// Verify error is recognized as config error
	if !IsConfigNotFoundError(err) {
		t.Error("Error should be recognized as config not found error")
	}
}

func TestLoadConfigFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Create invalid YAML
	if err := os.WriteFile(configPath, []byte("{invalid yaml content"), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := loadConfigFromFile(configPath)
	if err == nil {
		t.Error("loadConfigFromFile() should return error for invalid YAML")
	}

	// Verify error is recognized as config error
	if !IsConfigNotFoundError(err) {
		t.Error("Error should be recognized as config parse error")
	}
}

func TestLoadConfigFromFile_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test_id"
  client_secret: "default_secret"
  username: "ani_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "default_mal_secret"
  username: "mal_user"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variables
	t.Setenv("PORT", "9999")
	t.Setenv("CLIENT_SECRET_ANILIST", "env_secret")
	t.Setenv("CLIENT_SECRET_MYANIMELIST", "env_mal_secret")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile() error = %v", err)
	}

	// Verify env override for port
	if config.OAuth.Port != "9999" {
		t.Errorf("OAuth.Port = %v, want 9999 (env override)", config.OAuth.Port)
	}

	// Verify env override for secrets
	if config.Anilist.ClientSecret != "env_secret" {
		t.Errorf("Anilist.ClientSecret = %v, want env_secret (env override)", config.Anilist.ClientSecret)
	}

	if config.MyAnimeList.ClientSecret != "env_mal_secret" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want env_mal_secret (env override)", config.MyAnimeList.ClientSecret)
	}
}

func TestLoadConfigFromFile_EnvOverride_MAL_USERNAME(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file WITHOUT myanimelist.username
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test_id"
  client_secret: "default_secret"
  username: "ani_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "default_mal_secret"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set MAL_USERNAME via env
	t.Setenv("MAL_USERNAME", "env_mal_user")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile() error = %v", err)
	}

	if config.MyAnimeList.Username != "env_mal_user" {
		t.Errorf("MyAnimeList.Username = %v, want env_mal_user (env override)", config.MyAnimeList.Username)
	}
}

func TestLoadConfigFromFile_MissingMALUsername(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config file WITHOUT myanimelist.username
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test_id"
  client_secret: "test_secret"
  username: "test_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "mal_secret"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := loadConfigFromFile(configPath)
	if err == nil {
		t.Error("loadConfigFromFile() should return error when MAL_USERNAME is missing")
	}

	if !strings.Contains(err.Error(), "myanimelist.username") {
		t.Errorf("Error should mention myanimelist.username, got: %v", err)
	}
}

func TestGetConfigHelp_ReturnsValidString(t *testing.T) {
	help := getConfigHelp("config.yaml")

	if help == "" {
		t.Error("getConfigHelp() should return non-empty string")
	}

	// Verify key phrases are present
	keyPhrases := []string{
		"Configuration file not found",
		"To fix this",
		"Copy the example config",
		"config.example.yaml",
		"nano",
	}

	for _, phrase := range keyPhrases {
		if !strings.Contains(help, phrase) {
			t.Errorf("Help message should contain %q", phrase)
		}
	}
}

func TestIsConfigNotFoundError_ConfigNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"No error", nil, false},
		{"Random error", os.ErrClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConfigNotFoundError(tt.err)
			if got != tt.want {
				t.Errorf("IsConfigNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test actual config not found error separately
	t.Run("Config not found error", func(t *testing.T) {
		_, err := loadConfigFromFile("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error from loadConfigFromFile")
		}
		if !IsConfigNotFoundError(err) {
			t.Error("Config not found error should be recognized")
		}
	})
}

func TestLoadConfigFromFile_DefaultTokenPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config without token_file_path
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test_id"
  client_secret: "test_secret"
  auth_url: "https://anilist.co/api/v2/oauth/authorize"
  token_url: "https://anilist.co/api/v2/oauth/token"
  username: "test_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "mal_secret"
  auth_url: "https://myanimelist.net/v1/oauth2/authorize"
  token_url: "https://myanimelist.net/v1/oauth2/token"
  username: "test_user"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile() error = %v", err)
	}

	// Verify default token path is set
	if config.TokenFilePath == "" {
		t.Error("TokenFilePath should be set to default value")
	}

	// Should contain default path from getDefaultTokenPath()
	expectedPath, err := getDefaultTokenPath()
	if err != nil {
		t.Fatalf("getDefaultTokenPath() failed: %v", err)
	}
	if config.TokenFilePath != expectedPath {
		t.Errorf("TokenFilePath = %v, want %v", config.TokenFilePath, expectedPath)
	}
}

func TestLoadConfigFromFile_CustomTokenPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	customTokenPath := filepath.Join(tmpDir, "custom_tokens.json")

	// Create config with custom token path
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test_id"
  client_secret: "test_secret"
  auth_url: "https://anilist.co/api/v2/oauth/authorize"
  token_url: "https://anilist.co/api/v2/oauth/token"
  username: "test_user"
myanimelist:
  client_id: "mal_id"
  client_secret: "mal_secret"
  auth_url: "https://myanimelist.net/v1/oauth2/authorize"
  token_url: "https://myanimelist.net/v1/oauth2/token"
  username: "test_user"
token_file_path: "` + customTokenPath + `"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFromFile() error = %v", err)
	}

	if config.TokenFilePath != customTokenPath {
		t.Errorf("TokenFilePath = %v, want %v", config.TokenFilePath, customTokenPath)
	}
}

// =============================================================================
// WatchConfig Tests
// =============================================================================

func TestWatchConfig_GetInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		want     string // "0" for zero, "duration" for non-zero
		wantErr  bool
	}{
		{"Valid interval", "24h", "24h", false},
		{"Valid interval hours", "12h", "12h", false},
		{"Valid interval minutes", "90m", "90m", false},
		{"Empty interval", "", "0", false},
		{"Invalid interval", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := WatchConfig{Interval: tt.interval}
			got, err := w.GetInterval()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInterval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == "0" && got != 0 {
					t.Errorf("GetInterval() = %v, want 0", got)
				}
				if tt.want != "0" && got == 0 {
					t.Errorf("GetInterval() = 0, want non-zero")
				}
			}
		})
	}
}

func TestWatchConfig_GetInterval_Empty(t *testing.T) {
	w := WatchConfig{Interval: ""}
	got, err := w.GetInterval()
	if err != nil {
		t.Fatalf("GetInterval() unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("GetInterval() = %v, want 0", got)
	}
}

func TestWatchConfig_GetInterval_Invalid(t *testing.T) {
	w := WatchConfig{Interval: "not-a-duration"}
	_, err := w.GetInterval()
	if err == nil {
		t.Fatal("GetInterval() expected error, got nil")
	}
}

func TestLoadConfigFromEnv_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name           string
		setAnilistID   bool
		setAnilistUser bool
		setMALID       bool
		setMALUser     bool
		wantErr        bool
	}{
		{"All valid", true, true, true, true, false},
		{"Missing MAL_USERNAME", true, true, true, false, true},
		{"Missing ANILIST_USERNAME", true, false, true, true, true},
		{"Missing MAL_CLIENT_ID", true, true, false, true, true},
		{"Missing ANILIST_CLIENT_ID", false, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			t.Setenv("ANILIST_CLIENT_ID", "")
			t.Setenv("ANILIST_USERNAME", "")
			t.Setenv("MAL_CLIENT_ID", "")
			t.Setenv("MAL_USERNAME", "")

			if tt.setAnilistID {
				t.Setenv("ANILIST_CLIENT_ID", "test_id")
			}
			if tt.setAnilistUser {
				t.Setenv("ANILIST_USERNAME", "ani_user")
			}
			if tt.setMALID {
				t.Setenv("MAL_CLIENT_ID", "mal_id")
			}
			if tt.setMALUser {
				t.Setenv("MAL_USERNAME", "mal_user")
			}

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() failed: %v", err)
			}

			// Manually validate like loadConfigFromFile does
			err = validateConfig(cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultTokenPath(t *testing.T) {
	path, err := getDefaultTokenPath()
	if err != nil {
		t.Fatalf("getDefaultTokenPath() failed: %v", err)
	}

	if path == "" {
		t.Error("getDefaultTokenPath() returned empty string")
	}

	// Verify path contains expected components
	if !filepath.IsAbs(path) {
		t.Errorf("getDefaultTokenPath() returned non-absolute path: %s", path)
	}

	// Verify path ends with expected filename
	expectedSuffix := filepath.Join("anilist-mal-sync", "token.json")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("getDefaultTokenPath() = %s, expected suffix %s", path, expectedSuffix)
	}

	// Verify parent directory can be obtained without error
	dir := filepath.Dir(path)
	if dir == "" {
		t.Error("filepath.Dir(path) returned empty string")
	}

	// Verify the token filename is correct
	filename := filepath.Base(path)
	if filename != "token.json" {
		t.Errorf("getDefaultTokenPath() filename = %s, expected token.json", filename)
	}
}

func TestLoadConfigFromEnv_DefaultTokenPath(t *testing.T) {
	// Clear TOKEN_FILE_PATH to test default path
	t.Setenv("TOKEN_FILE_PATH", "")

	// Set required env vars
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Verify token path is set and uses default
	if cfg.TokenFilePath == "" {
		t.Error("loadConfigFromEnv() TokenFilePath is empty")
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(cfg.TokenFilePath) {
		t.Errorf("loadConfigFromEnv() TokenFilePath is not absolute: %s", cfg.TokenFilePath)
	}

	// Verify it contains the expected suffix
	expectedSuffix := filepath.Join("anilist-mal-sync", "token.json")
	if !strings.HasSuffix(cfg.TokenFilePath, expectedSuffix) {
		t.Errorf("loadConfigFromEnv() TokenFilePath = %s, expected suffix %s", cfg.TokenFilePath, expectedSuffix)
	}
}

func TestLoadConfigFromEnv_CustomTokenPath(t *testing.T) {
	customPath := "/custom/path/tokens.json"
	t.Setenv("TOKEN_FILE_PATH", customPath)

	// Set required env vars
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Verify custom path is used
	if cfg.TokenFilePath != customPath {
		t.Errorf("loadConfigFromEnv() TokenFilePath = %s, expected %s", cfg.TokenFilePath, customPath)
	}
}

// =============================================================================
// Suite 1: OAuth URL Defaults
// =============================================================================

func TestLoadConfigFromEnv_OAuthURLsHaveDefaults(t *testing.T) {
	// Set required env vars
	t.Setenv("ANILIST_CLIENT_ID", "test_anilist_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "test_mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Verify AniList OAuth URLs
	expectedAnilistAuthURL := "https://anilist.co/api/v2/oauth/authorize" // #nosec G101
	expectedAnilistTokenURL := "https://anilist.co/api/v2/oauth/token"    // #nosec G101

	if cfg.Anilist.AuthURL != expectedAnilistAuthURL {
		t.Errorf("Anilist.AuthURL = %v, want %v", cfg.Anilist.AuthURL, expectedAnilistAuthURL)
	}
	if cfg.Anilist.TokenURL != expectedAnilistTokenURL {
		t.Errorf("Anilist.TokenURL = %v, want %v", cfg.Anilist.TokenURL, expectedAnilistTokenURL)
	}

	// Verify MAL OAuth URLs
	expectedMALAuthURL := "https://myanimelist.net/v1/oauth2/authorize" // #nosec G101
	expectedMALTokenURL := "https://myanimelist.net/v1/oauth2/token"    // #nosec G101

	if cfg.MyAnimeList.AuthURL != expectedMALAuthURL {
		t.Errorf("MyAnimeList.AuthURL = %v, want %v", cfg.MyAnimeList.AuthURL, expectedMALAuthURL)
	}
	if cfg.MyAnimeList.TokenURL != expectedMALTokenURL {
		t.Errorf("MyAnimeList.TokenURL = %v, want %v", cfg.MyAnimeList.TokenURL, expectedMALTokenURL)
	}
}

// =============================================================================
// Suite 2: Required Fields - Single Missing
// =============================================================================

func TestLoadConfigFromEnv_MissingAnilistClientID(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	err = validateConfig(cfg)
	if err == nil {
		t.Error("validateConfig() should return error when ANILIST_CLIENT_ID is missing")
	}
}

func TestLoadConfigFromEnv_MissingAnilistUsername(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	err = validateConfig(cfg)
	if err == nil {
		t.Error("validateConfig() should return error when ANILIST_USERNAME is missing")
	}
}

func TestLoadConfigFromEnv_MissingMALClientID(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "")
	t.Setenv("MAL_USERNAME", "mal_user")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	err = validateConfig(cfg)
	if err == nil {
		t.Error("validateConfig() should return error when MAL_CLIENT_ID is missing")
	}
}

func TestLoadConfigFromEnv_MissingMALUsername(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	err = validateConfig(cfg)
	if err == nil {
		t.Error("validateConfig() should return error when MAL_USERNAME is missing")
	}
}

// =============================================================================
// Suite 3: Required Fields - Multiple Missing
// =============================================================================

func TestLoadConfigFromEnv_AllRequiredFieldsMissing(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "")
	t.Setenv("ANILIST_USERNAME", "")
	t.Setenv("MAL_CLIENT_ID", "")
	t.Setenv("MAL_USERNAME", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	err = validateConfig(cfg)
	if err == nil {
		t.Error("validateConfig() should return error when all required fields are missing")
	}
}

func TestLoadConfigFromEnv_ThreeRequiredFieldsMissing(t *testing.T) {
	tests := []struct {
		name           string
		setAnilistID   bool
		setAnilistUser bool
		setMALID       bool
		setMALUser     bool
	}{
		{"Only ANILIST_CLIENT_ID set", true, false, false, false},
		{"Only ANILIST_USERNAME set", false, true, false, false},
		{"Only MAL_CLIENT_ID set", false, false, true, false},
		{"Only MAL_USERNAME set", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ANILIST_CLIENT_ID", "")
			t.Setenv("ANILIST_USERNAME", "")
			t.Setenv("MAL_CLIENT_ID", "")
			t.Setenv("MAL_USERNAME", "")

			if tt.setAnilistID {
				t.Setenv("ANILIST_CLIENT_ID", "test_id")
			}
			if tt.setAnilistUser {
				t.Setenv("ANILIST_USERNAME", "test_user")
			}
			if tt.setMALID {
				t.Setenv("MAL_CLIENT_ID", "mal_id")
			}
			if tt.setMALUser {
				t.Setenv("MAL_USERNAME", "mal_user")
			}

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() failed: %v", err)
			}

			err = validateConfig(cfg)
			if err == nil {
				t.Error("validateConfig() should return error when 3 of 4 required fields are missing")
			}
		})
	}
}

func TestLoadConfigFromEnv_TwoRequiredFieldsMissing(t *testing.T) {
	tests := []struct {
		name           string
		setAnilistID   bool
		setAnilistUser bool
		setMALID       bool
		setMALUser     bool
	}{
		{"ANILIST_CLIENT_ID + ANILIST_USERNAME missing", false, false, true, true},
		{"ANILIST_CLIENT_ID + MAL_CLIENT_ID missing", false, true, false, true},
		{"ANILIST_CLIENT_ID + MAL_USERNAME missing", false, true, true, false},
		{"ANILIST_USERNAME + MAL_CLIENT_ID missing", true, false, false, true},
		{"ANILIST_USERNAME + MAL_USERNAME missing", true, false, true, false},
		{"MAL_CLIENT_ID + MAL_USERNAME missing", true, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ANILIST_CLIENT_ID", "")
			t.Setenv("ANILIST_USERNAME", "")
			t.Setenv("MAL_CLIENT_ID", "")
			t.Setenv("MAL_USERNAME", "")

			if tt.setAnilistID {
				t.Setenv("ANILIST_CLIENT_ID", "test_id")
			}
			if tt.setAnilistUser {
				t.Setenv("ANILIST_USERNAME", "test_user")
			}
			if tt.setMALID {
				t.Setenv("MAL_CLIENT_ID", "mal_id")
			}
			if tt.setMALUser {
				t.Setenv("MAL_USERNAME", "mal_user")
			}

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() failed: %v", err)
			}

			err = validateConfig(cfg)
			if err == nil {
				t.Error("validateConfig() should return error when 2 of 4 required fields are missing")
			}
		})
	}
}

func TestLoadConfigFromEnv_OneSiteConfigMissing(t *testing.T) {
	tests := []struct {
		name           string
		setAnilistID   bool
		setAnilistUser bool
		setMALID       bool
		setMALUser     bool
		expectError    bool
	}{
		{"All AniList missing (MAL set)", false, false, true, true, true},
		{"All MAL missing (AniList set)", true, true, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ANILIST_CLIENT_ID", "")
			t.Setenv("ANILIST_USERNAME", "")
			t.Setenv("MAL_CLIENT_ID", "")
			t.Setenv("MAL_USERNAME", "")

			if tt.setAnilistID {
				t.Setenv("ANILIST_CLIENT_ID", "test_id")
			}
			if tt.setAnilistUser {
				t.Setenv("ANILIST_USERNAME", "test_user")
			}
			if tt.setMALID {
				t.Setenv("MAL_CLIENT_ID", "mal_id")
			}
			if tt.setMALUser {
				t.Setenv("MAL_USERNAME", "mal_user")
			}

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() failed: %v", err)
			}

			err = validateConfig(cfg)
			if (err != nil) != tt.expectError {
				t.Errorf("validateConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// =============================================================================
// Suite 4: Optional Fields
// =============================================================================

func TestLoadConfigFromEnv_NoOptionalFields(t *testing.T) {
	// Set only required fields
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Clear optional fields
	t.Setenv("ANILIST_CLIENT_SECRET", "")
	t.Setenv("MAL_CLIENT_SECRET", "")
	t.Setenv("OAUTH_PORT", "")
	t.Setenv("OAUTH_REDIRECT_URI", "")
	t.Setenv("WATCH_INTERVAL", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Verify defaults
	if cfg.OAuth.Port != "18080" {
		t.Errorf("OAuth.Port = %v, want 18080", cfg.OAuth.Port)
	}
	if cfg.OAuth.RedirectURI != "http://localhost:18080/callback" {
		t.Errorf("OAuth.RedirectURI = %v, want http://localhost:18080/callback", cfg.OAuth.RedirectURI)
	}
	if cfg.Anilist.ClientSecret != "" {
		t.Errorf("Anilist.ClientSecret = %v, want empty", cfg.Anilist.ClientSecret)
	}
	if cfg.MyAnimeList.ClientSecret != "" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want empty", cfg.MyAnimeList.ClientSecret)
	}
	if cfg.Watch.Interval != "" {
		t.Errorf("Watch.Interval = %v, want empty", cfg.Watch.Interval)
	}
	// TokenFilePath should have default value
	if cfg.TokenFilePath == "" {
		t.Error("TokenFilePath should have default value")
	}
}

func TestLoadConfigFromEnv_AllOptionalFieldsSet(t *testing.T) {
	// Set required fields
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set all optional fields
	t.Setenv("ANILIST_CLIENT_SECRET", "anilist_secret")
	t.Setenv("MAL_CLIENT_SECRET", "mal_secret")
	t.Setenv("OAUTH_PORT", "9999")
	t.Setenv("OAUTH_REDIRECT_URI", "http://example.com/callback")
	t.Setenv("WATCH_INTERVAL", "12h")
	t.Setenv("TOKEN_FILE_PATH", "/custom/path/tokens.json")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Verify all values match env vars
	if cfg.Anilist.ClientSecret != "anilist_secret" {
		t.Errorf("Anilist.ClientSecret = %v, want anilist_secret", cfg.Anilist.ClientSecret)
	}
	if cfg.MyAnimeList.ClientSecret != "mal_secret" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want mal_secret", cfg.MyAnimeList.ClientSecret)
	}
	if cfg.OAuth.Port != "9999" {
		t.Errorf("OAuth.Port = %v, want 9999", cfg.OAuth.Port)
	}
	if cfg.OAuth.RedirectURI != "http://example.com/callback" {
		t.Errorf("OAuth.RedirectURI = %v, want http://example.com/callback", cfg.OAuth.RedirectURI)
	}
	if cfg.Watch.Interval != "12h" {
		t.Errorf("Watch.Interval = %v, want 12h", cfg.Watch.Interval)
	}
	if cfg.TokenFilePath != "/custom/path/tokens.json" {
		t.Errorf("TokenFilePath = %v, want /custom/path/tokens.json", cfg.TokenFilePath)
	}
}

func TestLoadConfigFromEnv_AnilistClientSecretOptional(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")
	t.Setenv("ANILIST_CLIENT_SECRET", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Should succeed with empty ClientSecret
	err = validateConfig(cfg)
	if err != nil {
		t.Errorf("validateConfig() should succeed without ANILIST_CLIENT_SECRET: %v", err)
	}
	if cfg.Anilist.ClientSecret != "" {
		t.Errorf("Anilist.ClientSecret = %v, want empty", cfg.Anilist.ClientSecret)
	}
}

func TestLoadConfigFromEnv_MALClientSecretOptional(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")
	t.Setenv("MAL_CLIENT_SECRET", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Should succeed with empty ClientSecret
	err = validateConfig(cfg)
	if err != nil {
		t.Errorf("validateConfig() should succeed without MAL_CLIENT_SECRET: %v", err)
	}
	if cfg.MyAnimeList.ClientSecret != "" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want empty", cfg.MyAnimeList.ClientSecret)
	}
}

// =============================================================================
// Suite 5: Edge Cases
// =============================================================================

func TestLoadConfigFromEnv_PORTPrecedence(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set both PORT and OAUTH_PORT
	t.Setenv("PORT", "9999")
	t.Setenv("OAUTH_PORT", "8888")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// OAUTH_PORT takes precedence (getEnvOrDefault checks OAUTH_PORT first)
	if cfg.OAuth.Port != "8888" {
		t.Errorf("OAuth.Port = %v, want 8888 (OAUTH_PORT takes precedence)", cfg.OAuth.Port)
	}
}

func TestLoadConfigFromEnv_OnlyPORTSet(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set only PORT (not OAUTH_PORT)
	t.Setenv("PORT", "7777")
	t.Setenv("OAUTH_PORT", "")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	if cfg.OAuth.Port != "7777" {
		t.Errorf("OAuth.Port = %v, want 7777", cfg.OAuth.Port)
	}
}

func TestLoadConfigFromEnv_LegacyAnilistSecretFallback(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set only legacy env var (not the new one)
	t.Setenv("ANILIST_CLIENT_SECRET", "")
	t.Setenv("CLIENT_SECRET_ANILIST", "legacy_secret")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// loadConfigFromEnv() does NOT support legacy fallback
	// It only reads from ANILIST_CLIENT_SECRET directly
	if cfg.Anilist.ClientSecret != "" {
		t.Errorf("Anilist.ClientSecret = %v, want empty (loadConfigFromEnv doesn't support legacy fallback)", cfg.Anilist.ClientSecret)
	}
}

func TestLoadConfigFromEnv_LegacyMALSecretFallback(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set only legacy env var (not the new one)
	t.Setenv("MAL_CLIENT_SECRET", "")
	t.Setenv("CLIENT_SECRET_MYANIMELIST", "legacy_secret")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// loadConfigFromEnv() does NOT support legacy fallback
	// It only reads from MAL_CLIENT_SECRET directly
	if cfg.MyAnimeList.ClientSecret != "" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want empty (loadConfigFromEnv doesn't support legacy fallback)", cfg.MyAnimeList.ClientSecret)
	}
}

func TestLoadConfigFromEnv_NewSecretPrecedence(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	// Set only new env vars (legacy not set)
	t.Setenv("ANILIST_CLIENT_SECRET", "new_secret")
	t.Setenv("MAL_CLIENT_SECRET", "new_mal_secret")

	cfg, err := loadConfigFromEnv()
	if err != nil {
		t.Fatalf("loadConfigFromEnv() failed: %v", err)
	}

	// Should use the new value
	if cfg.Anilist.ClientSecret != "new_secret" {
		t.Errorf("Anilist.ClientSecret = %v, want new_secret", cfg.Anilist.ClientSecret)
	}
	if cfg.MyAnimeList.ClientSecret != "new_mal_secret" {
		t.Errorf("MyAnimeList.ClientSecret = %v, want new_mal_secret", cfg.MyAnimeList.ClientSecret)
	}
}

func TestFavoritesSyncEnabled_FromEnv(t *testing.T) {
	t.Setenv("ANILIST_CLIENT_ID", "test_id")
	t.Setenv("ANILIST_USERNAME", "test_user")
	t.Setenv("MAL_CLIENT_ID", "mal_id")
	t.Setenv("MAL_USERNAME", "mal_user")

	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"Disabled", "false", false},
		{"Enabled", "true", true},
		{"Empty", "", false},
		{"Zero", "0", false},
		{"One", "1", true},
		{"Yes", "yes", true},
		{"No", "no", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FAVORITES_SYNC_ENABLED", tt.envValue)

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() failed: %v", err)
			}

			if cfg.Favorites.Enabled != tt.want {
				t.Errorf("Favorites.Enabled = %v, want %v", cfg.Favorites.Enabled, tt.want)
			}
		})
	}
}
