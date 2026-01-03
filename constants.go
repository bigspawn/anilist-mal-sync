package main

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
	OAuthStateLength    = 32 // Length of random string for OAuth state parameter (CSRF protection)
)

// DefaultIgnoreAnimeTitles are anime titles that don't exist in MAL and should be skipped
var DefaultIgnoreAnimeTitles = []string{
	"scott pilgrim takes off",
	"bocchi the rock! recap part 2",
}

// DefaultIgnoreMangaTitles are manga titles that don't exist in MAL and should be skipped
var DefaultIgnoreMangaTitles = []string{}
