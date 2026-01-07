package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func newSyncCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Synchronize anime/manga lists between services",
		Flags: []cli.Flag{
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
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "enable verbose logging",
			},
			&cli.BoolFlag{
				Name:  "reverse-direction",
				Usage: "sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
			},
		},
		Action: runSync,
	}
}

func runSync(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")

	// Set package-level vars for compatibility with existing code
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

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	app, err := NewApp(ctx, config)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
