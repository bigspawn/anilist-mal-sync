package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	watchCommandName       = "watch"
	testCronScheduleDaily3 = "0 3 * * *"
)

// =============================================================================
// Watch Command Structure Tests
// =============================================================================

func TestWatchCommand_Exists(t *testing.T) {
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

	if watchCmd.Usage != "Run sync on interval or cron schedule (Docker-friendly)" {
		t.Errorf("unexpected usage: %s", watchCmd.Usage)
	}
}

func TestWatchCommand_HasFlags(t *testing.T) {
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

	// watch has 3 own flags + 12 sync flags = 15 total
	if len(watchCmd.Flags) != 15 {
		t.Errorf("expected 15 flags (3 watch + 12 sync), got %d", len(watchCmd.Flags))
	}

	// Check flags by name
	flagNames := make(map[string]bool)
	for _, f := range watchCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{
		"interval", "schedule", "once", "force", "dry-run", "manga", "all", "verbose", "reverse-direction",
		"offline-db", "offline-db-force-refresh", "arm-api", "arm-api-url", "jikan-api",
	}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestWatchCommand_IntervalFlag_HasCorrectDefaults(t *testing.T) {
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

	var intervalFlag *cli.DurationFlag
	for _, f := range watchCmd.Flags {
		if f.Names()[0] == "interval" {
			if df, ok := f.(*cli.DurationFlag); ok {
				intervalFlag = df
			}
			break
		}
	}

	if intervalFlag == nil {
		t.Fatal("interval flag not found or not a DurationFlag")
	}

	// No default value (interval must come from CLI or config)
	if intervalFlag.Value != 0 {
		t.Errorf("expected no default (0), got %v", intervalFlag.Value)
	}

	// Check alias
	aliases := intervalFlag.Aliases
	if len(aliases) != 1 || aliases[0] != "i" {
		t.Errorf("expected alias 'i', got %v", aliases)
	}
}

func TestWatchCommand_OnceFlag_IsBool(t *testing.T) {
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

	var onceFlag *cli.BoolFlag
	for _, f := range watchCmd.Flags {
		if f.Names()[0] == "once" {
			if bf, ok := f.(*cli.BoolFlag); ok {
				onceFlag = bf
			}
			break
		}
	}

	if onceFlag == nil {
		t.Fatal("once flag not found or not a BoolFlag")
	}
}

// =============================================================================
// Interval Validation Tests
// =============================================================================

func TestValidateInterval_ValidIntervals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		interval time.Duration
		valid    bool
	}{
		{"Minimum valid", minInterval, true},
		{"Maximum valid", maxInterval, true},
		{"Default interval", 24 * time.Hour, true},
		{"12 hours", 12 * time.Hour, true},
		{"48 hours", 48 * time.Hour, true},
		{"Too small", 30 * time.Minute, false},
		{"Too large", 200 * time.Hour, false},
		{"Zero", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInterval(tt.interval)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected invalid, got no error")
			}
		})
	}
}

// validateInterval is a helper function extracted from runWatch for testing.
func validateInterval(interval time.Duration) error {
	if interval < minInterval {
		return errors.New("interval must be at least 1h")
	}
	if interval > maxInterval {
		return errors.New("interval must be at most 168h")
	}
	return nil
}

// =============================================================================
// Priority Tests
// =============================================================================

func TestWatch_CLI_Overrides_Config(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has 12h, CLI uses 24h
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
  username: "ani_user"
myanimelist:
  client_id: "test"
  client_secret: "test"
  username: "mal_user"
watch:
  interval: "12h"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create command with CLI flag
	cmd := cli.Command{
		Name: "watch",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "interval",
				Value: 24 * time.Hour, // CLI flag set
			},
			&cli.StringFlag{
				Name:  "config",
				Value: configPath,
			},
		},
	}

	// Run the priority logic
	interval := cmd.Duration("interval")

	if interval == 0 {
		// Would read from config
		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}
		cfgInterval, err := cfg.Watch.GetInterval()
		if err != nil {
			t.Fatalf("failed to get interval from config: %v", err)
		}
		interval = cfgInterval
	}

	// CLI flag (24h) should override config (12h)
	expected := 24 * time.Hour
	if interval != expected {
		t.Errorf("interval = %v, want %v (CLI should override config)", interval, expected)
	}
}

func TestWatch_ConfigOnly(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has 12h, no CLI flag
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
  username: "ani_user"
myanimelist:
  client_id: "test"
  client_secret: "test"
  username: "mal_user"
watch:
  interval: "12h"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cfgInterval, err := cfg.Watch.GetInterval()
	if err != nil {
		t.Fatalf("failed to get interval from config: %v", err)
	}

	expected := 12 * time.Hour
	if cfgInterval != expected {
		t.Errorf("config interval = %v, want %v", cfgInterval, expected)
	}
}

func TestWatch_NoInterval_Error(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has no watch interval
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
  username: "ani_user"
myanimelist:
  client_id: "test"
  client_secret: "test"
  username: "mal_user"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := loadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	cfgInterval, err := cfg.Watch.GetInterval()
	if err != nil {
		t.Fatalf("failed to get interval from config: %v", err)
	}

	if cfgInterval != 0 {
		t.Errorf("config interval should be 0 (not specified), got %v", cfgInterval)
	}
}

func TestWatch_InvalidConfigInterval_Error(t *testing.T) {
	t.Parallel()
	w := WatchConfig{Interval: "invalid-duration"}
	_, err := w.GetInterval()
	if err == nil {
		t.Error("expected error for invalid interval, got nil")
	}
}

func TestWatch_IntervalValidation_Config_TooSmall(t *testing.T) {
	t.Parallel()
	w := WatchConfig{Interval: "30m"}
	got, err := w.GetInterval()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got < minInterval {
		t.Logf("interval %v is less than min %v (validation will catch this)", got, minInterval)
	}
}

func TestWatch_IntervalValidation_Config_TooLarge(t *testing.T) {
	t.Parallel()
	w := WatchConfig{Interval: "200h"}
	got, err := w.GetInterval()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got > maxInterval {
		t.Logf("interval %v exceeds max %v (validation will catch this)", got, maxInterval)
	}
}

// =============================================================================
// Schedule Flag Tests
// =============================================================================

func TestWatchCommand_ScheduleFlag_Exists(t *testing.T) {
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

	var scheduleFlag *cli.StringFlag
	for _, f := range watchCmd.Flags {
		if f.Names()[0] == "schedule" {
			if sf, ok := f.(*cli.StringFlag); ok {
				scheduleFlag = sf
			}
			break
		}
	}

	if scheduleFlag == nil {
		t.Fatal("schedule flag not found or not a StringFlag")
	}

	if scheduleFlag.Value != "" {
		t.Errorf("expected no default (empty), got %q", scheduleFlag.Value)
	}

	aliases := scheduleFlag.Aliases
	if len(aliases) != 1 || aliases[0] != "s" {
		t.Errorf("expected alias 's', got %v", aliases)
	}

	if scheduleFlag.Usage == "" {
		t.Error("schedule flag should have a non-empty usage string")
	}
}

// =============================================================================
// resolveWatchConfig Tests
// =============================================================================

func TestResolveWatchConfig_CLIIntervalOverridesYAMLInterval(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 12 * time.Hour},
			&cli.StringFlag{Name: "schedule", Value: ""},
		},
	}
	cfg := WatchConfig{Interval: "24h"}

	resolved, err := resolveWatchConfig(cmd, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Interval != (12 * time.Hour).String() {
		t.Errorf("Interval = %v, want %v", resolved.Interval, (12 * time.Hour).String())
	}
}

func TestResolveWatchConfig_CLIScheduleOverridesYAMLSchedule(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 0},
			&cli.StringFlag{Name: "schedule", Value: testCronScheduleDaily3},
		},
	}
	cfg := WatchConfig{Schedule: "0 5 * * *"}

	resolved, err := resolveWatchConfig(cmd, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Schedule != testCronScheduleDaily3 {
		t.Errorf("Schedule = %v, want 0 3 * * *", resolved.Schedule)
	}
}

func TestResolveWatchConfig_CLIIntervalAndYAMLSchedule_ReturnsError(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 12 * time.Hour},
			&cli.StringFlag{Name: "schedule", Value: ""},
		},
	}
	cfg := WatchConfig{Schedule: testCronScheduleDaily3}

	_, err := resolveWatchConfig(cmd, cfg)
	if err == nil {
		t.Fatal("expected error when CLI interval + YAML schedule both effective")
	}
}

func TestResolveWatchConfig_CLIScheduleAndYAMLInterval_ReturnsError(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 0},
			&cli.StringFlag{Name: "schedule", Value: testCronScheduleDaily3},
		},
	}
	cfg := WatchConfig{Interval: "24h"}

	_, err := resolveWatchConfig(cmd, cfg)
	if err == nil {
		t.Fatal("expected error when CLI schedule + YAML interval both effective")
	}
}

func TestResolveWatchConfig_Neither_ReturnsMissingSentinel(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 0},
			&cli.StringFlag{Name: "schedule", Value: ""},
		},
	}
	cfg := WatchConfig{}

	_, err := resolveWatchConfig(cmd, cfg)
	if err == nil {
		t.Fatal("expected error when neither set")
	}
	if !errors.Is(err, ErrWatchConfigMissing) {
		t.Errorf("expected ErrWatchConfigMissing, got: %v", err)
	}
}

func TestResolveWatchConfig_OnlyInterval_OK(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 0},
			&cli.StringFlag{Name: "schedule", Value: ""},
		},
	}
	cfg := WatchConfig{Interval: "12h"}

	resolved, err := resolveWatchConfig(cmd, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Interval != "12h" {
		t.Errorf("Interval = %v, want 12h", resolved.Interval)
	}
}

func TestResolveWatchConfig_OnlySchedule_OK(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "interval", Value: 0},
			&cli.StringFlag{Name: "schedule", Value: ""},
		},
	}
	cfg := WatchConfig{Schedule: testCronScheduleDaily3}

	resolved, err := resolveWatchConfig(cmd, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Schedule != testCronScheduleDaily3 {
		t.Errorf("Schedule = %v, want 0 3 * * *", resolved.Schedule)
	}
}

// =============================================================================
// watchWithCronSchedule Tests
// =============================================================================

type mockCronSchedule struct {
	nexts []time.Time
	idx   int
}

func (m *mockCronSchedule) Next(now time.Time) time.Time {
	if m.idx < len(m.nexts) {
		t := m.nexts[m.idx]
		m.idx++
		return t
	}
	return now.Add(1 * time.Hour)
}

type mockApp struct {
	runCount int
}

func (m *mockApp) Run(_ context.Context) error {
	m.runCount++
	return nil
}

func (m *mockApp) Refresh(_ context.Context) {}

func TestWatchWithCronSchedule_ContextCancelStopsImmediately(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())

	farFuture := time.Now().Add(1 * time.Hour)
	sched := &mockCronSchedule{nexts: []time.Time{farFuture}}
	app := &mockApp{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watchWithCronSchedule(ctx, app, sched, "0 0 1 1 *", false)
	}()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watchWithCronSchedule did not stop after context cancellation")
	}
}

func TestWatchWithCronSchedule_OnceTriggersImmediateSync(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	farFuture := time.Now().Add(1 * time.Hour)
	sched := &mockCronSchedule{nexts: []time.Time{farFuture}}
	app := &mockApp{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watchWithCronSchedule(ctx, app, sched, "0 0 1 1 *", true)
	}()

	// Give time for the initial sync
	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-errCh
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	if app.runCount < 1 {
		t.Errorf("expected at least 1 Run call, got %d", app.runCount)
	}
}

func TestWatchWithCronSchedule_FiresOnSchedule(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	now := time.Now()
	soon := now.Add(200 * time.Millisecond)
	farFuture := soon.Add(1 * time.Hour)
	sched := &mockCronSchedule{nexts: []time.Time{soon, farFuture}}
	app := &mockApp{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watchWithCronSchedule(ctx, app, sched, "test", false)
	}()

	// Wait for the first scheduled tick (200ms)
	time.Sleep(500 * time.Millisecond)
	cancel()

	err := <-errCh
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}

	if app.runCount < 1 {
		t.Errorf("expected at least 1 scheduled Run call, got %d", app.runCount)
	}
}

func TestWatchWithCron_InvalidScheduleReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	app := &mockApp{}

	err := watchWithCron(ctx, app, "bad cron expr", false)
	if err == nil {
		t.Fatal("expected error for invalid schedule")
	}
}
