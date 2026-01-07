package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	minInterval     = 1 * time.Hour
	maxInterval     = 168 * time.Hour // 7 days
	defaultInterval = 24 * time.Hour
)

func newWatchCommand() *cli.Command {
	return &cli.Command{
		Name:  "watch",
		Usage: "Run sync on interval (Docker-friendly)",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:    "interval",
				Aliases: []string{"i"},
				Usage:   "Sync interval (1h-168h, default: 24h)",
				Value:   defaultInterval,
			},
			&cli.BoolFlag{
				Name:  "once",
				Usage: "Sync immediately then start watching",
			},
		},
		Action: runWatch,
	}
}

func runWatch(ctx context.Context, cmd *cli.Command) error {
	// Load config for compatibility with sync
	configPath := cmd.String("config")
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Create app for compatibility
	app, err := NewApp(ctx, config)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	interval := cmd.Duration("interval")

	// Validate interval
	if interval < minInterval {
		return fmt.Errorf("interval must be at least 1h (got %v)", interval)
	}
	if interval > maxInterval {
		return fmt.Errorf("interval must be at most 168h/7days (got %v)", interval)
	}

	// Optional immediate sync
	if cmd.Bool("once") {
		log.Printf("Running initial sync (--once flag set)...")
		if err := app.Run(ctx); err != nil {
			return fmt.Errorf("initial sync failed: %w", err)
		}
		log.Printf("Initial sync completed, starting watch mode")
	} else {
		log.Printf("Starting watch mode: sync every %v", interval)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Running scheduled sync...")
			if err := app.Run(ctx); err != nil {
				log.Printf("Sync error: %v", err)
			} else {
				log.Printf("Sync completed")
			}
		case <-ctx.Done():
			log.Printf("Watch mode stopped")
			return nil
		}
	}
}
