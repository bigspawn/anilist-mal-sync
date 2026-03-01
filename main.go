package main

import (
	"log"
	"os"
)

// Default values for flags - used when sync is not called via CLI (e.g., tests)
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
	if err := RunCLI(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
