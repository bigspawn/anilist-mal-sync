package main

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"
)

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
	log.Println("Starting AniList authentication...")

	oauth, err := NewAnilistOAuthWithoutInit(config)
	if err != nil {
		return fmt.Errorf("error creating anilist oauth: %w", err)
	}

	if !oauth.NeedInit() {
		log.Println("AniList: Already authenticated")
		return nil
	}

	if err := oauth.InitToken(ctx, config.OAuth.Port); err != nil {
		return fmt.Errorf("anilist authentication failed: %w", err)
	}

	log.Println("AniList: Authentication successful")
	return nil
}

func loginMyAnimeList(ctx context.Context, config Config) error {
	log.Println("Starting MyAnimeList authentication...")

	oauth, err := NewMyAnimeListOAuthWithoutInit(config)
	if err != nil {
		return fmt.Errorf("error creating myanimelist oauth: %w", err)
	}

	if !oauth.NeedInit() {
		log.Println("MyAnimeList: Already authenticated")
		return nil
	}

	if err := oauth.InitToken(ctx, config.OAuth.Port); err != nil {
		return fmt.Errorf("myanimelist authentication failed: %w", err)
	}

	log.Println("MyAnimeList: Authentication successful")
	return nil
}
