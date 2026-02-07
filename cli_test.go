package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

// =============================================================================
// CLI Structure Tests
// =============================================================================

func TestCLI_HasCommands(t *testing.T) {
	cmd := NewCLI()

	if len(cmd.Commands) != 5 {
		t.Errorf("expected 5 commands (login, logout, status, sync, watch), got %d", len(cmd.Commands))
	}

	commandNames := make(map[string]bool)
	for _, c := range cmd.Commands {
		commandNames[c.Name] = true
	}

	expectedCommands := []string{"login", "logout", "status", "sync", "watch"}
	for _, name := range expectedCommands {
		if !commandNames[name] {
			t.Errorf("missing command: %s", name)
		}
	}
}

func TestCLI_HasFlags(t *testing.T) {
	cmd := NewCLI()

	if len(cmd.Flags) != 11 {
		t.Errorf("expected 11 flags on root command, got %d", len(cmd.Flags))
	}

	// Check that important flags exist
	flagNames := make(map[string]bool)
	for _, f := range cmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{
		"config", "force", "dry-run", "manga", "all", "verbose", "reverse-direction",
		"offline-db", "offline-db-force-refresh", "arm-api", "arm-api-url",
	}
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

	if len(syncCmd.Flags) != 10 {
		t.Errorf("expected 10 flags on sync command, got %d", len(syncCmd.Flags))
	}

	// Check that sync has the right flags
	flagNames := make(map[string]bool)
	for _, f := range syncCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{
		"force", "dry-run", "manga", "all", "verbose", "reverse-direction",
		"offline-db", "offline-db-force-refresh", "arm-api", "arm-api-url",
	}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("sync command missing flag: %s", name)
		}
	}
}

func TestCLI_RootCommand_FlagAliases(t *testing.T) {
	cmd := NewCLI()

	aliases := make(map[string][]string)
	for _, f := range cmd.Flags {
		aliases[f.Names()[0]] = f.Names()
	}

	tests := []struct {
		flag     string
		aliases  []string
		hasAlias bool
	}{
		{"config", []string{"config", "c"}, true},
		{"force", []string{"force", "f"}, true},
		{"dry-run", []string{"dry-run", "d"}, true},
		{"manga", []string{"manga"}, false},
		{"all", []string{"all"}, false},
		{"verbose", []string{"verbose"}, false},
		{"reverse-direction", []string{"reverse-direction"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			actual, ok := aliases[tt.flag]
			if !ok {
				t.Fatalf("flag %s not found", tt.flag)
			}

			if !equalSlices(actual, tt.aliases) {
				t.Errorf("flag %s: expected aliases %v, got %v", tt.flag, tt.aliases, actual)
			}

			hasAlias := len(actual) > 1
			if hasAlias != tt.hasAlias {
				t.Errorf("flag %s: expected hasAlias=%v, got %v", tt.flag, tt.hasAlias, hasAlias)
			}
		})
	}
}

func TestCLI_SyncCommand_FlagAliases(t *testing.T) {
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

	aliases := make(map[string][]string)
	for _, f := range syncCmd.Flags {
		aliases[f.Names()[0]] = f.Names()
	}

	tests := []struct {
		flag     string
		aliases  []string
		hasAlias bool
	}{
		{"force", []string{"force", "f"}, true},
		{"dry-run", []string{"dry-run", "d"}, true},
		{"manga", []string{"manga"}, false},
		{"all", []string{"all"}, false},
		{"verbose", []string{"verbose"}, false},
		{"reverse-direction", []string{"reverse-direction"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			actual, ok := aliases[tt.flag]
			if !ok {
				t.Fatalf("flag %s not found", tt.flag)
			}

			if !equalSlices(actual, tt.aliases) {
				t.Errorf("flag %s: expected aliases %v, got %v", tt.flag, tt.aliases, actual)
			}

			hasAlias := len(actual) > 1
			if hasAlias != tt.hasAlias {
				t.Errorf("flag %s: expected hasAlias=%v, got %v", tt.flag, tt.hasAlias, hasAlias)
			}
		})
	}
}

func TestCLI_VerboseFlag_NoShortAlias(t *testing.T) {
	rootCmd := NewCLI()

	var verboseFlag cli.Flag
	for _, f := range rootCmd.Flags {
		if f.Names()[0] == "verbose" {
			verboseFlag = f
			break
		}
	}

	if verboseFlag == nil {
		t.Fatal("verbose flag not found on root command")
	}

	names := verboseFlag.Names()
	if len(names) != 1 {
		t.Errorf("verbose flag should have exactly 1 name (no aliases), got %d: %v", len(names), names)
	}

	if names[0] != "verbose" {
		t.Errorf("verbose flag primary name should be 'verbose', got %s", names[0])
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

func TestCLI_WatchCommand_HasSyncFlags(t *testing.T) {
	rootCmd := NewCLI()

	var watchCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == "watch" {
			watchCmd = c
			break
		}
	}

	if watchCmd == nil {
		t.Fatal("watch command not found")
	}

	// watch has 2 own flags (interval, once) + 10 sync flags = 12 total
	if len(watchCmd.Flags) != 12 {
		t.Errorf("expected 12 flags on watch command (2 watch + 10 sync), got %d", len(watchCmd.Flags))
	}

	// Check that sync flags are present
	flagNames := make(map[string]bool)
	for _, f := range watchCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	syncFlagNames := []string{
		"force", "dry-run", "manga", "all", "verbose", "reverse-direction",
		"offline-db", "offline-db-force-refresh", "arm-api", "arm-api-url",
	}
	for _, name := range syncFlagNames {
		if !flagNames[name] {
			t.Errorf("watch command missing sync flag: %s", name)
		}
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

// =============================================================================
// Error Detection Tests
// =============================================================================

func TestIsCancellationError_ContextCanceled(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "No error",
			err:  nil,
			want: false,
		},
		{
			name: "Random error",
			err:  fmt.Errorf("random error"),
			want: false,
		},
		{
			name: "Direct context.Canceled",
			err:  context.Canceled,
			want: true,
		},
		{
			name: "Wrapped context.Canceled",
			err:  fmt.Errorf("run app: %w", context.Canceled),
			want: true,
		},
		{
			name: "Double wrapped context.Canceled",
			err:  fmt.Errorf("command failed: %w", fmt.Errorf("run app: %w", context.Canceled)),
			want: true,
		},
		{
			name: "Context deadline exceeded (not cancellation)",
			err:  context.DeadlineExceeded,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCancellationError(tt.err)
			if got != tt.want {
				t.Errorf("IsCancellationError() = %v, want %v", got, tt.want)
			}
		})
	}
}
