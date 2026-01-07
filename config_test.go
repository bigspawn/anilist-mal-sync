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
myanimelist:
  client_id: "mal_id"
  client_secret: "default_mal_secret"
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

	// Should contain default path
	expectedPath := os.ExpandEnv("$HOME/.config/anilist-mal-sync/token.json")
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
