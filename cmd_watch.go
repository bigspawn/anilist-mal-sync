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
	watchFlags := append([]cli.Flag{
		&cli.DurationFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "Sync interval (1h-168h, required via --interval or config)",
		},
		&cli.BoolFlag{
			Name:  "once",
			Usage: "Sync immediately then start watching",
		},
	}, syncFlags...)

	return &cli.Command{
		Name:   "watch",
		Usage:  "Run sync on interval (Docker-friendly)",
		Flags:  watchFlags,
		Action: runWatch,
	}
}

func runWatch(ctx context.Context, cmd *cli.Command) error {
	// Set package-level vars for compatibility with existing code
	verboseVal := setSyncFlagsFromCmd(cmd)

	// Load config for compatibility with sync
	configPath := cmd.String("config")
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	applySyncFlagsToConfig(cmd, &config)

	// Initialize logger and add to context
	logger := NewLogger(verboseVal)
	ctx = logger.WithContext(ctx)

	// Create app for compatibility
	app, err := NewApp(ctx, config)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	interval := cmd.Duration("interval")

	// Priority: CLI flag > Config > Error
	if interval == 0 {
		cfgInterval, err := config.Watch.GetInterval()
		if err != nil {
			return fmt.Errorf("invalid interval in config: %w", err)
		}
		if cfgInterval == 0 {
			return fmt.Errorf("interval required (use --interval or set watch.interval in config)")
		}
		interval = cfgInterval
	}

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
		nextTime := time.Now().Add(interval)
		log.Printf("Initial sync completed, starting watch mode - next sync in %v at %s", interval, nextTime.Format("2006-01-02 15:04:05"))
	} else {
		nextTime := time.Now().Add(interval)
		log.Printf("Starting watch mode: next sync in %v at %s", interval, nextTime.Format("2006-01-02 15:04:05"))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Running scheduled sync...")
			app.Refresh(ctx)
			if err := app.Run(ctx); err != nil {
				log.Printf("Sync error: %v", err)
			} else {
				nextTime := time.Now().Add(interval)
				log.Printf("Sync completed - next sync in %v at %s", interval, nextTime.Format("2006-01-02 15:04:05"))
			}
		case <-ctx.Done():
			log.Printf("Watch mode stopped")
			return nil
		}
	}
}
