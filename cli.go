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

// syncFlags are the common flags shared between sync and watch commands
var syncFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "force sync all entries",
	},
	&cli.BoolFlag{
		Name:    "dry-run",
		Aliases: []string{"d"},
		Usage:   "dry run without updating target service",
	},
	&cli.BoolFlag{
		Name:  "manga",
		Usage: "sync manga instead of anime",
	},
	&cli.BoolFlag{
		Name:  "all",
		Usage: "sync all anime and manga",
	},
	&cli.BoolFlag{
		Name:  "verbose",
		Usage: "enable verbose logging",
	},
	&cli.BoolFlag{
		Name:  "reverse-direction",
		Usage: "sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
	},
	&cli.BoolFlag{
		Name:  "offline-db",
		Usage: "enable offline database for anime ID mapping (ignored for --manga)",
		Value: true,
	},
	&cli.BoolFlag{
		Name:  "offline-db-force-refresh",
		Usage: "force re-download offline database",
	},
	&cli.BoolFlag{
		Name:  "arm-api",
		Usage: "enable ARM API for anime ID mapping (ignored for --manga, fallback after offline DB)",
	},
	&cli.StringFlag{
		Name:  "arm-api-url",
		Usage: "ARM API base URL",
	},
}

// setSyncFlagsFromCmd sets global sync variables from command flags and returns verbose value
func setSyncFlagsFromCmd(cmd *cli.Command) bool {
	forceVal := cmd.Bool("force")
	dryVal := cmd.Bool("dry-run")
	mangaVal := cmd.Bool("manga")
	allVal := cmd.Bool("all")
	verboseVal := cmd.Bool("verbose")
	reverseVal := cmd.Bool("reverse-direction")

	forceSync = &forceVal
	dryRun = &dryVal
	mangaSync = &mangaVal
	allSync = &allVal
	verbose = &verboseVal
	reverseDirection = &reverseVal

	return verboseVal
}

// applySyncFlagsToConfig applies CLI sync flag overrides to config.
func applySyncFlagsToConfig(cmd *cli.Command, cfg *Config) {
	if cmd.IsSet("offline-db") {
		cfg.OfflineDatabase.Enabled = cmd.Bool("offline-db")
	}
	if cmd.IsSet("offline-db-force-refresh") && cmd.Bool("offline-db-force-refresh") {
		cfg.OfflineDatabase.ForceRefresh = true
	}
	if cmd.IsSet("arm-api") {
		cfg.ARMAPI.Enabled = cmd.Bool("arm-api")
	}
	if cmd.IsSet("arm-api-url") {
		if v := cmd.String("arm-api-url"); v != "" {
			cfg.ARMAPI.BaseURL = v
		}
	}
}

// NewCLI creates the root CLI command
func NewCLI() *cli.Command {
	// Define flags for backward compatibility with old CLI behavior
	configFlag := &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "path to config file (optional, uses env vars if not specified)",
	}
	forceSyncFlag := &cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "force sync all entries",
	}
	dryRunFlag := &cli.BoolFlag{
		Name:    "dry-run",
		Aliases: []string{"d"},
		Usage:   "dry run without updating target service",
	}
	mangaSyncFlag := &cli.BoolFlag{
		Name:  "manga",
		Usage: "sync manga instead of anime",
	}
	allSyncFlag := &cli.BoolFlag{
		Name:  "all",
		Usage: "sync all anime and manga",
	}
	verboseFlag := &cli.BoolFlag{
		Name:  "verbose",
		Usage: "enable verbose logging",
	}
	reverseDirectionFlag := &cli.BoolFlag{
		Name:  "reverse-direction",
		Usage: "sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
	}
	offlineDbFlag := &cli.BoolFlag{
		Name:  "offline-db",
		Usage: "enable offline database for ID mapping (default: true)",
		Value: true,
	}
	offlineDbForceRefreshFlag := &cli.BoolFlag{
		Name:  "offline-db-force-refresh",
		Usage: "force re-download offline database",
	}
	armAPIFlag := &cli.BoolFlag{
		Name:  "arm-api",
		Usage: "enable ARM API for ID mapping (fallback after offline DB)",
	}
	armAPIURLFlag := &cli.StringFlag{
		Name:  "arm-api-url",
		Usage: "ARM API base URL",
	}

	return &cli.Command{
		Name:        "anilist-mal-sync",
		Usage:       "Synchronize anime and manga lists between AniList and MyAnimeList",
		Version:     "1.0.0",
		Description: "Sync your anime/manga lists between AniList and MyAnimeList.",
		Flags: []cli.Flag{
			configFlag,
			forceSyncFlag,
			dryRunFlag,
			mangaSyncFlag,
			allSyncFlag,
			verboseFlag,
			reverseDirectionFlag,
			offlineDbFlag,
			offlineDbForceRefreshFlag,
			armAPIFlag,
			armAPIURLFlag,
		},
		Commands: []*cli.Command{
			newLoginCommand(),
			newLogoutCommand(),
			newStatusCommand(),
			newSyncCommand(),
			newWatchCommand(),
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

// RunCLI executes the CLI application
func RunCLI() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := NewCLI()

	// Run and show help only for CLI usage errors
	if err := cmd.Run(ctx, os.Args); err != nil {
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
		return fmt.Errorf("command failed")
	}

	return nil
}

// IsCancellationError checks if error is due to context cancellation (e.g., Ctrl+C)
func IsCancellationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled)
}

// IsCLIUsageError checks if error is related to incorrect CLI usage
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
