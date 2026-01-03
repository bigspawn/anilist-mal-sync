package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
)

var (
	configFile = flag.String(FlagConfigFile, DefaultConfigFile, "path to config file")
	forceSync  = flag.Bool(FlagForceSync, DefaultForceSync, "force sync all entries")
	dryRun     = flag.Bool(FlagDryRun, DefaultDryRun, "dry run without making any changes")
	mangaSync  = flag.Bool(FlagMangaSync, DefaultMangaSync, "sync manga instead of anime")
	allSync    = flag.Bool(FlagAllSync, DefaultAllSync, "sync both anime and manga")
	verbose    = flag.Bool(FlagVerbose, DefaultVerbose, "enable verbose logging")
	direction  = flag.String(
		FlagDirection,
		DefaultDirection,
		"sync direction: 'anilist-to-mal' (AniList → MAL, default), 'mal-to-anilist' (MAL → AniList), or 'both' (bidirectional)",
	)
)

func main() {
	flag.Parse()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	config, err := Load(*configFile)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	app, err := NewApp(ctx, config, *forceSync, *dryRun, *verbose, *mangaSync, *allSync, *direction)
	if err != nil {
		log.Fatalf("error creating app: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("error running app: %v", err)
	}
}
