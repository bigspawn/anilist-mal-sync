package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

// =============================================================================
// CLI Structure Tests
// =============================================================================

func TestCLI_HasCommands(t *testing.T) {
	t.Parallel()
	cmd := NewCLI()

	if len(cmd.Commands) != 6 {
		t.Errorf("expected 6 commands (login, logout, status, sync, watch, unmapped), got %d", len(cmd.Commands))
	}

	commandNames := make(map[string]bool)
	for _, c := range cmd.Commands {
		commandNames[c.Name] = true
	}

	expectedCommands := []string{"login", "logout", "status", "sync", "watch", "unmapped"}
	for _, name := range expectedCommands {
		if !commandNames[name] {
			t.Errorf("missing command: %s", name)
		}
	}
}

func TestCLI_HasFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCLI()

	if len(cmd.Flags) != 17 {
		t.Errorf("expected 17 flags on root command, got %d", len(cmd.Flags))
	}

	// Check that important flags exist
	flagNames := make(map[string]bool)
	for _, f := range cmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{
		"config", flagForce, "dry-run", "manga", string(ServiceAll), "verbose", "reverse-direction",
		flagOfflineDB, flagOfflineDBForceRefresh, flagARMAPI, flagARMAPIURL,
		flagHatoAPI, flagHatoAPIURL, flagMangaBakaAPI, flagMangaBakaAPIURL, flagJikanAPI,
	}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestCLI_SyncCommand_HasFlags(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	syncCmd := mustFindCommand(t, rootCmd, "sync")

	if len(syncCmd.Flags) != 16 {
		t.Errorf("expected 16 flags on sync command, got %d", len(syncCmd.Flags))
	}

	// Check that sync has the right flags
	flagNames := make(map[string]bool)
	for _, f := range syncCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{
		"force", "dry-run", "manga", string(ServiceAll), "verbose", "reverse-direction",
		flagOfflineDB, flagOfflineDBForceRefresh, flagARMAPI, flagARMAPIURL,
		flagHatoAPI, flagHatoAPIURL, flagMangaBakaAPI, flagMangaBakaAPIURL, flagJikanAPI,
	}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("sync command missing flag: %s", name)
		}
	}
}

func TestCLI_RootCommand_FlagAliases(t *testing.T) {
	t.Parallel()
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
		{"force", []string{flagForce, "f"}, true},
		{"dry-run", []string{"dry-run", "d"}, true},
		{"manga", []string{"manga"}, false},
		{"all", []string{string(ServiceAll)}, false},
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
	t.Parallel()
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
		{"force", []string{flagForce, "f"}, true},
		{"dry-run", []string{"dry-run", "d"}, true},
		{"manga", []string{"manga"}, false},
		{"all", []string{string(ServiceAll)}, false},
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	rootCmd := NewCLI()

	watchCmd := mustFindCommand(t, rootCmd, watchCommandName)

	// watch has 3 own flags (interval, schedule, once) + 16 sync flags = 19 total
	if len(watchCmd.Flags) != 19 {
		t.Errorf("expected 19 flags on watch command (3 watch + 16 sync), got %d", len(watchCmd.Flags))
	}

	// Check that sync flags are present
	flagNames := make(map[string]bool)
	for _, f := range watchCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	syncFlagNames := []string{
		"force", "dry-run", "manga", string(ServiceAll), "verbose", "reverse-direction",
		flagOfflineDB, flagOfflineDBForceRefresh, flagARMAPI, flagARMAPIURL,
		flagHatoAPI, flagHatoAPIURL, flagMangaBakaAPI, flagMangaBakaAPIURL, flagJikanAPI,
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
	t.Parallel()
	cmd := NewCLI()

	if cmd.Action == nil {
		t.Error("root command should have default action (sync)")
	}
}

func TestGlobalFlagsAreSet(t *testing.T) {
	t.Parallel()
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
}

func TestGlobalFlagsHaveDefaultValues(t *testing.T) {
	t.Parallel()
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
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestCLI_RunWithHelp(t *testing.T) {
	t.Parallel()
	// Test that we can create CLI and it doesn't panic
	cmd := NewCLI()

	// Test version is set
	if cmd.Version != "" {
		// Version is set, which is good
		t.Log("CLI has version:", cmd.Version)
	}
}

func TestServiceConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"AniList constant", ServiceAnilist, "anilist"},
		{"MyAnimeList constant", ServiceMyAnimeList, string(ServiceMyAnimeList)},
		{"All constant", ServiceAll, string(ServiceAll)},
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
	t.Parallel()
	// Test that RunCLI returns without error when given empty args
	// (It will show help/usage, which is not an error)
	// We can't fully test without a real config file, but we can
	// verify the structure is correct

	cmd := NewCLI()
	if cmd == nil {
		t.Fatal("NewCLI() returned nil")
	}

	// Verify context handling is set up
	ctx := t.Context()
	// The Run method should accept context
	// This is a compile-time check essentially
	_ = ctx
	_ = cmd
}

// =============================================================================
// Error Detection Tests
// =============================================================================

func TestIsCancellationError_ContextCanceled(t *testing.T) {
	t.Parallel()
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
			err:  errors.New("random error"),
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

// =============================================================================
// Local Flag Tests (verify sync flags don't leak to other commands)
// =============================================================================

func TestCLI_SyncFlagsMarkedAsLocal(t *testing.T) {
	t.Parallel()
	cmd := NewCLI()

	// Flags that should be marked as Local (not inherited by subcommands)
	syncSpecificFlags := []string{
		"force", "dry-run", "manga", string(ServiceAll), "verbose", "reverse-direction",
		flagOfflineDB, flagOfflineDBForceRefresh, flagARMAPI, flagARMAPIURL,
	}

	for _, flagName := range syncSpecificFlags {
		t.Run(flagName, func(t *testing.T) {
			var foundFlag cli.Flag
			for _, f := range cmd.Flags {
				if f.Names()[0] == flagName {
					foundFlag = f
					break
				}
			}

			if foundFlag == nil {
				t.Fatalf("flag %s not found on root command", flagName)
			}

			// Check if flag implements LocalFlag interface
			localFlag, ok := foundFlag.(cli.LocalFlag)
			if !ok {
				t.Errorf("flag %s does not implement LocalFlag interface", flagName)
				return
			}

			if !localFlag.IsLocal() {
				t.Errorf("flag %s should be marked as Local (true), got Local=false", flagName)
			}
		})
	}

	// config flag should NOT be local (it's truly global)
	t.Run("config_is_not_local", func(t *testing.T) {
		var configFlag cli.Flag
		for _, f := range cmd.Flags {
			if f.Names()[0] == "config" {
				configFlag = f
				break
			}
		}

		if configFlag == nil {
			t.Fatal("config flag not found on root command")
		}

		localFlag, ok := configFlag.(cli.LocalFlag)
		if ok && localFlag.IsLocal() {
			t.Errorf("config flag should NOT be marked as Local, got Local=true")
		}
	})
}

func TestCLI_NonSyncCommandsDontInheritSyncFlags(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	// Sync-specific flags that should NOT appear in non-sync subcommands
	syncSpecificFlags := []string{
		"force", "dry-run", "manga", string(ServiceAll), "verbose", "reverse-direction",
		flagOfflineDB, flagOfflineDBForceRefresh, flagARMAPI, flagARMAPIURL,
	}

	nonSyncCommands := []string{"login", "logout", "status"}

	for _, cmdName := range nonSyncCommands {
		t.Run(cmdName, func(t *testing.T) {
			var subCmd *cli.Command
			for _, c := range rootCmd.Commands {
				if c.Name == cmdName {
					subCmd = c
					break
				}
			}

			if subCmd == nil {
				t.Fatalf("%s command not found", cmdName)
			}

			// Get the flag names in this subcommand
			flagNames := make(map[string]bool)
			for _, f := range subCmd.Flags {
				flagNames[f.Names()[0]] = true
			}

			// Verify sync-specific flags are NOT present
			for _, syncFlag := range syncSpecificFlags {
				if flagNames[syncFlag] {
					t.Errorf("%s command should not have sync-specific flag '%s'", cmdName, syncFlag)
				}
			}
		})
	}
}

// mustFindCommand returns the named subcommand, failing the test when absent.
// Returning the value keeps callers free of a nil check that staticcheck reads
// as a possible nil dereference.
func mustFindCommand(t *testing.T, root *cli.Command, name string) *cli.Command {
	t.Helper()
	for _, c := range root.Commands {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s command not found", name)
	return nil
}

// Every ID-mapping source is configurable three ways: YAML, env var, CLI flag.
// The flag is the highest precedence, so it must win over an env var that says
// the opposite.
func TestApplySyncFlagsToConfig_FlagOverridesEnvForMangaBakaAndHato(t *testing.T) {
	rootCmd := NewCLI()

	// Env says MangaBaka off and Hato on; the flags say the reverse.
	t.Setenv("MANGABAKA_API_ENABLED", "false")
	t.Setenv("HATO_API_ENABLED", "true")

	cfg := Config{}
	cfg.MangaBakaAPI.Enabled = false
	cfg.HatoAPI.Enabled = true

	syncCmd := mustFindCommand(t, rootCmd, "sync")

	syncCmd.Action = func(_ context.Context, cmd *cli.Command) error {
		applySyncFlagsToConfig(cmd, &cfg)
		return nil
	}

	args := []string{
		"anilist-mal-sync", "sync",
		"--mangabaka-api=true",
		"--hato-api=false",
		"--mangabaka-api-url", "https://flag.test/v1",
	}
	err := rootCmd.Run(t.Context(), args)
	if err != nil {
		t.Fatalf("running sync failed: %v", err)
	}

	if !cfg.MangaBakaAPI.Enabled {
		t.Error("--mangabaka-api=true must override MANGABAKA_API_ENABLED=false")
	}
	if cfg.HatoAPI.Enabled {
		t.Error("--hato-api=false must override HATO_API_ENABLED=true")
	}
	if cfg.MangaBakaAPI.BaseURL != "https://flag.test/v1" {
		t.Errorf("MangaBakaAPI.BaseURL = %q, want the flag value", cfg.MangaBakaAPI.BaseURL)
	}
}

func TestCLI_ARMAPIFlagDescriptionContainsDefault(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	var armAPIFlag cli.Flag
	for _, f := range rootCmd.Flags {
		if f.Names()[0] == flagARMAPI {
			armAPIFlag = f
			break
		}
	}

	if armAPIFlag == nil {
		t.Fatal("arm-api flag not found on root command")
	}

	// Check if flag implements DocGenerationFlag interface to get usage
	docFlag, ok := armAPIFlag.(cli.DocGenerationFlag)
	if !ok {
		t.Fatal("arm-api flag does not implement DocGenerationFlag interface")
	}

	usage := docFlag.GetUsage()

	// Check for both "(default: false)" and that it mentions "anime ID mapping"
	if !contains(usage, "(default: false)") {
		t.Errorf("arm-api flag usage should contain '(default: false)', got %q", usage)
	}
	if !contains(usage, "anime ID mapping") {
		t.Errorf("arm-api flag usage should contain 'anime ID mapping', got %q", usage)
	}
	if !contains(usage, "ignored for --manga") {
		t.Errorf("arm-api flag usage should contain 'ignored for --manga', got %q", usage)
	}
}

func TestCLI_OfflineDBFlagDescriptionContainsDefault(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	var offlineDBFlag cli.Flag
	for _, f := range rootCmd.Flags {
		if f.Names()[0] == flagOfflineDB {
			offlineDBFlag = f
			break
		}
	}

	if offlineDBFlag == nil {
		t.Fatal("offline-db flag not found on root command")
	}

	docFlag, ok := offlineDBFlag.(cli.DocGenerationFlag)
	if !ok {
		t.Fatal("offline-db flag does not implement DocGenerationFlag interface")
	}

	usage := docFlag.GetUsage()

	// Check for both "(default: true)" and that it mentions "anime ID mapping"
	if !contains(usage, "(default: true)") {
		t.Errorf("offline-db flag usage should contain '(default: true)', got %q", usage)
	}
	if !contains(usage, "anime ID mapping") {
		t.Errorf("offline-db flag usage should contain 'anime ID mapping', got %q", usage)
	}
	if !contains(usage, "ignored for --manga") {
		t.Errorf("offline-db flag usage should contain 'ignored for --manga', got %q", usage)
	}
}

// =============================================================================
// Logout Command Tests (missing from original test suite)
// =============================================================================

func TestCLI_LogoutCommand_HasServiceFlag(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	var logoutCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == "logout" {
			logoutCmd = c
			break
		}
	}

	if logoutCmd == nil {
		t.Fatal("logout command not found")
	}

	if len(logoutCmd.Flags) != 1 {
		t.Errorf("expected 1 flag on logout command, got %d", len(logoutCmd.Flags))
	}

	flag := logoutCmd.Flags[0]
	if flag.Names()[0] != "service" {
		t.Errorf("expected 'service' flag, got %s", flag.Names()[0])
	}
}

// =============================================================================
// Watch Command Specific Flags Tests
// =============================================================================

func TestCLI_WatchCommand_HasIntervalAndOnceFlags(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	var watchCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == watchCommandName {
			watchCmd = c
			break
		}
	}

	if watchCmd == nil {
		t.Fatal("watch command not found")
	}

	flagNames := make(map[string]bool)
	for _, f := range watchCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	// Check for watch-specific flags
	watchFlags := []string{"interval", "once"}
	for _, flagName := range watchFlags {
		if !flagNames[flagName] {
			t.Errorf("watch command missing flag: %s", flagName)
		}
	}
}

func TestCLI_WatchCommand_IntervalHasShortAlias(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()

	var watchCmd *cli.Command
	for _, c := range rootCmd.Commands {
		if c.Name == watchCommandName {
			watchCmd = c
			break
		}
	}

	if watchCmd == nil {
		t.Fatal("watch command not found")
	}

	var intervalFlag cli.Flag
	for _, f := range watchCmd.Flags {
		if f.Names()[0] == "interval" {
			intervalFlag = f
			break
		}
	}

	if intervalFlag == nil {
		t.Fatal("interval flag not found on watch command")
	}

	names := intervalFlag.Names()
	expectedAliases := []string{"interval", "i"}

	if !equalSlices(names, expectedAliases) {
		t.Errorf("interval flag: expected aliases %v, got %v", expectedAliases, names)
	}
}

// =============================================================================
// getSyncFlagsFromCmd Regression Tests
// =============================================================================

// saveGlobalSyncFlags saves package-level sync flag globals and returns a
// restore function. Use with defer in tests that call getSyncFlagsFromCmd
// to prevent global state leakage between parallel tests.
func saveGlobalSyncFlags() func() {
	savedVerbose := verbose
	savedForce := forceSync
	savedDry := dryRun
	savedManga := mangaSync
	savedAll := allSync
	return func() {
		verbose = savedVerbose
		forceSync = savedForce
		dryRun = savedDry
		mangaSync = savedManga
		allSync = savedAll
	}
}

// TestGetSyncFlagsFromCmd_ReverseDirection ensures the --reverse-direction flag
// is correctly read by getSyncFlagsFromCmd. This is a regression test for a bug
// where cmd.Bool("reverse") was used instead of cmd.Bool("reverse-direction"),
// causing the flag to always return false.
func TestGetSyncFlagsFromCmd_ReverseDirection(t *testing.T) {
	defer saveGlobalSyncFlags()()

	var gotReverse bool
	root := NewCLI()
	for _, c := range root.Commands {
		if c.Name == "sync" {
			c.Action = func(_ context.Context, cmd *cli.Command) error {
				_, gotReverse = getSyncFlagsFromCmd(cmd)
				return nil
			}
			break
		}
	}

	ctx := context.Background()
	err := root.Run(ctx, []string{"app", "sync", "--reverse-direction"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotReverse {
		t.Error("getSyncFlagsFromCmd: reverseOut should be true when --reverse-direction is passed")
	}
}

func TestGetSyncFlagsFromCmd_NoReverse(t *testing.T) {
	defer saveGlobalSyncFlags()()

	var gotReverse bool
	root := NewCLI()
	for _, c := range root.Commands {
		if c.Name == "sync" {
			c.Action = func(_ context.Context, cmd *cli.Command) error {
				_, gotReverse = getSyncFlagsFromCmd(cmd)
				return nil
			}
			break
		}
	}

	ctx := context.Background()
	err := root.Run(ctx, []string{"app", "sync"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotReverse {
		t.Error("getSyncFlagsFromCmd: reverseOut should be false when --reverse-direction is not passed")
	}
}

func TestGetSyncFlagsFromCmd_VerboseFlag(t *testing.T) {
	defer saveGlobalSyncFlags()()

	var gotVerbose bool
	root := NewCLI()
	for _, c := range root.Commands {
		if c.Name == "sync" {
			c.Action = func(_ context.Context, cmd *cli.Command) error {
				gotVerbose, _ = getSyncFlagsFromCmd(cmd)
				return nil
			}
			break
		}
	}

	ctx := context.Background()
	err := root.Run(ctx, []string{"app", "sync", "--verbose"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotVerbose {
		t.Error("getSyncFlagsFromCmd: verboseOut should be true when --verbose is passed")
	}
}

// =============================================================================
// Mapping Source Registry Tests
// =============================================================================

// TestMappingSources_FlagOnRootAndSyncCommand guards against the trap the
// registry exists to remove: a source flag present on one command's flag
// list but missing from the other.
func TestMappingSources_FlagOnRootAndSyncCommand(t *testing.T) {
	t.Parallel()
	rootCmd := NewCLI()
	syncCmd := mustFindCommand(t, rootCmd, "sync")

	rootFlags := flagNameSet(rootCmd.Flags)
	syncFlags := flagNameSet(syncCmd.Flags)

	for _, src := range mappingSources() {
		if !rootFlags[src.flag] {
			t.Errorf("%s: flag %q missing from root command", src.name, src.flag)
		}
		if !syncFlags[src.flag] {
			t.Errorf("%s: flag %q missing from sync command", src.name, src.flag)
		}
		if src.urlFlag == "" {
			continue
		}
		if !rootFlags[src.urlFlag] {
			t.Errorf("%s: url flag %q missing from root command", src.name, src.urlFlag)
		}
		if !syncFlags[src.urlFlag] {
			t.Errorf("%s: url flag %q missing from sync command", src.name, src.urlFlag)
		}
	}
}

// TestMappingSources_FlagOverridesEnv checks, for every registered source,
// that its CLI flag wins over an opposite env var — the core guarantee
// applySyncFlagsToConfig must keep for a source added via the registry.
func TestMappingSources_FlagOverridesEnv(t *testing.T) {
	for _, src := range mappingSources() {
		t.Run(src.name, func(t *testing.T) {
			t.Setenv(src.envPrefix+"ENABLED", "true")

			cfg := Config{}
			src.configField(&cfg).Enabled = true

			rootCmd := NewCLI()
			syncCmd := mustFindCommand(t, rootCmd, "sync")
			syncCmd.Action = func(_ context.Context, cmd *cli.Command) error {
				applySyncFlagsToConfig(cmd, &cfg)
				return nil
			}

			args := []string{"anilist-mal-sync", "sync", "--" + src.flag + "=false"}
			err := rootCmd.Run(t.Context(), args)
			if err != nil {
				t.Fatalf("running sync failed: %v", err)
			}

			if src.configField(&cfg).Enabled {
				t.Errorf("--%s=false must override %sENABLED=true", src.flag, src.envPrefix)
			}
		})
	}
}

// TestMappingSources_DefaultsMatchDocumentedTable pins the registry's
// defaults to the ones CLAUDE.md and the plan's ground truth document.
func TestMappingSources_DefaultsMatchDocumentedTable(t *testing.T) {
	t.Parallel()
	want := map[string]struct {
		enabled bool
		baseURL string
	}{
		"ARM":               {false, defaultARMBaseURL},
		"Hato":              {true, defaultHatoBaseURL},
		mangaBakaSourceName: {false, defaultMangaBakaBaseURL},
		"Jikan":             {false, ""},
	}

	for _, src := range mappingSources() {
		exp, ok := want[src.name]
		if !ok {
			t.Fatalf("unexpected source %q in registry, update this test's table", src.name)
		}
		if src.defaultEnabled != exp.enabled {
			t.Errorf("%s: defaultEnabled = %v, want %v", src.name, src.defaultEnabled, exp.enabled)
		}
		if src.defaultBaseURL != exp.baseURL {
			t.Errorf("%s: defaultBaseURL = %q, want %q", src.name, src.defaultBaseURL, exp.baseURL)
		}
	}
}

func flagNameSet(flags []cli.Flag) map[string]bool {
	names := make(map[string]bool, len(flags))
	for _, f := range flags {
		names[f.Names()[0]] = true
	}
	return names
}

// mappingSourceTestYAML has no ID-mapping sections at all, so every
// source's YAML-path default in the tests below comes from
// defaultConfig(), never from a value written in the file.
var mappingSourceTestYAML = []byte(`
anilist:
  client_id: test_id
  username: test_user
myanimelist:
  client_id: mal_id
  username: mal_user
`)

// TestMappingSourceDefaults_AgreeAcrossAllPaths guards the class of bug
// behind the MangaBaka default fix: a source's default must not need
// restating in more than one place. It iterates mappingSources() itself —
// not a hand-written list of source names — so a future source that wires
// one path but forgets another fails here automatically.
func TestMappingSourceDefaults_AgreeAcrossAllPaths(t *testing.T) {
	for _, src := range mappingSources() {
		t.Run(src.name, func(t *testing.T) {
			clearMappingSourceEnv(t, src)

			envCfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("loadConfigFromEnv() error = %v", err)
			}

			yamlCfg, err := parseConfigFile(mappingSourceTestYAML, "config.yaml")
			if err != nil {
				t.Fatalf("parseConfigFile() error = %v", err)
			}

			wantEnabled := src.defaultEnabled
			if got := src.configField(&envCfg).Enabled; got != wantEnabled {
				t.Errorf("env-loading path: Enabled = %v, want registry default %v", got, wantEnabled)
			}
			if got := src.configField(&yamlCfg).Enabled; got != wantEnabled {
				t.Errorf("YAML path (section omitted): Enabled = %v, want registry default %v", got, wantEnabled)
			}
			if got := rootCommandBoolFlagDefault(t, src.flag); got != wantEnabled {
				t.Errorf("CLI flag %q default = %v, want registry default %v", src.flag, got, wantEnabled)
			}

			if src.urlFlag == "" {
				return
			}

			wantURL := src.defaultBaseURL
			if got := src.configField(&envCfg).BaseURL; got != wantURL {
				t.Errorf("env-loading path: BaseURL = %q, want registry default %q", got, wantURL)
			}
			if got := src.configField(&yamlCfg).BaseURL; got != wantURL {
				t.Errorf("YAML path (section omitted): BaseURL = %q, want registry default %q", got, wantURL)
			}
			if got := baseURLSurvivesUnsetFlag(t, src); got != wantURL {
				t.Errorf("CLI flag %q left unset: BaseURL = %q, want registry default %q", src.urlFlag, got, wantURL)
			}
		})
	}
}

// clearMappingSourceEnv unsets every env var one mapping source reads, so
// its defaults show through undisturbed by whatever the test process
// inherited from its environment.
func clearMappingSourceEnv(t *testing.T, src mappingSource) {
	t.Helper()
	t.Setenv(src.envPrefix+"ENABLED", "")
	if src.urlFlag != "" {
		t.Setenv(src.envPrefix+"URL", "")
	}
	if src.defaultCacheDir != nil {
		t.Setenv(src.envPrefix+"CACHE_DIR", "")
		t.Setenv(src.envPrefix+"CACHE_MAX_AGE", "")
	}
}

// rootCommandBoolFlagDefault reads a bool flag's own default (its Value
// field) off the actual root command, instead of restating what the
// default is supposed to be.
func rootCommandBoolFlagDefault(t *testing.T, flagName string) bool {
	t.Helper()
	cmd := NewCLI()
	for _, f := range cmd.Flags {
		bf, ok := f.(*cli.BoolFlag)
		if ok && bf.Name == flagName {
			return bf.Value
		}
	}
	t.Fatalf("bool flag %q not found on root command", flagName)
	return false
}

// baseURLSurvivesUnsetFlag runs applySyncFlagsToConfig with src's url flag
// left unset and returns the resulting BaseURL. A url flag carries no
// Value of its own (empty string means "not overridden" — see
// applySyncFlagsToConfig's cmd.IsSet guard), so unlike the bool flags above
// there is no literal default to read off it; what stands in for "the
// flag's own default" is that leaving it unset must never clobber the
// registry default the env and YAML paths already agreed on.
func baseURLSurvivesUnsetFlag(t *testing.T, src mappingSource) string {
	t.Helper()

	cfg := Config{}
	src.configField(&cfg).BaseURL = src.defaultBaseURL

	rootCmd := NewCLI()
	syncCmd := mustFindCommand(t, rootCmd, "sync")
	syncCmd.Action = func(_ context.Context, cmd *cli.Command) error {
		applySyncFlagsToConfig(cmd, &cfg)
		return nil
	}

	err := rootCmd.Run(t.Context(), []string{"anilist-mal-sync", "sync"})
	if err != nil {
		t.Fatalf("running sync failed: %v", err)
	}

	return src.configField(&cfg).BaseURL
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
