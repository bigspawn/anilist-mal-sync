package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func newSyncCommand() *cli.Command {
	return &cli.Command{
		Name:   "sync",
		Usage:  "Synchronize anime/manga lists between services",
		Flags:  syncFlags,
		Action: runSync,
	}
}

func runSync(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")

	// Set package-level vars for compatibility with existing code
	verboseVal := setSyncFlagsFromCmd(cmd)

	// Initialize logger and add to context
	logger := NewLogger(verboseVal)
	ctx = logger.WithContext(ctx)

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
