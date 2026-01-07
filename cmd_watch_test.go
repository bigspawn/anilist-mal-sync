package main

import (
	"errors"
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

	// Check default value
	if intervalFlag.Value != defaultInterval {
		t.Errorf("expected default interval %v, got %v", defaultInterval, intervalFlag.Value)
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
