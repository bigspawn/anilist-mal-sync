package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// Test NewLogger
func TestNewLogger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "verbose mode",
			verbose: true,
		},
		{
			name:    "quiet mode",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.verbose)

			if logger == nil {
				t.Error("NewLogger should not return nil")
			}
		})
	}
}

func TestNewLogger_VerboseFlag(t *testing.T) {
	t.Parallel()
	loggerVerbose := NewLogger(true)
	loggerQuiet := NewLogger(false)

	if loggerVerbose.level != LogLevelDebug {
		t.Error("Verbose logger should have Debug level")
	}
	if loggerQuiet.level != LogLevelInfo {
		t.Error("Quiet logger should have Info level")
	}
}

func TestLogger_SetOutput(t *testing.T) {
	t.Parallel()
	logger := NewLogger(false)

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("SetOutput should redirect output")
	}
}

func TestLogger_WithContext(t *testing.T) {
	t.Parallel()
	logger := NewLogger(false)
	ctx := context.Background()

	ctxWithLogger := logger.WithContext(ctx)

	loggerFromCtx := LoggerFromContext(ctxWithLogger)
	if loggerFromCtx == nil {
		t.Error("WithContext should store logger in context")
	}
	if loggerFromCtx != logger {
		t.Error("LoggerFromContext should return the same logger")
	}
}

func TestLoggerFromContext_EmptyContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := LoggerFromContext(ctx)

	if logger != nil {
		t.Error("LoggerFromContext should return nil for empty context")
	}
}

func TestLoggerContext_Propagation(t *testing.T) {
	t.Parallel()
	logger1 := NewLogger(true)
	ctx1 := context.Background()
	ctx1 = logger1.WithContext(ctx1)

	logger2 := NewLogger(false)
	ctx2 := logger2.WithContext(ctx1)

	// ctx2 should have logger2 (overwrites logger1)
	retrieved := LoggerFromContext(ctx2)
	if retrieved != logger2 {
		t.Error("Context should have the latest logger")
	}

	if retrieved == logger1 {
		t.Error("Retrieved logger should not be the first logger")
	}
}

// Test NewTokenFile
func TestNewTokenFile(t *testing.T) {
	t.Parallel()
	tokenFile := NewTokenFile()

	if tokenFile == nil {
		t.Error("NewTokenFile should not return nil")
	}
}

// Test NewStrategyChain
func TestNewStrategyChain_Empty(t *testing.T) {
	t.Parallel()
	chain := NewStrategyChain()

	if len(chain.strategies) != 0 {
		t.Errorf("Empty chain should have 0 strategies, got %d", len(chain.strategies))
	}
}

func TestNewStrategyChain_WithStrategies(t *testing.T) {
	t.Parallel()
	strategy1 := IDStrategy{}
	strategy2 := TitleStrategy{}
	strategy3 := MALIDStrategy{}

	chain := NewStrategyChain(strategy1, strategy2, strategy3)

	if len(chain.strategies) != 3 {
		t.Errorf("Chain should have 3 strategies, got %d", len(chain.strategies))
	}
}

// Test NewSyncReport
func TestNewSyncReport_Structure(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	if report.Warnings == nil {
		t.Error("Warnings should be initialized (not nil)")
	}
	if len(report.Warnings) != 0 {
		t.Error("Warnings should be empty initially")
	}
	if report.HasWarnings() {
		t.Error("HasWarnings should return false for empty report")
	}
}

// Test Statistics methods
func TestStatistics_IncrementTotal(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.IncrementTotal()
	if stats.TotalCount != 1 {
		t.Errorf("After 1 IncrementTotal, TotalCount = %d, want 1", stats.TotalCount)
	}

	stats.IncrementTotal()
	stats.IncrementTotal()
	if stats.TotalCount != 3 {
		t.Errorf("After 3 IncrementTotal, TotalCount = %d, want 3", stats.TotalCount)
	}
}

func TestStatistics_StatusCounts(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordUpdate(UpdateResult{Title: "Test1", Status: "watching"})
	stats.RecordUpdate(UpdateResult{Title: "Test2", Status: "watching"})
	stats.RecordSkip(UpdateResult{Title: "Test3", Status: "completed", SkipReason: "test"})

	if stats.StatusCounts["watching"] != 2 {
		t.Errorf("StatusCounts[watching] = %d, want 2", stats.StatusCounts["watching"])
	}
	if stats.StatusCounts["completed"] != 1 {
		t.Errorf("StatusCounts[completed] = %d, want 1", stats.StatusCounts["completed"])
	}
	if stats.UpdatedCount != 2 {
		t.Errorf("UpdatedCount = %d, want 2", stats.UpdatedCount)
	}
	if stats.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", stats.SkippedCount)
	}
}

func TestStatistics_RecordError(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	testErr := errors.New("test error")
	stats.RecordError(UpdateResult{
		Title:  "Failed Anime",
		Status: "watching",
		Error:  testErr,
	})

	if stats.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", stats.ErrorCount)
	}

	if len(stats.ErrorItems) != 1 {
		t.Errorf("ErrorItems length = %d, want 1", len(stats.ErrorItems))
	}

	if stats.ErrorItems[0].Title != "Failed Anime" {
		t.Errorf("ErrorItems[0].Title = %s, want 'Failed Anime'", stats.ErrorItems[0].Title)
	}
}

func TestStatistics_StartTime(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	if stats.StartTime.IsZero() {
		t.Error("StartTime should be set by NewStatistics")
	}

	// Reset should update StartTime
	oldTime := stats.StartTime
	time.Sleep(10 * time.Millisecond)
	stats.Reset()

	if stats.StartTime.Before(oldTime) {
		t.Error("Reset should update StartTime")
	}
}

func TestStatistics_ResetClearsSlices(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	stats.RecordUpdate(UpdateResult{Title: "Test", Status: "watching"})
	stats.RecordSkip(UpdateResult{Title: "Test", Status: "completed", SkipReason: "test"})

	if len(stats.UpdatedItems) != 1 {
		t.Error("UpdatedItems should have 1 item before Reset")
	}

	stats.Reset()

	if len(stats.UpdatedItems) != 0 {
		t.Error("UpdatedItems should be empty after Reset")
	}
	if len(stats.SkippedItems) != 0 {
		t.Error("SkippedItems should be empty after Reset")
	}
	if len(stats.ErrorItems) != 0 {
		t.Error("ErrorItems should be empty after Reset")
	}
}

func TestStatistics_StatusCountsMap(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	// StatusCounts should be initialized
	if stats.StatusCounts == nil {
		t.Error("StatusCounts should be initialized")
	}

	// Add some counts
	stats.StatusCounts["watching"] = 5
	stats.StatusCounts["completed"] = 10

	if stats.StatusCounts["watching"] != 5 {
		t.Error("Failed to set StatusCounts value")
	}

	// Reset should clear the map
	stats.Reset()

	if len(stats.StatusCounts) != 0 {
		t.Error("StatusCounts should be empty after Reset")
	}
}

func TestStatistics_ItemsSlices(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	// Initially empty
	if len(stats.UpdatedItems) != 0 {
		t.Error("UpdatedItems should be empty initially")
	}
	if len(stats.SkippedItems) != 0 {
		t.Error("SkippedItems should be empty initially")
	}
	if len(stats.ErrorItems) != 0 {
		t.Error("ErrorItems should be empty initially")
	}

	// Add items
	stats.RecordUpdate(UpdateResult{Title: "U1", Status: "watching"})
	stats.RecordSkip(UpdateResult{Title: "S1", Status: "completed", SkipReason: "test"})
	stats.RecordError(UpdateResult{Title: "E1", Status: "dropped", Error: errors.New("err")})

	if len(stats.UpdatedItems) != 1 {
		t.Error("UpdatedItems should have 1 item")
	}
	if len(stats.SkippedItems) != 1 {
		t.Error("SkippedItems should have 1 item")
	}
	if len(stats.ErrorItems) != 1 {
		t.Error("ErrorItems should have 1 item")
	}
}

func TestStatistics_EndTime(t *testing.T) {
	t.Parallel()
	stats := NewStatistics()

	if stats.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if !stats.EndTime.IsZero() {
		t.Error("EndTime should be zero initially")
	}

	// Print should set EndTime
	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(context.Background())

	stats.Print(ctx, "Test")

	if stats.EndTime.IsZero() {
		t.Error("Print should set EndTime")
	}

	if stats.EndTime.Before(stats.StartTime) {
		t.Error("EndTime should be after StartTime")
	}
}

// Test multiple statistics instances
func TestMultipleStatisticsInstances(t *testing.T) {
	t.Parallel()
	stats1 := NewStatistics()
	stats2 := NewStatistics()

	stats1.IncrementTotal()
	stats2.IncrementTotal()
	stats2.IncrementTotal()

	if stats1.TotalCount != 1 {
		t.Errorf("stats1.TotalCount = %d, want 1", stats1.TotalCount)
	}
	if stats2.TotalCount != 2 {
		t.Errorf("stats2.TotalCount = %d, want 2", stats2.TotalCount)
	}
}

// Test Config methods
func TestConfig_GetHTTPTimeout_Default(t *testing.T) {
	t.Parallel()
	config := Config{
		HTTPTimeout: "", // empty means default
	}

	got := config.GetHTTPTimeout()
	expected := 30 * time.Second // default is 30s

	if got != expected {
		t.Errorf("GetHTTPTimeout() = %v, want %v", got, expected)
	}
}

func TestConfig_GetHTTPTimeout_Custom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		httpTimeout string
		expected    time.Duration
	}{
		{
			name:        "30 seconds",
			httpTimeout: "30s",
			expected:    30 * time.Second,
		},
		{
			name:        "60 seconds",
			httpTimeout: "60s",
			expected:    60 * time.Second,
		},
		{
			name:        "120 seconds",
			httpTimeout: "120s",
			expected:    120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				HTTPTimeout: tt.httpTimeout,
			}

			got := config.GetHTTPTimeout()
			if got != tt.expected {
				t.Errorf("GetHTTPTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test error detection functions
func TestIsConfigNotFoundError_Wrapped(t *testing.T) {
	t.Parallel()
	// Test with wrapped error
	baseErr := errors.New("config file not found")
	err := fmt.Errorf("failed to load config: %w", baseErr)

	got := IsConfigNotFoundError(err)
	if !got {
		t.Error("IsConfigNotFoundError should return true for wrapped config error")
	}

	// Test with non-config error
	otherErr := errors.New("some other error")
	got = IsConfigNotFoundError(otherErr)
	if got {
		t.Error("IsConfigNotFoundError should return false for other error")
	}
}

func TestIsConfigNotFoundError_Direct(t *testing.T) {
	t.Parallel()
	// Test with nil error
	got := IsConfigNotFoundError(nil)
	if got {
		t.Error("IsConfigNotFoundError should return false for nil error")
	}

	// Test with direct string error (not wrapped)
	strErr := errors.New("config not found")
	got = IsConfigNotFoundError(strErr)
	if got {
		t.Error("IsConfigNotFoundError should return false for direct string error")
	}
}

func TestIsCancellationError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "context canceled error",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "deadline exceeded error",
			err:      context.DeadlineExceeded,
			expected: false, // Not considered a cancellation error by the implementation
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCancellationError(tt.err)
			if got != tt.expected {
				t.Errorf("IsCancellationError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCLIUsageError_Wrapped(t *testing.T) {
	t.Parallel()
	// Test with wrapped error - implementation doesn't unwrap errors,
	// so this should return false unless the outer message matches
	baseErr := errors.New("flag parsing failed")
	err := fmt.Errorf("cli error: %w", baseErr)

	got := IsCLIUsageError(err)
	if got {
		t.Error("IsCLIUsageError should return false for wrapped cli error (doesn't unwrap)")
	}

	// Test with non-cli error
	otherErr := errors.New("some other error")
	got = IsCLIUsageError(otherErr)
	if got {
		t.Error("IsCLIUsageError should return false for other error")
	}
}

func TestIsCLIUsageError_Direct(t *testing.T) {
	t.Parallel()
	// Test with nil error
	got := IsCLIUsageError(nil)
	if got {
		t.Error("IsCLIUsageError should return false for nil error")
	}

	// Test with direct string error (not wrapped)
	strErr := errors.New("cli error")
	got = IsCLIUsageError(strErr)
	if got {
		t.Error("IsCLIUsageError should return false for direct string error")
	}
}

// Test UpdateResult struct
func TestUpdateResult_IsSkipped(t *testing.T) {
	t.Parallel()
	result := UpdateResult{
		Title:      "Test",
		Status:     "watching",
		Skipped:    true,
		SkipReason: "no changes",
	}

	if !result.Skipped {
		t.Error("Expected Skipped to be true")
	}
}

func TestUpdateResult_HasError(t *testing.T) {
	t.Parallel()
	testErr := errors.New("test error")
	result := UpdateResult{
		Title:  "Failed",
		Status: "watching",
		Error:  testErr,
	}

	if result.Error == nil {
		t.Error("Expected Error to be set")
	}
}

func TestUpdateResult_EmptyFields(t *testing.T) {
	t.Parallel()
	result := UpdateResult{}

	if result.Title != "" {
		t.Error("Empty result should have empty Title")
	}
	if result.Status != "" {
		t.Error("Empty result should have empty Status")
	}
	if result.Skipped {
		t.Error("Empty result should not be skipped")
	}
	if result.Error != nil {
		t.Error("Empty result should have nil error")
	}
}

// Test TargetID type
func TestTargetID_Conversion(t *testing.T) {
	t.Parallel()
	id := TargetID(12345)
	// TargetID is just an int alias, but we can test conversions
	if int(id) != 12345 {
		t.Errorf("TargetID conversion failed: %d != 12345", int(id))
	}
}

func TestTargetID_ZeroValue(t *testing.T) {
	t.Parallel()
	var id TargetID
	if id != 0 {
		t.Errorf("Zero TargetID should be 0, got %d", id)
	}
}

// Test Warning struct
func TestWarning_Structure(t *testing.T) {
	t.Parallel()
	warning := Warning{
		Title:     "Test Title",
		Reason:    "Test Reason",
		Detail:    "(1 vs 12)",
		MediaType: "Anime",
	}

	if warning.Title != "Test Title" {
		t.Errorf("Title = %s, want 'Test Title'", warning.Title)
	}
	if warning.Reason != "Test Reason" {
		t.Errorf("Reason = %s, want 'Test Reason'", warning.Reason)
	}
	if warning.Detail != "(1 vs 12)" {
		t.Errorf("Detail = %s, want '(1 vs 12)'", warning.Detail)
	}
	if warning.MediaType != "Anime" {
		t.Errorf("MediaType = %s, want 'Anime'", warning.MediaType)
	}
}

func TestWarning_EmptyFields(t *testing.T) {
	t.Parallel()
	warning := Warning{}

	if warning.Title != "" {
		t.Error("Empty warning should have empty Title")
	}
	if warning.Reason != "" {
		t.Error("Empty warning should have empty Reason")
	}
}

// Test SyncReport Warnings
func TestReport_WarningsField(t *testing.T) {
	t.Parallel()
	report := NewSyncReport()

	// After our fix to NewSyncReport, Warnings should be initialized (not nil)
	if report.Warnings == nil {
		t.Error("Warnings should be initialized (not nil)")
	}

	// But it should be empty
	if len(report.Warnings) != 0 {
		t.Error("Warnings should be empty initially")
	}

	// HasWarnings should return false
	if report.HasWarnings() {
		t.Error("HasWarnings should return false for empty report")
	}

	// Add a warning
	report.AddWarning("Test", "reason", "detail", "Anime")

	if !report.HasWarnings() {
		t.Error("HasWarnings should return true after adding warning")
	}

	if len(report.Warnings) != 1 {
		t.Errorf("Warnings should have 1 item, got %d", len(report.Warnings))
	}
}

// Test helper functions
func TestBufferUsage(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	buf.WriteString("test")
	if buf.String() != "test" {
		t.Error("Buffer should contain 'test'")
	}

	buf.Reset()
	if buf.String() != "" {
		t.Error("Buffer should be empty after Reset")
	}
}
