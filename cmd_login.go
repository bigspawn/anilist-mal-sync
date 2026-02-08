package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

// colorPrint prints colored text to stdout, ignoring write errors
func colorPrint(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

func newLoginCommand() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate with AniList and/or MyAnimeList",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "service",
				Aliases: []string{"s"},
				Usage:   "service to authenticate (anilist, myanimelist, all)",
				Value:   ServiceAll,
			},
		},
		Action: runLogin,
	}
}

func runLogin(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")
	service := cmd.String("service")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	switch service {
	case ServiceAnilist:
		return loginAnilist(ctx, config)
	case ServiceMyAnimeList:
		return loginMyAnimeList(ctx, config)
	case ServiceAll:
		if err := loginMyAnimeList(ctx, config); err != nil {
			return err
		}
		return loginAnilist(ctx, config)
	default:
		return fmt.Errorf("unknown service: %s (use: anilist, myanimelist, all)", service)
	}
}

func loginAnilist(ctx context.Context, config Config) error {
	colorPrint("\n%s=== AniList Authentication ===%s\n", colorBold+colorCyan, colorReset)
	log.Println("Starting AniList authentication...")

	oauth, err := NewAnilistOAuthWithoutInit(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating anilist oauth: %w", err)
	}

	if !oauth.NeedInit() {
		colorPrint("%s✓ AniList: Already authenticated%s\n\n", colorGreen, colorReset)
		printNextSteps()
		return nil
	}

	if err := oauth.InitToken(ctx, config.OAuth.Port); err != nil {
		return fmt.Errorf("anilist authentication failed: %w", err)
	}

	colorPrint("%s✓ AniList: Authentication successful%s\n\n", colorGreen, colorReset)
	printNextSteps()
	return nil
}

func loginMyAnimeList(ctx context.Context, config Config) error {
	colorPrint("\n%s=== MyAnimeList Authentication ===%s\n", colorBold+colorCyan, colorReset)
	log.Println("Starting MyAnimeList authentication...")

	oauth, err := NewMyAnimeListOAuthWithoutInit(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating myanimelist oauth: %w", err)
	}

	if !oauth.NeedInit() {
		colorPrint("%s✓ MyAnimeList: Already authenticated%s\n\n", colorGreen, colorReset)
		printNextSteps()
		return nil
	}

	if err := oauth.InitToken(ctx, config.OAuth.Port); err != nil {
		return fmt.Errorf("myanimelist authentication failed: %w", err)
	}

	colorPrint("%s✓ MyAnimeList: Authentication successful%s\n\n", colorGreen, colorReset)
	printNextSteps()
	return nil
}

func printNextSteps() {
	colorPrint("%sNext steps:%s\n", colorBold+colorYellow, colorReset)
	colorPrint("  Run %sstatus%s to check authentication status\n", colorCyan, colorReset)
	colorPrint("  Run %sanilist-mal-sync sync%s to start synchronization\n", colorCyan, colorReset)
	colorPrint("  Or run %slogin --service <service>%s for additional authentication\n\n", colorCyan, colorReset)
}
