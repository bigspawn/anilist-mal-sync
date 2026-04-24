package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v3"
)

const defaultInterval = 24 * time.Hour

// cronSchedule mirrors cron.Schedule for testability without importing the library in tests.
type cronSchedule interface {
	Next(time.Time) time.Time
}

// syncRunner abstracts App methods needed by watch functions for testability.
type syncRunner interface {
	Run(context.Context) error
	Refresh(context.Context)
}

func newWatchCommand() *cli.Command {
	watchFlags := append([]cli.Flag{
		&cli.DurationFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Usage:   "Sync interval (1h-168h, required via --interval or config)",
		},
		&cli.StringFlag{
			Name:    "schedule",
			Aliases: []string{"s"},
			Usage:   "Cron expression (e.g. '0 3 * * *'). Mutually exclusive with --interval",
		},
		&cli.BoolFlag{
			Name:  "once",
			Usage: "Sync immediately then start watching",
		},
	}, syncFlags...)

	return &cli.Command{
		Name:   "watch",
		Usage:  "Run sync on interval or cron schedule (Docker-friendly)",
		Flags:  watchFlags,
		Action: runWatch,
	}
}

// resolveWatchConfig merges CLI flags with the loaded config into final values
// and validates the result.
func resolveWatchConfig(cmd *cli.Command, cfg WatchConfig) (WatchConfig, error) {
	resolved := cfg

	if cmd.Duration("interval") != 0 {
		resolved.Interval = cmd.Duration("interval").String()
	}

	if cmd.String("schedule") != "" {
		resolved.Schedule = cmd.String("schedule")
	}

	err := resolved.Validate()
	if err != nil {
		return WatchConfig{}, err
	}

	return resolved, nil
}

func runWatch(ctx context.Context, cmd *cli.Command) error {
	verboseVal, reverseVal := getSyncFlagsFromCmd(cmd)

	configPath := cmd.String("config")
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	applySyncFlagsToConfig(cmd, &config)

	logger := NewLogger(verboseVal)
	ctx = logger.WithContext(ctx)

	app, err := NewApp(ctx, config, reverseVal)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	resolved, err := resolveWatchConfig(cmd, config.Watch)
	if err != nil {
		return err
	}

	once := cmd.Bool("once")

	if resolved.Schedule != "" {
		return watchWithCron(ctx, app, resolved.Schedule, once)
	}

	interval, _ := resolved.GetInterval()

	return watchWithInterval(ctx, app, interval, once)
}

func watchWithInterval(ctx context.Context, runner syncRunner, interval time.Duration, once bool) error {
	if once {
		log.Printf("Running initial sync (--once flag set)...")
		err := runner.Run(ctx)
		if err != nil {
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
			runner.Refresh(ctx)
			err := runner.Run(ctx)
			if err != nil {
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

func watchWithCron(ctx context.Context, runner syncRunner, scheduleExpr string, once bool) error {
	sched, err := (&WatchConfig{Schedule: scheduleExpr}).ParseSchedule()
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	return watchWithCronSchedule(ctx, runner, sched, scheduleExpr, once)
}

func watchWithCronSchedule(ctx context.Context, runner syncRunner, sched cronSchedule, scheduleExpr string, once bool) error {
	if once {
		log.Printf("Running initial sync (--once flag set)...")
		err := runner.Run(ctx)
		if err != nil {
			return fmt.Errorf("initial sync failed: %w", err)
		}
	}

	nextTime := sched.Next(time.Now())
	log.Printf("Starting watch mode with schedule %q - next sync at %s", scheduleExpr, nextTime.Format("2006-01-02 15:04:05"))

	timer := time.NewTimer(time.Until(nextTime))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.Printf("Running scheduled sync...")
			runner.Refresh(ctx)
			err := runner.Run(ctx)
			if err != nil {
				log.Printf("Sync error: %v", err)
			}
			next := sched.Next(time.Now())
			log.Printf("Sync completed - next sync at %s", next.Format("2006-01-02 15:04:05"))
			timer.Reset(time.Until(next))
		case <-ctx.Done():
			log.Printf("Watch mode stopped")
			return nil
		}
	}
}
