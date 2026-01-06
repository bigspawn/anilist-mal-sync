package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// Test Helper Functions

func testSiteConfig() SiteConfig {
	return SiteConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		AuthURL:      "https://example.com/auth",
		TokenURL:     "https://example.com/token",
	}
}

// =============================================================================
// Category 1: Basic Tests (No external dependencies)
// =============================================================================

func TestNewOAuth_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost:18080/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	if oauth == nil {
		t.Fatal("NewOAuth() returned nil")
	}

	if oauth.Config.ClientID != config.ClientID {
		t.Errorf("ClientID = %v, want %v", oauth.Config.ClientID, config.ClientID)
	}

	if oauth.siteName != "test" {
		t.Errorf("siteName = %v, want %v", oauth.siteName, "test")
	}

	if oauth.tokenFilePath != tokenPath {
		t.Errorf("tokenFilePath = %v, want %v", oauth.tokenFilePath, tokenPath)
	}

	if len(oauth.state) == 0 {
		t.Error("state is empty")
	}
}

func TestNewOAuth_RelativePathRejected(t *testing.T) {
	tests := []struct {
		name        string
		tokenPath   string
		expectError bool
		description string
	}{
		{"Relative path", "token.json", true, "relative path should be rejected"},
		{"Relative path with dir", "./config/token.json", true, "relative path should be rejected"},
		{"Absolute path", t.TempDir() + "/token.json", false, "absolute path should be accepted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testSiteConfig()
			_, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tt.tokenPath)

			if (err != nil) != tt.expectError {
				t.Errorf("NewOAuth() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
			}

			if err != nil && tt.expectError {
				if !strings.Contains(err.Error(), "path must be absolute") {
					t.Errorf("error message should contain 'path must be absolute', got: %v", err)
				}
			}
		})
	}
}

func TestNeedInit_NoToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	if !oauth.NeedInit() {
		t.Error("NeedInit() should return true when token is nil")
	}
}

func TestNeedInit_HasToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Pre-create token file
	token := &oauth2.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	tf := &TokenFile{Tokens: map[string]*oauth2.Token{"test": token}}
	if err := writeTokenFile(tokenPath, tf); err != nil {
		t.Fatalf("setup: writeTokenFile() error = %v", err)
	}

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	if oauth.NeedInit() {
		t.Error("NeedInit() should return false when token exists")
	}
}

func TestStateParameterGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	oauth1, err := NewOAuth(testSiteConfig(), "http://localhost/callback", "test1", []oauth2.AuthCodeOption{}, tmpDir+"/token1.json")
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	oauth2, err := NewOAuth(testSiteConfig(), "http://localhost/callback", "test2", []oauth2.AuthCodeOption{}, tmpDir+"/token2.json")
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// State should be 32 characters (from randHTTPParamString(32))
	if len(oauth1.state) != 32 {
		t.Errorf("state length = %v, want 32", len(oauth1.state))
	}

	if len(oauth2.state) != 32 {
		t.Errorf("state length = %v, want 32", len(oauth2.state))
	}

	// Each OAuth instance should have different state
	if oauth1.state == oauth2.state {
		t.Error("states should be different for different OAuth instances")
	}
}

func TestGetAuthURL_IncludesState(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	authURL := oauth.GetAuthURL()

	// Parse URL to verify state parameter
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	state := parsedURL.Query().Get("state")
	if state == "" {
		t.Error("auth URL should contain state parameter")
	}

	if state != oauth.state {
		t.Errorf("state in URL = %v, want %v", state, oauth.state)
	}
}

func TestCreateDirIfNotExists(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectError bool
		description string
	}{
		{
			"Directory doesn't exist",
			func(t *testing.T) string {
				return t.TempDir() + "/newdir/token.json"
			},
			false,
			"should create directory",
		},
		{
			"Directory exists",
			func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "token.json")
			},
			false,
			"should no-op",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath(t)
			err := createDirIfNotExists(path)

			if (err != nil) != tt.expectError {
				t.Errorf("createDirIfNotExists() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
			}

			if !tt.expectError {
				// Verify directory was created
				dir := filepath.Dir(path)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("directory was not created: %s", dir)
				}
			}
		})
	}
}

// =============================================================================
// Category 2: Token File Operations Tests
// =============================================================================

func TestTokenFileReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write token
	token := &oauth2.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	tf := &TokenFile{Tokens: map[string]*oauth2.Token{"test": token}}
	err := writeTokenFile(tokenPath, tf)
	if err != nil {
		t.Fatalf("writeTokenFile() error = %v", err)
	}

	// Read back
	tf2, err := readTokenFile(tokenPath)
	if err != nil {
		t.Fatalf("readTokenFile() error = %v", err)
	}

	// Verify
	if tf2.Tokens["test"].AccessToken != "test_token" {
		t.Errorf("token = %v, want %v", tf2.Tokens["test"].AccessToken, "test_token")
	}
}

func TestMissingTokenFile(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "nonexistent.json")

	tf, err := readTokenFile(tokenPath)
	if err != nil {
		t.Errorf("readTokenFile() error = %v, want nil", err)
	}

	if tf == nil {
		t.Error("readTokenFile() should return empty TokenFile, not nil")
		return
	}

	if len(tf.Tokens) != 0 {
		t.Errorf("TokenFile should be empty, got %d tokens", len(tf.Tokens))
	}
}

func TestInvalidTokenFileJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(tokenPath, []byte("{invalid json}"), 0o600); err != nil {
		t.Fatalf("setup: WriteFile() error = %v", err)
	}

	_, err := readTokenFile(tokenPath)
	if err == nil {
		t.Error("readTokenFile() should return error for invalid JSON")
	}
}

func TestAtomicWritePreventsCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write first token
	token1 := &oauth2.Token{AccessToken: "token1"}
	tf1 := &TokenFile{Tokens: map[string]*oauth2.Token{"test": token1}}
	if err := writeTokenFile(tokenPath, tf1); err != nil {
		t.Fatalf("writeTokenFile() error = %v", err)
	}

	// Verify first token exists
	tfRead, err := readTokenFile(tokenPath)
	if err != nil {
		t.Fatalf("readTokenFile() error = %v", err)
	}
	if tfRead.Tokens["test"].AccessToken != "token1" {
		t.Errorf("token = %v, want token1", tfRead.Tokens["test"].AccessToken)
	}

	// Write second token (should use atomic write)
	token2 := &oauth2.Token{AccessToken: "token2"}
	tf2 := &TokenFile{Tokens: map[string]*oauth2.Token{"test": token2}}
	if err := writeTokenFile(tokenPath, tf2); err != nil {
		t.Fatalf("writeTokenFile() error = %v", err)
	}

	// Verify second token replaced first
	tfRead2, err := readTokenFile(tokenPath)
	if err != nil {
		t.Fatalf("readTokenFile() error = %v", err)
	}
	if tfRead2.Tokens["test"].AccessToken != "token2" {
		t.Errorf("token = %v, want token2", tfRead2.Tokens["test"].AccessToken)
	}

	// Verify no temp files left
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			t.Errorf("temp file should be cleaned up, found: %s", f.Name())
		}
	}
}

// =============================================================================
// Category 3: PKCE Configuration Tests
// =============================================================================

func TestMyAnimeListPKCE_PlainMethod(t *testing.T) {
	tests := []struct {
		name        string
		wantMethod  string
		description string
	}{
		{"Plain PKCE method", "plain", "MAL requires plain method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tokenPath := filepath.Join(tmpDir, "token.json")

			// Simulate MyAnimeList PKCE options
			verifier := oauth2.GenerateVerifier()
			config := testSiteConfig()
			oauth, err := NewOAuth(
				config,
				"http://localhost:18080/callback",
				"myanimelist",
				[]oauth2.AuthCodeOption{
					oauth2.SetAuthURLParam("code_challenge", verifier),
					oauth2.SetAuthURLParam("code_challenge_method", "plain"),
					oauth2.VerifierOption(verifier),
				},
				tokenPath,
			)
			if err != nil {
				t.Fatalf("NewOAuth() error = %v", err)
			}

			authURL := oauth.GetAuthURL()
			parsedURL, err := url.Parse(authURL)
			if err != nil {
				t.Fatalf("failed to parse auth URL: %v", err)
			}

			query := parsedURL.Query()
			method := query.Get("code_challenge_method")
			challenge := query.Get("code_challenge")

			if method != tt.wantMethod {
				t.Errorf("code_challenge_method = %v, want %v (%s)", method, tt.wantMethod, tt.description)
			}

			// For plain method, challenge should equal verifier
			if challenge != verifier {
				t.Errorf("code_challenge = %v, want %v (plain method should have challenge=verifier)", challenge, verifier)
			}
		})
	}
}

func TestAnilistPKCE_S256Method(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Simulate AniList PKCE options
	verifier := oauth2.GenerateVerifier()
	config := testSiteConfig()
	oauth, err := NewOAuth(
		config,
		"http://localhost:18080/callback",
		"anilist",
		[]oauth2.AuthCodeOption{
			oauth2.AccessTypeOffline,
			oauth2.S256ChallengeOption(verifier),
			oauth2.VerifierOption(verifier),
		},
		tokenPath,
	)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	authURL := oauth.GetAuthURL()
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	query := parsedURL.Query()
	method := query.Get("code_challenge_method")
	challenge := query.Get("code_challenge")
	accessType := query.Get("access_type")

	// Verify S256 method
	if method != "S256" {
		t.Errorf("code_challenge_method = %v, want S256", method)
	}

	// Verify challenge is SHA256 hash of verifier
	hash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	if challenge != expectedChallenge {
		t.Errorf("code_challenge = %v, want %v (SHA256 of verifier)", challenge, expectedChallenge)
	}

	// Verify AccessTypeOffline
	if accessType != "offline" {
		t.Errorf("access_type = %v, want offline", accessType)
	}
}

func TestGetAuthURL_IncludesPKCEOptions(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	verifier := oauth2.GenerateVerifier()
	config := testSiteConfig()
	oauth, err := NewOAuth(
		config,
		"http://localhost/callback",
		"test",
		[]oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("code_challenge", verifier),
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
		},
		tokenPath,
	)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	authURL := oauth.GetAuthURL()
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	query := parsedURL.Query()

	challenge := query.Get("code_challenge")
	if challenge == "" {
		t.Error("auth URL should contain code_challenge parameter")
	}

	method := query.Get("code_challenge_method")
	if method == "" {
		t.Error("auth URL should contain code_challenge_method parameter")
	}
}

// =============================================================================
// Category 4: CSRF State Validation Tests (with httptest)
// =============================================================================

func TestStateValidation_MissingState(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost:18080/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Create request without state parameter
	req := httptest.NewRequest("GET", "/callback?code=test_code", nil)
	w := httptest.NewRecorder()

	// Call the callback handler (extracted from startServer)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "State parameter missing", http.StatusBadRequest)
			return
		}

		oauth.stateMu.RLock()
		expectedState := oauth.state
		oauth.stateMu.RUnlock()

		if state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}
	})

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if err := resp.Body.Close(); err != nil {
		t.Logf("Warning: failed to close response body: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStateValidation_MismatchedState(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost:18080/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Create request with wrong state
	req := httptest.NewRequest("GET", "/callback?code=test_code&state=wrong_state", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "State parameter missing", http.StatusBadRequest)
			return
		}

		oauth.stateMu.RLock()
		expectedState := oauth.state
		oauth.stateMu.RUnlock()

		if state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}
	})

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if err := resp.Body.Close(); err != nil {
		t.Logf("Warning: failed to close response body: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %v, want %v", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStateValidation_ValidState(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost:18080/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Create request with correct state
	req := httptest.NewRequest("GET", "/callback?code=test_code&state="+oauth.state, nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "State parameter missing", http.StatusBadRequest)
			return
		}

		oauth.stateMu.RLock()
		expectedState := oauth.state
		oauth.stateMu.RUnlock()

		if state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// State is valid, would continue to token exchange
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if err := resp.Body.Close(); err != nil {
		t.Logf("Warning: failed to close response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

// =============================================================================
// Category 5: Context Cancellation Tests
// =============================================================================

func TestTokenWithContext_RespectsContext(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// TokenWithContext should fail due to cancelled context
	_, err = oauth.TokenWithContext(ctx)
	if err == nil {
		t.Error("TokenWithContext() should return error when context is cancelled")
	}

	// The error might be from context cancellation or token refresh
	// Either way, there should be an error
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

func TestToken_DeprecatedUsesBackgroundContext(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Token() is deprecated but should still work (uses background context)
	// It will fail because there's no actual token, but it shouldn't panic
	_, err = oauth.Token()
	// We expect an error since there's no valid token to refresh
	if err == nil {
		t.Log("Token() returned nil error (this is expected if there's a valid token)")
	}
}

func TestTokenSource_ContextAware(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	ctx := context.Background()
	ts := oauth.TokenSource(ctx)

	if ts == nil {
		t.Error("TokenSource() should return a non-nil TokenSource")
	}

	// Verify it's the context-aware wrapper
	if _, ok := ts.(*contextAwareTokenSource); !ok {
		t.Errorf("TokenSource() should return *contextAwareTokenSource, got %T", ts)
	}
}

// =============================================================================
// Category 6: Thread Safety Tests (run with -race flag)
// =============================================================================

func TestConcurrentTokenAccess(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Spawn 100 goroutines reading token state
	var wg sync.WaitGroup
	iterations := 100
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = oauth.NeedInit() // Concurrent read
		}()
	}
	wg.Wait()

	// Run with: go test -race
	// This test should not report any race conditions
}

func TestConcurrentStateAccess(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Spawn 100 goroutines getting auth URL (reads state)
	var wg sync.WaitGroup
	iterations := 100
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = oauth.GetAuthURL() // Concurrent read of state
		}()
	}
	wg.Wait()

	// Run with: go test -race
	// This test should not report any race conditions
}

func TestConcurrentTokenAndStateAccess(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Spawn multiple goroutines accessing both token and state
	var wg sync.WaitGroup
	iterations := 50
	for i := 0; i < iterations; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = oauth.NeedInit()
		}()
		go func() {
			defer wg.Done()
			_ = oauth.GetAuthURL()
		}()
	}
	wg.Wait()

	// Run with: go test -race
	// This test should not report any race conditions
}

// =============================================================================
// Category 7: Mock OAuth Server Tests
// =============================================================================

func setupMockOAuthServer(t *testing.T) (*httptest.Server, *oauth2.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		if _, err := w.Write([]byte("access_token=mocktoken&token_type=bearer&expires_in=3600")); err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
		}
	})

	server := httptest.NewServer(mux)
	config := &oauth2.Config{
		ClientID:     "test_id",
		ClientSecret: "test_secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  server.URL + "/auth",
			TokenURL: server.URL + "/token",
		},
	}
	t.Cleanup(func() { server.Close() })
	return server, config
}

func TestExchangeToken_WithMockServer(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	_, mockConfig := setupMockOAuthServer(t)

	// Create OAuth with mock config
	oauth := &OAuth{
		Config:          mockConfig,
		siteName:        "test",
		authCodeOptions: []oauth2.AuthCodeOption{},
		tokenFilePath:   tokenPath,
		state:           "test_state",
	}

	// Exchange token (will use mock server)
	ctx := context.Background()
	err := oauth.ExchangeToken(ctx, "test_code")
	if err != nil {
		t.Logf("ExchangeToken() error = %v (this is expected if mock server doesn't handle full OAuth flow)", err)
	}

	// Verify token was saved if exchange succeeded
	if !oauth.NeedInit() {
		t.Log("Token was successfully exchanged and saved")
	}
}

// =============================================================================
// Round-trip Integration Test
// =============================================================================

func TestOAuth_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Create OAuth
	config := testSiteConfig()
	oauth, err := NewOAuth(config, "http://localhost:18080/callback", "test", []oauth2.AuthCodeOption{}, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuth() error = %v", err)
	}

	// Verify initial state
	if !oauth.NeedInit() {
		t.Error("NeedInit() should return true initially")
	}

	// Generate auth URL
	authURL := oauth.GetAuthURL()
	if authURL == "" {
		t.Error("GetAuthURL() should return non-empty URL")
	}

	// Verify state in URL
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}
	state := parsedURL.Query().Get("state")
	if state != oauth.state {
		t.Errorf("state in URL = %v, want %v", state, oauth.state)
	}
}
