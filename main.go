package main

import (
	"log"
	"os"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
// Falls back to "dev" when built without the flag (e.g. go run .).
var version = "dev"

// Default values for flags - used when sync is not called via CLI (e.g., tests).
var (
	defaultForce   = false
	defaultDryRun  = false
	defaultManga   = false
	defaultAll     = false
	defaultVerbose = false
)

// Package-level vars for flags that don't affect domain-object behaviour.
// reverseDirection has been removed: it is now passed explicitly as a bool
// parameter to NewApp and stored in App.reverse.
var (
	forceSync = &defaultForce
	dryRun    = &defaultDryRun
	mangaSync = &defaultManga
	allSync   = &defaultAll
	verbose   = &defaultVerbose
)

func main() {
	err := RunCLI()
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
