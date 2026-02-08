package main

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"
)

func newLogoutCommand() *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "Remove authentication tokens",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "service",
				Aliases: []string{"s"},
				Usage:   "service to logout (anilist, myanimelist, all)",
				Value:   ServiceAll,
			},
		},
		Action: runLogout,
	}
}

func runLogout(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")
	service := cmd.String("service")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	switch service {
	case ServiceAnilist:
		return logoutAnilist(ctx, config)
	case ServiceMyAnimeList:
		return logoutMyAnimeList(ctx, config)
	case ServiceAll:
		if err := logoutMyAnimeList(ctx, config); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := logoutAnilist(ctx, config); err != nil {
			log.Printf("Warning: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown service: %s (use: anilist, myanimelist, all)", service)
	}
}

func logoutAnilist(ctx context.Context, config Config) error {
	oauth, err := NewAnilistOAuthWithoutInit(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating anilist oauth: %w", err)
	}

	if oauth.NeedInit() {
		log.Println("AniList: Not logged in")
		return nil
	}

	if err := oauth.DeleteToken(); err != nil {
		return fmt.Errorf("error removing anilist token: %w", err)
	}

	log.Println("AniList: Logged out successfully")
	return nil
}

func logoutMyAnimeList(ctx context.Context, config Config) error {
	oauth, err := NewMyAnimeListOAuthWithoutInit(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating myanimelist oauth: %w", err)
	}

	if oauth.NeedInit() {
		log.Println("MyAnimeList: Not logged in")
		return nil
	}

	if err := oauth.DeleteToken(); err != nil {
		return fmt.Errorf("error removing myanimelist token: %w", err)
	}

	log.Println("MyAnimeList: Logged out successfully")
	return nil
}
