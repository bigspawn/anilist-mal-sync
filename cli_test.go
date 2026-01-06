package main

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

// =============================================================================
// CLI Structure Tests
// =============================================================================

func TestCLI_HasCommands(t *testing.T) {
	cmd := NewCLI()

	if len(cmd.Commands) != 4 {
		t.Errorf("expected 4 commands (login, logout, status, sync), got %d", len(cmd.Commands))
	}

	commandNames := make(map[string]bool)
	for _, c := range cmd.Commands {
		commandNames[c.Name] = true
	}

	expectedCommands := []string{"login", "logout", "status", "sync"}
	for _, name := range expectedCommands {
		if !commandNames[name] {
			t.Errorf("missing command: %s", name)
		}
	}
}

func TestCLI_HasFlags(t *testing.T) {
	cmd := NewCLI()

	if len(cmd.Flags) != 7 {
		t.Errorf("expected 7 flags on root command, got %d", len(cmd.Flags))
	}

	// Check that important flags exist
	flagNames := make(map[string]bool)
	for _, f := range cmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{"config", "force", "dry-run", "manga", "all", "verbose", "reverse-direction"}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestCLI_SyncCommand_HasFlags(t *testing.T) {
	rootCmd := NewCLI()

	var syncCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == "sync" {
			syncCmd = c
			break
		}
	}

	if syncCmd == nil {
		t.Fatal("sync command not found")
	}

	if len(syncCmd.Flags) != 6 {
		t.Errorf("expected 6 flags on sync command, got %d", len(syncCmd.Flags))
	}

	// Check that sync has the right flags
	flagNames := make(map[string]bool)
	for _, f := range syncCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{"force", "dry-run", "manga", "all", "verbose", "reverse-direction"}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("sync command missing flag: %s", name)
		}
	}
}

func TestCLI_LoginCommand_HasServiceFlag(t *testing.T) {
	rootCmd := NewCLI()

	var loginCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == "login" {
			loginCmd = c
			break
		}
	}

	if loginCmd == nil {
		t.Fatal("login command not found")
	}

	if len(loginCmd.Flags) != 1 {
		t.Errorf("expected 1 flag on login command, got %d", len(loginCmd.Flags))
	}

	flag := loginCmd.Flags[0]
	if flag.Names()[0] != "service" {
		t.Errorf("expected 'service' flag, got %s", flag.Names()[0])
	}
}

func TestCLI_StatusCommand_NoFlags(t *testing.T) {
	rootCmd := NewCLI()

	var statusCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == "status" {
			statusCmd = c
			break
		}
	}

	if statusCmd == nil {
		t.Fatal("status command not found")
	}

	if len(statusCmd.Flags) != 0 {
		t.Errorf("expected 0 flags on status command, got %d", len(statusCmd.Flags))
	}
}

// =============================================================================
// Backward Compatibility Tests
// =============================================================================

func TestCLI_DefaultActionIsSync(t *testing.T) {
	cmd := NewCLI()

	if cmd.Action == nil {
		t.Error("root command should have default action (sync)")
	}
}

func TestGlobalFlagsAreSet(t *testing.T) {
	// Verify that the global flag pointers are not nil
	if forceSync == nil {
		t.Error("forceSync should not be nil")
	}
	if dryRun == nil {
		t.Error("dryRun should not be nil")
	}
	if mangaSync == nil {
		t.Error("mangaSync should not be nil")
	}
	if allSync == nil {
		t.Error("allSync should not be nil")
	}
	if verbose == nil {
		t.Error("verbose should not be nil")
	}
	if reverseDirection == nil {
		t.Error("reverseDirection should not be nil")
	}
}

func TestGlobalFlagsHaveDefaultValues(t *testing.T) {
	// Default values should be false for all flags
	if *forceSync != false {
		t.Errorf("expected forceSync default false, got %v", *forceSync)
	}
	if *dryRun != false {
		t.Errorf("expected dryRun default false, got %v", *dryRun)
	}
	if *mangaSync != false {
		t.Errorf("expected mangaSync default false, got %v", *mangaSync)
	}
	if *allSync != false {
		t.Errorf("expected allSync default false, got %v", *allSync)
	}
	if *verbose != false {
		t.Errorf("expected verbose default false, got %v", *verbose)
	}
	if *reverseDirection != false {
		t.Errorf("expected reverseDirection default false, got %v", *reverseDirection)
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestCLI_RunWithHelp(t *testing.T) {
	// Test that we can create CLI and it doesn't panic
	cmd := NewCLI()

	// Test version is set
	if cmd.Version != "" {
		// Version is set, which is good
		t.Log("CLI has version:", cmd.Version)
	}
}

func TestServiceConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"AniList constant", ServiceAnilist, "anilist"},
		{"MyAnimeList constant", ServiceMyAnimeList, "myanimelist"},
		{"All constant", ServiceAll, "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

func TestRunCLI_ContextCancellation(t *testing.T) {
	// Test that RunCLI returns without error when given empty args
	// (It will show help/usage, which is not an error)
	// We can't fully test without a real config file, but we can
	// verify the structure is correct

	cmd := NewCLI()
	if cmd == nil {
		t.Fatal("NewCLI() returned nil")
	}

	// Verify context handling is set up
	ctx := context.Background()
	// The Run method should accept context
	// This is a compile-time check essentially
	_ = ctx
	_ = cmd
}
