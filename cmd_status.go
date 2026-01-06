package main

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"
)

func newStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Check authentication status for services",
		Action: runStatus,
	}
}

func runStatus(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")

	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	fmt.Println("Authentication Status:")
	fmt.Println("======================")

	// Check AniList
	oauthAnilist, err := NewAnilistOAuthWithoutInit(config)
	if err != nil {
		fmt.Printf("AniList:      Error - %v\n", err)
	} else {
		printServiceStatus("AniList", oauthAnilist)
	}

	// Check MyAnimeList
	oauthMAL, err := NewMyAnimeListOAuthWithoutInit(config)
	if err != nil {
		fmt.Printf("MyAnimeList:  Error - %v\n", err)
	} else {
		printServiceStatus("MyAnimeList", oauthMAL)
	}

	fmt.Println()
	fmt.Printf("Token file: %s\n", config.TokenFilePath)

	return nil
}

func printServiceStatus(serviceName string, oauth *OAuth) {
	if oauth.NeedInit() {
		fmt.Printf("%-13s Not authenticated\n", serviceName+":")
		return
	}

	if oauth.IsTokenValid() {
		expiry := oauth.TokenExpiry()
		if expiry.IsZero() {
			fmt.Printf("%-13s Authenticated (no expiry)\n", serviceName+":")
		} else {
			remaining := time.Until(expiry).Round(time.Minute)
			fmt.Printf("%-13s Authenticated (expires in %v)\n", serviceName+":", remaining)
		}
	} else {
		fmt.Printf("%-13s Token expired (refresh may be attempted on use)\n", serviceName+":")
	}
}
