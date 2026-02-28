package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_InfoDryRun(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)

	logger.InfoDryRun("Would update (%d):", 5)

	output := buf.String()
	assert.Contains(t, output, "→", "Should contain arrow icon")
	assert.Contains(t, output, "Would update (5)", "Should contain message")
}

func TestLogger_InfoDryRunWithColor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewLogger(true)
	logger.SetOutput(&buf)

	logger.InfoDryRun("Test message")

	output := buf.String()
	assert.Contains(t, output, "→", "Should contain arrow icon")
	assert.Contains(t, output, "Test message", "Should contain message")
}

func TestLogger_InfoDryRunBelowLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewLogger(false) // LogLevelInfo
	logger.SetOutput(&buf)

	// Set level to LogLevelError - InfoDryRun should not output
	logger.level = LogLevelError
	logger.InfoDryRun("Should not appear")

	output := buf.String()
	assert.NotContains(t, output, "Should not appear", "Should not log when below level")
	assert.NotContains(t, output, "→", "Should not contain arrow when below level")
}
