package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

const (
	ServiceAnilist     = "anilist"
	ServiceMyAnimeList = "myanimelist"
	ServiceAll         = "all"
)

// NewCLI creates the root CLI command
func NewCLI() *cli.Command {
	// Define flags for backward compatibility with old CLI behavior
	configFlag := &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "path to config file",
		Value:   "config.yaml",
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
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "enable verbose logging",
	}
	reverseDirectionFlag := &cli.BoolFlag{
		Name:  "reverse-direction",
		Usage: "sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
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
		},
		Commands: []*cli.Command{
			newLoginCommand(),
			newLogoutCommand(),
			newStatusCommand(),
			newSyncCommand(),
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

	// Run and show help on error
	if err := cmd.Run(ctx, os.Args); err != nil {
		// Show help on any error
		fmt.Fprintf(os.Stderr, "\nError: %v\n\n", err)
		//nolint:gosec // G104: best effort help display
		cli.ShowAppHelp(cmd) //nolint:errcheck // best effort help display
		return fmt.Errorf("command failed")
	}

	return nil
}
