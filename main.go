package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
)

var (
	configFile       = flag.String("c", "config.yaml", "path to config file")
	forceSync        = flag.Bool("f", false, "force sync all animes")
	dryRun           = flag.Bool("d", false, "dry run without updating MyAnimeList")
	mangaSync        = flag.Bool("manga", false, "sync manga instead of anime")
	allSync          = flag.Bool("all", false, "sync all animes and mangas")
	verbose          = flag.Bool("verbose", false, "enable verbose logging")
	reverseDirection = flag.Bool(
		"reverse-direction",
		false,
		"sync from MyAnimeList to AniList (default is AniList to MyAnimeList)",
	)
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	config, err := loadConfigFromFile(*configFile)
	if err != nil {
		cancel()
		log.Fatalf("error: %v", err)
	}

	app, err := NewApp(ctx, config, *forceSync, *dryRun, *verbose)
	if err != nil {
		cancel()
		log.Fatalf("create app: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		cancel()
		log.Fatalf("run app: %v", err)
	}
	cancel()
}
