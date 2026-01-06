package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// colorPrintf is a helper for colored output to stdout
func colorPrintf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

type TokenFile struct {
	Tokens map[string]*oauth2.Token `json:"tokens"`
}

func NewTokenFile() *TokenFile {
	return &TokenFile{Tokens: make(map[string]*oauth2.Token)}
}

type OAuth struct {
	token           *oauth2.Token
	tokenMu         sync.RWMutex
	siteName        string
	authCodeOptions []oauth2.AuthCodeOption
	tokenFilePath   string
	state           string
	stateMu         sync.RWMutex

	Config *oauth2.Config
}

func NewOAuth(
	config SiteConfig,
	redirectURI string,
	siteName string,
	authCodeOptions []oauth2.AuthCodeOption,
	tokenFilePath string,
) (*OAuth, error) {
	if !path.IsAbs(tokenFilePath) {
		return nil, fmt.Errorf("path must be absolute: %s", tokenFilePath)
	}

	if err := createDirIfNotExists(tokenFilePath); err != nil {
		return nil, err
	}

	oauth := &OAuth{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  redirectURI,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
		siteName:        siteName,
		authCodeOptions: authCodeOptions,
		tokenFilePath:   tokenFilePath,
		state:           randHTTPParamString(32),
	}

	oauth.loadTokenFromFile()

	return oauth, nil
}

func (oauth *OAuth) GetAuthURL() string {
	oauth.stateMu.RLock()
	defer oauth.stateMu.RUnlock()
	return oauth.Config.AuthCodeURL(oauth.state, oauth.authCodeOptions...)
}

func (oauth *OAuth) ExchangeToken(ctx context.Context, code string) error {
	token, err := oauth.Config.Exchange(ctx, code, oauth.authCodeOptions...)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %w", err)
	}

	oauth.tokenMu.Lock()
	oauth.token = token
	oauth.tokenMu.Unlock()

	return oauth.saveTokenToFile()
}

func (oauth *OAuth) TokenSource(ctx context.Context) oauth2.TokenSource {
	oauth.tokenMu.RLock()
	defer oauth.tokenMu.RUnlock()

	// Create a context-aware token source that carries the context
	// through to Token() refreshes for proper cancellation support
	return &contextAwareTokenSource{
		oauth: oauth,
		ctx:   ctx,
	}
}

// contextAwareTokenSource wraps OAuth with a context for Token() calls.
// Storing context in struct is forced by oauth2.TokenSource interface which
// doesn't accept context in Token() method. This is necessary for proper
// context propagation during token refresh (commit 89f04a5).
type contextAwareTokenSource struct {
	oauth *OAuth
	ctx   context.Context //nolint:containedctx // forced by oauth2.TokenSource interface
}

func (s *contextAwareTokenSource) Token() (*oauth2.Token, error) {
	return s.oauth.TokenWithContext(s.ctx)
}

func (oauth *OAuth) Token() (*oauth2.Token, error) {
	// Deprecated: Use TokenWithContext for proper context propagation
	return oauth.TokenWithContext(context.Background())
}

func (oauth *OAuth) TokenWithContext(ctx context.Context) (*oauth2.Token, error) {
	oauth.tokenMu.Lock()
	defer oauth.tokenMu.Unlock()

	log.Printf("Refreshing token for %s", oauth.siteName)

	t, err := oauth.Config.TokenSource(ctx, oauth.token).Token()
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}

	log.Printf("Token refreshed for %s", oauth.siteName)

	oauth.token = t

	if err = oauth.saveTokenToFile(); err != nil {
		return nil, fmt.Errorf("error saving token: %w", err)
	}

	log.Printf("Token saved for %s", oauth.siteName)

	return t, nil
}

func (oauth *OAuth) NeedInit() bool {
	oauth.tokenMu.RLock()
	defer oauth.tokenMu.RUnlock()
	return oauth.token == nil
}

// InitToken starts the OAuth flow if token is not present.
// Returns error if context is cancelled during flow or token acquisition fails.
func (oauth *OAuth) InitToken(ctx context.Context, port string) error {
	if !oauth.NeedInit() {
		return nil // Token already exists
	}

	getToken(ctx, oauth, port)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if oauth.NeedInit() {
		return fmt.Errorf("failed to obtain token for %s", oauth.siteName)
	}

	return nil
}

// DeleteToken removes the token for this site from the token file.
func (oauth *OAuth) DeleteToken() error {
	tokenFile, err := readTokenFile(oauth.tokenFilePath)
	if err != nil {
		return fmt.Errorf("error reading token file: %w", err)
	}

	delete(tokenFile.Tokens, oauth.siteName)

	oauth.tokenMu.Lock()
	oauth.token = nil
	oauth.tokenMu.Unlock()

	return writeTokenFile(oauth.tokenFilePath, tokenFile)
}

// IsTokenValid checks if token exists and is not expired.
func (oauth *OAuth) IsTokenValid() bool {
	oauth.tokenMu.RLock()
	defer oauth.tokenMu.RUnlock()

	if oauth.token == nil {
		return false
	}

	// Token with zero expiry is considered always valid (some services don't provide expiry)
	if oauth.token.Expiry.IsZero() {
		return true
	}

	return oauth.token.Expiry.After(time.Now())
}

// TokenExpiry returns token expiry time or zero time if no token.
func (oauth *OAuth) TokenExpiry() time.Time {
	oauth.tokenMu.RLock()
	defer oauth.tokenMu.RUnlock()

	if oauth.token == nil {
		return time.Time{}
	}
	return oauth.token.Expiry
}

func (oauth *OAuth) loadTokenFromFile() {
	tokenFile, err := readTokenFile(oauth.tokenFilePath)
	if err != nil {
		log.Println("Error reading token file:", err)
		return
	}

	if token, exists := tokenFile.Tokens[oauth.siteName]; exists {
		log.Printf("Token loaded for %s", oauth.siteName)
		oauth.tokenMu.Lock()
		oauth.token = token
		oauth.tokenMu.Unlock()
	}
}

func (oauth *OAuth) saveTokenToFile() error {
	tokenFile, err := readTokenFile(oauth.tokenFilePath)
	if err != nil {
		log.Println("Error reading token file:", err)
		return fmt.Errorf("error reading token file: %w", err)
	}

	tokenFile.Tokens[oauth.siteName] = oauth.token

	return writeTokenFile(oauth.tokenFilePath, tokenFile)
}

func readTokenFile(tokenFilePath string) (*TokenFile, error) {
	// #nosec G304 - Token file path is user's config directory for OAuth tokens
	file, err := os.Open(tokenFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewTokenFile(), nil
		}
		return nil, fmt.Errorf("error opening token file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing token file: %v", err)
		}
	}()

	tokenFile := NewTokenFile()
	err = json.NewDecoder(file).Decode(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding token file: %w", err)
	}

	return tokenFile, nil
}

func writeTokenFile(tokenFilePath string, tokenFile *TokenFile) error {
	// Use atomic write pattern: write to temp file, then rename
	// This prevents partial writes and corruption if process crashes
	// #nosec G304 - Token file path is user's config directory for OAuth tokens
	dir := filepath.Dir(tokenFilePath)

	// Create temporary file in same directory (ensures same filesystem)
	tmpFile, err := os.CreateTemp(dir, "token*.tmp")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write to temp file
	if err := json.NewEncoder(tmpFile).Encode(tokenFile); err != nil {
		cleanupFile(tmpFile, tmpPath)
		return fmt.Errorf("error encoding token file: %w", err)
	}

	// Ensure data is flushed to disk before rename
	if err := tmpFile.Sync(); err != nil {
		cleanupFile(tmpFile, tmpPath)
		return fmt.Errorf("error syncing temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		// File already closed with error, just remove temp file
		if err := os.Remove(tmpPath); err != nil {
			log.Printf("Error removing temp file %s: %v", tmpPath, err)
		}
		return fmt.Errorf("error closing temp file: %w", err)
	}

	// Atomic rename (overwrites target if exists)
	if err := os.Rename(tmpPath, tokenFilePath); err != nil {
		return fmt.Errorf("error renaming temp file: %w", err)
	}

	return nil
}

// cleanupFile closes the file and removes it, logging any errors.
// In cleanup paths, we still log errors for observability.
func cleanupFile(f *os.File, path string) {
	if err := f.Close(); err != nil {
		log.Printf("Error closing temp file %s: %v", path, err)
	}
	if err := os.Remove(path); err != nil {
		log.Printf("Error removing temp file %s: %v", path, err)
	}
}

func startServer(oauth *OAuth, port string, done chan<- bool) *http.Server {
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		callbackCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Validate state parameter for CSRF protection
		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "State parameter missing", http.StatusBadRequest)
			log.Printf("State parameter missing in callback")
			return
		}

		oauth.stateMu.RLock()
		expectedState := oauth.state
		oauth.stateMu.RUnlock()

		if state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			log.Printf("State mismatch: expected=%s, got=%s", expectedState, state)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code parameter missing", http.StatusBadRequest)
			log.Printf("Code parameter missing in callback")
			return
		}

		err := oauth.ExchangeToken(callbackCtx, code)
		if err != nil {
			http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
			log.Printf("Error exchanging code for token: %v", err)
			return
		}

		if !oauth.NeedInit() {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)

			//nolint:lll //ok
			_, e := w.Write([]byte(`<html><body><h2>Authorization successful. You can close this window</h2>.<br><script>window.close();</script></body></html>`))
			if e != nil {
				log.Printf("Error writing response: %v", e)
				return
			}

			done <- true
		} else {
			http.Error(w, "Token not set", http.StatusInternalServerError)
			log.Printf("Token not set after exchange")
		}
	})

	server.Handler = mux

	go func() {
		log.Printf("Server started at http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v", err)
		}
		log.Println("Server stopped")
	}()

	// Color codes for URL highlighting
	const (
		colorReset = "\033[0m"
		colorBold  = "\033[1m"
		colorCyan  = "\033[36m"
		colorBlue  = "\033[34m"
	)

	authURL := oauth.GetAuthURL()
	colorPrintf("\n%s➜  Open the following URL in your browser:%s\n", colorBold+colorCyan, colorReset)
	colorPrintf("%s%s%s\n\n", colorBold+colorBlue, authURL, colorReset)

	return server
}

func getToken(ctx context.Context, oauth *OAuth, port string) {
	done := make(chan bool, 1)
	server := startServer(oauth, port, done)

	defer func(ctx context.Context) {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}(ctx)

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, exiting...")
		return
	case <-done:
		log.Println("OAuth flow completed successfully")
	}
}

func createDirIfNotExists(path string) error {
	path = filepath.Clean(path)
	dir := filepath.Dir(path)

	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		return nil
	}
	return fmt.Errorf("error checking directory: %w", err)
}
