package main

import (
	"log"
	"os"
)

// Package-level vars for backward compatibility with app.go/updater.go
// These are set by sync command
var (
	forceSync        *bool
	dryRun           *bool
	mangaSync        *bool
	allSync          *bool
	verbose          *bool
	reverseDirection *bool
)

// Default values for flags - used when sync is not called via CLI (e.g., tests)
var (
	defaultForce        = false
	defaultDryRun       = false
	defaultManga        = false
	defaultAll          = false
	defaultVerbose      = false
	defaultReverse      = false
)

func init() {
	forceSync = &defaultForce
	dryRun = &defaultDryRun
	mangaSync = &defaultManga
	allSync = &defaultAll
	verbose = &defaultVerbose
	reverseDirection = &defaultReverse
}

func main() {
	if err := RunCLI(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
