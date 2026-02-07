package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	aodGitHubAPIURL = "https://api.github.com/repos/manami-project/anime-offline-database/releases/latest"
	aodAssetName    = "anime-offline-database-minified.json"
	aodMetadataFile = "version.txt"
	aodDatabaseFile = "anime-offline-database.json"
)

// AODEntry represents a single anime entry from the offline database.
type AODEntry struct {
	Sources []string `json:"sources"`
	Title   string   `json:"title"`
	Type    string   `json:"type"`
}

// OfflineDatabase provides ID mapping between AniList and MyAnimeList
// using the anime-offline-database project.
type OfflineDatabase struct {
	malToAniList map[int]int
	anilistToMAL map[int]int
	lastUpdate   string
	entries      int
}

// GetAniListID returns the AniList ID for a given MAL ID.
func (db *OfflineDatabase) GetAniListID(malID int) (int, bool) {
	id, ok := db.malToAniList[malID]
	return id, ok
}

// GetMALID returns the MAL ID for a given AniList ID.
func (db *OfflineDatabase) GetMALID(anilistID int) (int, bool) {
	id, ok := db.anilistToMAL[anilistID]
	return id, ok
}

// LoadOfflineDatabase loads the offline database, downloading or updating as needed.
func LoadOfflineDatabase(ctx context.Context, cfg OfflineDatabaseConfig) (*OfflineDatabase, error) {
	if cfg.CacheDir == "" {
		cfg.CacheDir = getDefaultCacheDir()
	}

	dbPath := filepath.Join(cfg.CacheDir, aodDatabaseFile)
	metaPath := filepath.Join(cfg.CacheDir, aodMetadataFile)

	exists := fileExists(dbPath)

	if err := ensureDatabase(ctx, cfg, exists, dbPath, metaPath); err != nil {
		return nil, err
	}

	return parseAODFile(dbPath)
}

func ensureDatabase(ctx context.Context, cfg OfflineDatabaseConfig, exists bool, dbPath, metaPath string) error {
	needsDownload := cfg.ForceRefresh || !exists
	if needsDownload {
		return handleDownload(ctx, cfg.CacheDir, dbPath, metaPath, exists)
	}
	if cfg.AutoUpdate {
		updateIfNeeded(ctx, dbPath, metaPath)
	}
	return nil
}

func handleDownload(ctx context.Context, cacheDir, dbPath, metaPath string, exists bool) error {
	if err := downloadAndCache(ctx, cacheDir, dbPath, metaPath); err != nil {
		if !exists {
			return fmt.Errorf("download offline database: %w", err)
		}
		LogWarn(ctx, "Failed to download offline database: %v (using cached version)", err)
	}
	return nil
}

func downloadAndCache(ctx context.Context, cacheDir, dbPath, metaPath string) error {
	LogStage(ctx, "Downloading offline database...")

	downloadURL, tag, err := getLatestReleaseInfo(ctx)
	if err != nil {
		return fmt.Errorf("get latest release: %w", err)
	}

	// #nosec G301 - Cache directory for non-sensitive data
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	if err := downloadAODFile(ctx, downloadURL, dbPath); err != nil {
		return fmt.Errorf("download file: %w", err)
	}

	// #nosec G306 - Version metadata is non-sensitive
	if err := os.WriteFile(metaPath, []byte(tag), 0o600); err != nil {
		LogWarn(ctx, "Failed to save version metadata: %v", err)
	}

	LogInfoSuccess(ctx, "Downloaded offline database (version %s)", tag)
	return nil
}

func updateIfNeeded(ctx context.Context, dbPath, metaPath string) {
	cachedVersion, _ := getCachedVersion(metaPath)

	latestURL, latestTag, err := getLatestReleaseInfo(ctx)
	if err != nil {
		LogDebug(ctx, "Cannot check for offline database updates: %v (using cached version)", err)
		return
	}

	if cachedVersion == latestTag {
		LogDebug(ctx, "Offline database is up to date (version %s)", cachedVersion)
		return
	}

	LogInfo(ctx, "Updating offline database: %s → %s", cachedVersion, latestTag)

	if err := downloadAODFile(ctx, latestURL, dbPath); err != nil {
		LogWarn(ctx, "Failed to update offline database: %v (using cached version)", err)
		return
	}

	if err := os.WriteFile(metaPath, []byte(latestTag), 0o600); err != nil {
		LogWarn(ctx, "Failed to save version metadata: %v", err)
	}
}

// downloadAODFile downloads the file to a temp location and atomically renames it.
func downloadAODFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(destPath), "aod-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write file: %w", err)
	}
	_ = tmpFile.Close()

	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}

// parseAODFile parses the offline database JSON file using streaming decoder
// and builds ID mapping indices.
func parseAODFile(filePath string) (*OfflineDatabase, error) {
	// #nosec G304 - File path comes from controlled cache directory
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only file

	db := &OfflineDatabase{
		malToAniList: make(map[int]int),
		anilistToMAL: make(map[int]int),
	}

	decoder := json.NewDecoder(f)

	// Read opening brace
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("read opening token: %w", err)
	}

	for decoder.More() {
		// Read field name
		tok, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("read field token: %w", err)
		}

		key, ok := tok.(string)
		if !ok {
			continue
		}

		switch key {
		case "data":
			if err := parseDataArray(decoder, db); err != nil {
				return nil, err
			}
		case "lastUpdate":
			var lastUpdate string
			if err := decoder.Decode(&lastUpdate); err != nil {
				return nil, fmt.Errorf("decode lastUpdate: %w", err)
			}
			db.lastUpdate = lastUpdate
		default:
			// Skip unknown fields
			var raw json.RawMessage
			if err := decoder.Decode(&raw); err != nil {
				return nil, fmt.Errorf("skip field %s: %w", key, err)
			}
		}
	}

	return db, nil
}

func parseDataArray(decoder *json.Decoder, db *OfflineDatabase) error {
	// Read opening bracket of data array
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("read data array start: %w", err)
	}

	for decoder.More() {
		var entry AODEntry
		if err := decoder.Decode(&entry); err != nil {
			return fmt.Errorf("decode entry: %w", err)
		}

		indexEntry(entry, db)
	}

	// Read closing bracket
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("read data array end: %w", err)
	}

	return nil
}

func indexEntry(entry AODEntry, db *OfflineDatabase) {
	var malID, anilistID int

	for _, src := range entry.Sources {
		if id, ok := extractIDFromURL(src, "https://myanimelist.net/anime/"); ok {
			malID = id
		}
		if id, ok := extractIDFromURL(src, "https://anilist.co/anime/"); ok {
			anilistID = id
		}
	}

	if malID > 0 && anilistID > 0 {
		db.malToAniList[malID] = anilistID
		db.anilistToMAL[anilistID] = malID
		db.entries++
	}
}

// extractIDFromURL extracts a numeric ID from a URL with the given prefix.
// Example: extractIDFromURL("https://myanimelist.net/anime/1535", "https://myanimelist.net/anime/") returns (1535, true)
func extractIDFromURL(url, prefix string) (int, bool) {
	if !strings.HasPrefix(url, prefix) {
		return 0, false
	}

	idStr := strings.TrimPrefix(url, prefix)
	// Handle URLs with trailing path segments (e.g., "/anime/1535/some-title")
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// GitHub release API response (minimal fields)
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// getLatestReleaseInfo queries the GitHub API for the latest release download URL and tag.
func getLatestReleaseInfo(ctx context.Context) (downloadURL, tag string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, aodGitHubAPIURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best effort close

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github API status: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == aodAssetName {
			return asset.BrowserDownloadURL, release.TagName, nil
		}
	}

	return "", "", fmt.Errorf("asset %s not found in release %s", aodAssetName, release.TagName)
}

// getCachedVersion reads the cached version tag from the metadata file.
func getCachedVersion(metadataPath string) (string, error) {
	// #nosec G304 - File path comes from controlled cache directory
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getDefaultCacheDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", "aod-cache")
}

// BuildFromEntries builds an OfflineDatabase from a slice of AODEntry (used in tests).
func BuildFromEntries(entries []AODEntry) *OfflineDatabase {
	db := &OfflineDatabase{
		malToAniList: make(map[int]int),
		anilistToMAL: make(map[int]int),
	}
	for _, entry := range entries {
		indexEntry(entry, db)
	}
	return db
}
