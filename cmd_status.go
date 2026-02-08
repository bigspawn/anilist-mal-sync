package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v3"
)

func newStatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Check authentication status for services",
		Action: runStatus,
	}
}

func runStatus(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	log.Println("Authentication Status:")
	log.Println("======================")

	// Check AniList
	oauthAnilist, err := NewAnilistOAuthWithoutInit(ctx, config)
	if err != nil {
		log.Printf("AniList:      Error - %v\n", err)
	} else {
		printServiceStatus("AniList", oauthAnilist)
	}

	// Check MyAnimeList
	oauthMAL, err := NewMyAnimeListOAuthWithoutInit(ctx, config)
	if err != nil {
		log.Printf("MyAnimeList:  Error - %v\n", err)
	} else {
		printServiceStatus("MyAnimeList", oauthMAL)
	}

	log.Println()
	log.Printf("Token file: %s\n", config.TokenFilePath)

	return nil
}

func printServiceStatus(serviceName string, oauth *OAuth) {
	if oauth.NeedInit() {
		log.Printf("%-13s Not authenticated\n", serviceName+":")
		return
	}

	if oauth.IsTokenValid() {
		expiry := oauth.TokenExpiry()
		if expiry.IsZero() {
			log.Printf("%-13s Authenticated (no expiry)\n", serviceName+":")
		} else {
			remaining := time.Until(expiry).Round(time.Minute)
			log.Printf("%-13s Authenticated (expires in %v)\n", serviceName+":", remaining)
		}
	} else {
		log.Printf("%-13s Token expired (refresh may be attempted on use)\n", serviceName+":")
	}
}
