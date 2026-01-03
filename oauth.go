package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

type TokenFile struct {
	Tokens map[string]*oauth2.Token `json:"tokens"`
}

func NewTokenFile() *TokenFile {
	return &TokenFile{Tokens: make(map[string]*oauth2.Token)}
}

type OAuth struct {
	token           *oauth2.Token
	siteName        string
	authCodeOptions []oauth2.AuthCodeOption
	tokenFilePath   string
	state           string // Random state for CSRF protection

	Config *oauth2.Config
}

func NewOAuth(
	config SiteConfig,
	redirectURI string,
	siteName string,
	authCodeOptions []oauth2.AuthCodeOption,
	tokenFilePath string,
) (*OAuth, error) {
	absPath, err := filepath.Abs(tokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("invalid token file path: %w", err)
	}

	if err := createDirIfNotExists(absPath); err != nil {
		return nil, err
	}

	state, err := randHTTPParamString(OAuthStateLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OAuth state parameter: %w", err)
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
		tokenFilePath:   absPath,
		state:           state, // Random state for CSRF protection
	}

	oauth.loadTokenFromFile()

	return oauth, nil
}

func (oauth *OAuth) GetAuthURL() string {
	return oauth.Config.AuthCodeURL(oauth.state, oauth.authCodeOptions...)
}

func (oauth *OAuth) ExchangeToken(ctx context.Context, code string) error {
	token, err := oauth.Config.Exchange(ctx, code, oauth.authCodeOptions...)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %w", err)
	}
	oauth.token = token
	return oauth.saveTokenToFile()
}

func (oauth *OAuth) TokenSource() oauth2.TokenSource {
	return oauth2.ReuseTokenSourceWithExpiry(oauth.token, oauth, 24*time.Hour)
}

// Token refreshes the OAuth token. Called by oauth2.ReuseTokenSource which doesn't
// provide a context, so Background is used (oauth2 library limitation).
func (oauth *OAuth) Token() (*oauth2.Token, error) {
	log.Printf("Refreshing token for %s", oauth.siteName)

	t, err := oauth.Config.TokenSource(context.Background(), oauth.token).Token()
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
	return oauth.token == nil
}

func (oauth *OAuth) loadTokenFromFile() {
	tokenFile, err := readTokenFile(oauth.tokenFilePath)
	if err != nil {
		log.Println("Error reading token file:", err)
		return
	}

	if token, exists := tokenFile.Tokens[oauth.siteName]; exists {
		log.Printf("Token loaded for %s", oauth.siteName)
		oauth.token = token
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
	// create file with restrictive permissions (0600) where possible
	file, err := os.OpenFile(tokenFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, TokenFilePerms)
	if err != nil {
		return fmt.Errorf("error creating token file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing token file: %v", err)
		}
	}()

	return json.NewEncoder(file).Encode(tokenFile)
}

func shutdownServer(ctx context.Context, server *http.Server) {
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		cancel()
		log.Printf("Error shutting down server: %v", err)
	}
	cancel()
	log.Println("Server shut down")
}

func startServer(ctx context.Context, oauth *OAuth, port string, done chan<- bool) {
	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Use request context with timeout for request-scoped operations
		reqCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Validate state parameter for CSRF protection
		state := r.URL.Query().Get("state")
		if state != oauth.state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			log.Printf("OAuth callback received invalid state parameter")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			log.Printf("OAuth callback missing authorization code")
			return
		}

		err := oauth.ExchangeToken(reqCtx, code)
		if err != nil {
			http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
			log.Printf("Error exchanging code for token: %v", err)
			return
		}

		if !oauth.NeedInit() {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)

			_, e := w.Write([]byte(`<html><body>Authorization successful. You can close this window.<br><script>window.close();</script></body></html>`))
			if e != nil {
				log.Printf("Error writing response: %v", e)
			}

			done <- true

			go shutdownServer(ctx, server)
		} else {
			http.Error(w, "Token not set", http.StatusInternalServerError)
			log.Printf("Token not set after exchange")
			return
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

	log.Println("Navigate to the following URL for authorization:", oauth.GetAuthURL())
}

func getToken(ctx context.Context, oauth *OAuth, port string) {
	done := make(chan bool, 1) // Buffered to prevent blocking if callback sends after context cancellation

	go startServer(ctx, oauth, port, done)

	select {
	case <-ctx.Done():
		// Context cancelled - server will continue running but will be cleaned up on process exit
		// This is acceptable for a one-time OAuth flow
		return
	case <-done:
		// OAuth flow completed successfully
		return
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
		if err = os.MkdirAll(dir, ConfigDirPerms); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		return nil
	}
	return fmt.Errorf("error checking directory: %w", err)
}
