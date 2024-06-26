package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("c", "config.yaml", "path to config file")
	forceSync := flag.Bool("f", false, "force sync all animes")
	dryRun := flag.Bool("d", false, "dry run without updating MyAnimeList")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config, err := loadConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	app, err := NewApp(ctx, config, *forceSync, *dryRun)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
