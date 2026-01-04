package main

import "time"

// Media type constants
const (
	MediaTypeAnime = "anime"
	MediaTypeManga = "manga"
)

// API limits and pagination
const (
	MALListLimit         = 100 // Maximum items per page for MAL API
	MALSearchLimit       = 3   // Maximum search results for MAL API
	AnilistSearchPage    = 1   // Page number for AniList searches
	AnilistSearchPerPage = 10  // Results per page for AniList searches
)

// File permissions
const (
	TokenFilePerms = 0o600 // Read/write for owner only
	ConfigDirPerms = 0o750 // Read/write/execute for owner, read/execute for group
)

// Service names for logging
const (
	ServiceAnilist = "AniList"
	ServiceMAL     = "MAL"
)

// OAuth constants
const (
	MALOAuthCodeLength = 43 // Length of random string for MAL OAuth code challenge/verifier
	OAuthStateLength   = 32 // Length of random string for OAuth state parameter (CSRF protection)
)

// Environment variable names
const (
	EnvVarPort                    = "PORT"
	EnvVarClientSecretAnilist     = "CLIENT_SECRET_ANILIST"
	EnvVarClientSecretMyAnimeList = "CLIENT_SECRET_MYANIMELIST"
)

// Timeout and duration constants
const (
	HTTPClientTimeout     = 10 * time.Minute // HTTP client timeout for API requests
	ServerShutdownTimeout = 5 * time.Second  // Timeout for graceful server shutdown
	RequestTimeout        = 5 * time.Second  // Timeout for individual HTTP requests
	ReadHeaderTimeout     = 10 * time.Second // Timeout for reading HTTP headers
	TokenExpiryDuration   = 24 * time.Hour   // Token expiry duration for OAuth token source
)

// Backoff policy constants
const (
	BackoffInitialInterval     = 1 * time.Second  // Initial backoff interval
	BackoffMaxInterval         = 30 * time.Second // Maximum backoff interval
	BackoffMaxElapsedTime      = 2 * time.Minute  // Maximum elapsed time for backoff
	BackoffMultiplier          = 2.0              // Backoff multiplier
	BackoffRandomizationFactor = 0.1              // Randomization factor for jitter
)

// Title matching thresholds
const (
	TitleSimilarityThreshold  = 98.0  // Minimum similarity percentage for fuzzy matching
	TitleLevenshteinThreshold = 98.0  // Minimum Levenshtein similarity percentage
	PerfectMatchThreshold     = 100.0 // Perfect match threshold (100%)
	NoMatchThreshold          = 0.0   // No match threshold (0%)
	PercentMultiplier         = 100.0 // Multiplier to convert ratio to percentage
)

// Sync direction string constants
const (
	DirectionAnilistToMAL = "anilist-to-mal"
	DirectionMALToAnilist = "mal-to-anilist"
	DirectionBoth         = "both"
)

// AniList API status string constants
const (
	AnilistStatusCurrent   = "CURRENT"
	AnilistStatusCompleted = "COMPLETED"
	AnilistStatusPaused    = "PAUSED"
	AnilistStatusDropped   = "DROPPED"
	AnilistStatusPlanning  = "PLANNING"
	AnilistStatusRepeating = "REPEATING"
)

// OAuth site name constants (used for token storage keys)
const (
	SiteNameAnilist     = "anilist"
	SiteNameMyAnimeList = "myanimelist"
)

// Updater prefix constants (for logging and identification)
const (
	UpdaterPrefixAnilistToMALAnime = "AniList to MAL Anime"
	UpdaterPrefixAnilistToMALManga = "AniList to MAL Manga"
	UpdaterPrefixMALToAnilistAnime = "MAL to AniList Anime"
	UpdaterPrefixMALToAnilistManga = "MAL to AniList Manga"
)
