package main

import "log"

// Command-line flag names
const (
	FlagConfigFile = "c"
	FlagForceSync  = "f"
	FlagDryRun     = "d"
	FlagMangaSync  = "manga"
	FlagAllSync    = "all"
	FlagVerbose    = "verbose"
	FlagDirection  = "direction"
)

// Default values for command-line flags
const (
	DefaultConfigFile = "config.yaml"
	DefaultForceSync  = false
	DefaultDryRun     = false
	DefaultMangaSync  = false
	DefaultAllSync    = false
	DefaultVerbose    = false
	DefaultDirection  = DirectionAnilistToMAL
)

// Default values for config file options
const (
	DefaultOAuthPort          = "18080"
	DefaultOAuthRedirectURI   = "http://localhost:18080/callback"
	DefaultScoreNormalization = true
	DefaultTokenFilePath      = "$HOME/.config/anilist-mal-sync/token.json"
)

// SyncDirection represents the direction of synchronization
type SyncDirection int

const (
	SyncDirectionAnilistToMAL  SyncDirection = iota // AniList → MAL (default)
	SyncDirectionMALToAnilist                       // MAL → AniList (reverse)
	SyncDirectionBidirectional                      // Both directions
)

// MediaTypeOption represents what media types to sync
type MediaTypeOption int

const (
	MediaTypeAnimeOnly MediaTypeOption = iota // Sync anime only (default)
	MediaTypeMangaOnly                        // Sync manga only
	MediaTypeBoth                             // Sync both anime and manga
)

// SyncOptions holds all synchronization options
type SyncOptions struct {
	// Media type options
	MediaType MediaTypeOption

	// Sync direction
	Direction SyncDirection

	// Behavior flags
	ForceSync bool
	DryRun    bool
	Verbose   bool

	// Config-driven options
	ScoreNormalization bool
}

// NewSyncOptions creates SyncOptions from command-line flags and config
func NewSyncOptions(
	forceSync, dryRun, verbose, mangaSync, allSync bool,
	directionFlag string,
	config Config,
) SyncOptions {
	// Determine media type
	var mediaType MediaTypeOption
	if allSync {
		mediaType = MediaTypeBoth
	} else if mangaSync {
		mediaType = MediaTypeMangaOnly
	} else {
		mediaType = MediaTypeAnimeOnly
	}

	var direction SyncDirection
	switch directionFlag {
	case DirectionMALToAnilist:
		direction = SyncDirectionMALToAnilist
	case DirectionBoth:
		direction = SyncDirectionBidirectional
	case DirectionAnilistToMAL, "":
		direction = SyncDirectionAnilistToMAL
	default:
		log.Printf("Warning: Invalid direction '%s', defaulting to '%s'. Valid options: '%s', '%s', '%s'", directionFlag, DirectionAnilistToMAL, DirectionAnilistToMAL, DirectionMALToAnilist, DirectionBoth)
		direction = SyncDirectionAnilistToMAL
	}

	return SyncOptions{
		MediaType:          mediaType,
		Direction:          direction,
		ForceSync:          forceSync,
		DryRun:             dryRun,
		Verbose:            verbose,
		ScoreNormalization: *config.ScoreNormalization,
	}
}

// IsAnimeSync returns true if anime should be synced
func (o SyncOptions) IsAnimeSync() bool {
	return o.MediaType == MediaTypeAnimeOnly || o.MediaType == MediaTypeBoth
}

// IsMangaSync returns true if manga should be synced
func (o SyncOptions) IsMangaSync() bool {
	return o.MediaType == MediaTypeMangaOnly || o.MediaType == MediaTypeBoth
}

// IsReverseDirection returns true if syncing MAL -> AniList
func (o SyncOptions) IsReverseDirection() bool {
	return o.Direction == SyncDirectionMALToAnilist
}

// IsBidirectional returns true if syncing both directions
func (o SyncOptions) IsBidirectional() bool {
	return o.Direction == SyncDirectionBidirectional
}
