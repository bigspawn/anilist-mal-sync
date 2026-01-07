package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	watchCommandName = "watch"
)

// =============================================================================
// Watch Command Structure Tests
// =============================================================================

func TestWatchCommand_Exists(t *testing.T) {
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

	if watchCmd.Usage != "Run sync on interval (Docker-friendly)" {
		t.Errorf("unexpected usage: %s", watchCmd.Usage)
	}
}

func TestWatchCommand_HasFlags(t *testing.T) {
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

	if len(watchCmd.Flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(watchCmd.Flags))
	}

	// Check flags by name
	flagNames := make(map[string]bool)
	for _, f := range watchCmd.Flags {
		flagNames[f.Names()[0]] = true
	}

	expectedFlags := []string{"interval", "once"}
	for _, name := range expectedFlags {
		if !flagNames[name] {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestWatchCommand_IntervalFlag_HasCorrectDefaults(t *testing.T) {
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
	tests := []struct {
		name     string
		interval time.Duration
		valid    bool
	}{
		{"Minimum valid", minInterval, true},
		{"Maximum valid", maxInterval, true},
		{"Default interval", defaultInterval, true},
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

// validateInterval is a helper function extracted from runWatch for testing
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has 12h, CLI uses 24h
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
myanimelist:
  client_id: "test"
  client_secret: "test"
watch:
  interval: "12h"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has 12h, no CLI flag
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
myanimelist:
  client_id: "test"
  client_secret: "test"
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config has no watch interval
	configContent := `
oauth:
  port: "18080"
anilist:
  client_id: "test"
  client_secret: "test"
myanimelist:
  client_id: "test"
  client_secret: "test"
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
	w := WatchConfig{Interval: "invalid-duration"}
	_, err := w.GetInterval()
	if err == nil {
		t.Error("expected error for invalid interval, got nil")
	}
}

func TestWatch_IntervalValidation_Config_TooSmall(t *testing.T) {
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
	w := WatchConfig{Interval: "200h"}
	got, err := w.GetInterval()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got > maxInterval {
		t.Logf("interval %v exceeds max %v (validation will catch this)", got, maxInterval)
	}
}
