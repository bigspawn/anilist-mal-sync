package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v3"
)

const (
	ServiceAnilist     = "anilist"
	ServiceMyAnimeList = "myanimelist"
	ServiceAll         = "all"
)

// Flag names for the ID-mapping sources. Each name is spelled in three places —
// the shared sync flags, the root command's own flag list, and the config
// override — so a literal would drift.
// Flag names shared between newSyncFlags and getSyncFlagsFromCmd/applySyncFlagsToConfig.
const (
	flagForce            = "force"
	flagDryRun           = "dry-run"
	flagManga            = "manga"
	flagAllMedia         = "all"
	flagVerbose          = "verbose"
	flagReverseDirection = "reverse-direction"
	flagFavorites        = "favorites"
)

const (
	flagARMAPI          = "arm-api"
	flagARMAPIURL       = "arm-api-url"
	flagHatoAPI         = "hato-api"
	flagHatoAPIURL      = "hato-api-url"
	flagMangaBakaAPI    = "mangabaka-api"
	flagMangaBakaAPIURL = "mangabaka-api-url"
	flagJikanAPI        = "jikan-api"

	flagOfflineDB             = "offline-db"
	flagOfflineDBForceRefresh = "offline-db-force-refresh"
)

// syncFlags are the common flags shared between sync and watch commands.
var syncFlags = newSyncFlags(false)

// newSyncFlags builds the sync/watch flag set. local marks each flag Local,
// which NewCLI's root-command copy needs and the sync/watch commands don't;
// building both from one function is what keeps them from drifting apart.
func newSyncFlags(local bool) []cli.Flag {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    flagForce,
			Aliases: []string{"f"},
			Usage:   "force sync all entries",
			Local:   local,
		},
		&cli.BoolFlag{
			Name:    flagDryRun,
			Aliases: []string{"d"},
			Usage:   "dry run without updating target service",
			Local:   local,
		},
		&cli.BoolFlag{
			Name:  flagManga,
			Usage: "sync manga instead of anime",
			Local: local,
		},
		&cli.BoolFlag{
			Name:  flagAllMedia,
			Usage: "sync all anime and manga",
			Local: local,
		},
		&cli.BoolFlag{
			Name:  flagVerbose,
			Usage: "enable verbose logging",
			Local: local,
		},
		&cli.BoolFlag{
			Name:  flagReverseDirection,
			Usage: "sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
			Local: local,
		},
		&cli.BoolFlag{
			Name:  flagOfflineDB,
			Usage: "enable offline database for anime ID mapping (ignored for --manga) (default: true)",
			Value: true,
			Local: local,
		},
		&cli.BoolFlag{
			Name:  flagOfflineDBForceRefresh,
			Usage: "force re-download offline database",
			Local: local,
		},
	}

	for _, src := range mappingSources() {
		flags = append(flags, &cli.BoolFlag{
			Name:  src.flag,
			Usage: src.usage,
			Value: src.defaultEnabled,
			Local: local,
		})
		if src.urlFlag != "" {
			flags = append(flags, &cli.StringFlag{
				Name:  src.urlFlag,
				Usage: src.urlUsage,
				Local: local,
			})
		}
	}

	flags = append(flags, &cli.BoolFlag{
		Name:  flagFavorites,
		Usage: "sync favorites between services (requires Jikan API for MAL favorites)",
		Local: local,
	})

	return flags
}

// getSyncFlagsFromCmd extracts sync flags, updates package-level globals,
// and returns verbose and reverse values explicitly.
func getSyncFlagsFromCmd(cmd *cli.Command) (verboseOut bool, reverseOut bool) {
	forceVal := cmd.Bool(flagForce)
	dryVal := cmd.Bool(flagDryRun)
	mangaVal := cmd.Bool(flagManga)
	allVal := cmd.Bool(flagAllMedia)
	verboseVal := cmd.Bool(flagVerbose)
	reverseVal := cmd.Bool(flagReverseDirection)

	forceSync = &forceVal
	dryRun = &dryVal
	mangaSync = &mangaVal
	allSync = &allVal
	verbose = &verboseVal

	return verboseVal, reverseVal
}

// applySyncFlagsToConfig applies CLI sync flag overrides to config.
func applySyncFlagsToConfig(cmd *cli.Command, cfg *Config) {
	if cmd.IsSet(flagOfflineDB) {
		cfg.OfflineDatabase.Enabled = cmd.Bool(flagOfflineDB)
	}
	if cmd.IsSet(flagOfflineDBForceRefresh) && cmd.Bool(flagOfflineDBForceRefresh) {
		cfg.OfflineDatabase.ForceRefresh = true
	}
	for _, src := range mappingSources() {
		sc := src.configField(cfg)
		if cmd.IsSet(src.flag) {
			sc.Enabled = cmd.Bool(src.flag)
		}
		if src.urlFlag != "" && cmd.IsSet(src.urlFlag) {
			if v := cmd.String(src.urlFlag); v != "" {
				sc.BaseURL = v
			}
		}
	}
	if cmd.IsSet(flagFavorites) {
		cfg.Favorites.Enabled = cmd.Bool(flagFavorites)
		// Favorites sync requires Jikan API to read MAL favorites
		if cfg.Favorites.Enabled {
			cfg.JikanAPI.Enabled = true
		}
	}
}

// NewCLI creates the root CLI command.
func NewCLI() *cli.Command {
	// Define flags for backward compatibility with old CLI behavior
	configFlag := &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "path to config file (optional, uses env vars if not specified)",
	}

	// The root command runs sync when no subcommand is given, so it needs its
	// own copy of every sync flag, marked Local so cli/v3 doesn't also expose
	// it as a global flag on subcommands. newSyncFlags(true) is the same
	// table-driven build as syncFlags, just Local instead of shared.
	flags := append([]cli.Flag{configFlag}, newSyncFlags(true)...)

	return &cli.Command{
		Name:        "anilist-mal-sync",
		Usage:       "Synchronize anime and manga lists between AniList and MyAnimeList",
		Version:     version,
		Description: "Sync your anime/manga lists between AniList and MyAnimeList.",
		Flags:       flags,
		Commands: []*cli.Command{
			newLoginCommand(),
			newLogoutCommand(),
			newStatusCommand(),
			newSyncCommand(),
			newWatchCommand(),
			newUnmappedCommand(),
		},
		// Default action when no command specified - runs sync for backward compatibility
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// If there are positional arguments (unknown command), show help
			if cmd.Args().Present() {
				return fmt.Errorf("unknown command: %s", cmd.Args().First())
			}
			return runSync(ctx, cmd)
		},
	}
}

// RunCLI executes the CLI application.
func RunCLI() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := NewCLI()

	// Run and show help only for CLI usage errors
	err := cmd.Run(ctx, os.Args)
	if err != nil {
		// Show help only for CLI usage errors (unknown command, invalid flags)
		// Don't show help for runtime errors (network, API, etc.)
		if IsCLIUsageError(err) {
			fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
			//nolint:gosec // G104: best effort help display
			cli.ShowAppHelp(cmd) //nolint:errcheck // best effort help display
		} else if !IsConfigNotFoundError(err) && !IsCancellationError(err) {
			// For other errors, just print the error message
			fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
		}
		return errors.New("command failed")
	}

	return nil
}

// IsCancellationError checks if error is due to context cancellation (e.g., Ctrl+C).
func IsCancellationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled)
}

// IsCLIUsageError checks if error is related to incorrect CLI usage.
func IsCLIUsageError(err error) bool {
	if err == nil {
		return false
	}
	// Check for unknown command error (from Action at line 80)
	// CLI usage errors typically start with "unknown command:" or flag errors
	// Runtime errors typically contain "run app:", "error syncing", "error getting", etc.
	errMsg := err.Error()
	if strings.HasPrefix(errMsg, "unknown command:") {
		return true
	}
	// If error contains runtime error indicators, it's not a CLI usage error
	runtimeIndicators := []string{
		"run app:",
		"error syncing",
		"error getting",
		"error loading",
		"error creating",
		"context deadline exceeded",
		"connection refused",
		"no such host",
	}
	for _, indicator := range runtimeIndicators {
		if strings.Contains(errMsg, indicator) {
			return false
		}
	}
	return false
}
