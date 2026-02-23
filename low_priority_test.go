package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rl404/verniy"
)

// ============================================
// Logger Tests - Additional unique tests
// ============================================

func TestLogger_Colorize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		useColor bool
		color    string
		text     string
	}{
		{
			name:     "with color enabled",
			useColor: true,
			color:    "\033[31m",
			text:     "error",
		},
		{
			name:     "with color disabled",
			useColor: false,
			color:    "\033[31m",
			text:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := &Logger{useColors: tt.useColor}
			result := logger.colorize(tt.color, tt.text)

			if tt.useColor {
				if !strings.Contains(result, tt.color) {
					t.Errorf("expected result to contain color code")
				}
				if !strings.Contains(result, tt.text) {
					t.Errorf("expected result to contain text")
				}
				if !strings.Contains(result, "\033[0m") {
					t.Errorf("expected result to contain reset code")
				}
			} else if result != tt.text {
				t.Errorf("expected %q, got %q", tt.text, result)
			}
		})
	}
}

func TestLog_WithContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewLogger(false)
	logger.SetOutput(&buf)

	ctx := logger.WithContext(context.Background())

	LogInfo(ctx, "test info")
	if !strings.Contains(buf.String(), "test info") {
		t.Error("expected 'test info' in output")
	}

	buf.Reset()
	LogWarn(ctx, "test warn")
	if !strings.Contains(buf.String(), "test warn") {
		t.Error("expected 'test warn' in output")
	}

	buf.Reset()
	LogError(ctx, "test error")
	if !strings.Contains(buf.String(), "test error") {
		t.Error("expected 'test error' in output")
	}
}

// ============================================
// OAuth Tests
// ============================================

func TestCreateDirIfNotExists_NewDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Only create a single subdirectory, not nested ones
	newDirPath := filepath.Join(tmpDir, "newdir", "file.json")

	err := createDirIfNotExists(newDirPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The function creates the parent directory, not the full path
	parentDir := filepath.Join(tmpDir, "newdir")
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected parent directory to be created")
	}
}

func TestCreateDirIfNotExists_ExistingDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := createDirIfNotExists(tmpDir)
	if err != nil {
		t.Errorf("unexpected error for existing dir: %v", err)
	}
}

// ============================================
// Config Tests - Additional unique tests
// ============================================

func TestGetEnvOrDefault(t *testing.T) {
	t.Parallel()

	// Set a test env var
	t.Setenv("TEST_VAR", "test_value")

	// Test with existing env var
	got := getEnvOrDefault("TEST_VAR", "default")
	if got != "test_value" {
		t.Errorf("got %q, want %q", got, "test_value")
	}

	// Test with non-existing env var
	got = getEnvOrDefault("NON_EXISTING_VAR", "default")
	if got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
}

func TestGetDefaultTokenPathOrEmpty(t *testing.T) {
	t.Parallel()

	path := getDefaultTokenPathOrEmpty()

	// Should return a non-empty path on most systems
	// The path format depends on the OS
	if path != "" {
		if !strings.Contains(path, "anilist-mal-sync") {
			t.Error("expected path to contain 'anilist-mal-sync'")
		}
		if !strings.Contains(path, "token.json") {
			t.Error("expected path to contain 'token.json'")
		}
	}
}

// ============================================
// Rand Tests
// ============================================

func TestRandHTTPParamString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		length  int
		wantLen int
	}{
		{
			name:    "length 16",
			length:  16,
			wantLen: 16,
		},
		{
			name:    "length 32",
			length:  32,
			wantLen: 32,
		},
		{
			name:    "length 64",
			length:  64,
			wantLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := randHTTPParamString(tt.length)
			if len(got) != tt.wantLen {
				t.Errorf("got length %d, want %d", len(got), tt.wantLen)
			}

			// Check that all characters are from the expected set
			validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, c := range got {
				if !strings.ContainsRune(validChars, c) {
					t.Errorf("invalid character %q in result", c)
				}
			}

			// Multiple calls should produce different strings (very high probability)
			got2 := randHTTPParamString(tt.length)
			if tt.length >= 16 && got == got2 {
				t.Error("multiple calls produced the same string (unlikely)")
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	timeout := 100 * time.Millisecond

	newCtx, cancel := withTimeout(ctx, timeout)

	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	if cancel == nil {
		t.Fatal("expected non-nil cancel function")
	}

	// Verify deadline is set
	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Error("expected context to have deadline")
	}
	if !deadline.After(time.Now()) {
		t.Error("expected deadline to be in the future")
	}

	cancel()
}

// ============================================
// Utils Tests
// ============================================

func TestMin3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b, c int
		want    int
	}{
		{
			name: "a is smallest",
			a:    1, b: 5, c: 10,
			want: 1,
		},
		{
			name: "b is smallest",
			a:    10, b: 2, c: 5,
			want: 2,
		},
		{
			name: "c is smallest",
			a:    10, b: 5, c: 3,
			want: 3,
		},
		{
			name: "all equal",
			a:    5, b: 5, c: 5,
			want: 5,
		},
		{
			name: "negative values",
			a:    -5, b: 0, c: 5,
			want: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := min3(tt.a, tt.b, tt.c)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "lowercase conversion",
			title: "Neon Genesis Evangelion",
			want:  "neon genesis evangelion",
		},
		{
			name:  "removes content in parentheses",
			title: "Anime (TV)",
			want:  "anime",
		},
		{
			name:  "removes colons",
			title: "Fullmetal Alchemist: Brotherhood",
			want:  "fullmetal alchemist brotherhood",
		},
		{
			name:  "removes exclamation marks",
			title: "Bocchi the Rock!",
			want:  "bocchi the rock",
		},
		{
			name:  "removes question marks",
			title: "What's the Title?",
			want:  "what's the title", // apostrophes are kept
		},
		{
			name:  "removes periods",
			title: "Dr. Stone",
			want:  "dr stone",
		},
		{
			name:  "replaces dashes with spaces",
			title: "one-piece",
			want:  "one piece",
		},
		{
			name:  "replaces underscores with spaces",
			title: "one_piece_anime",
			want:  "one piece anime",
		},
		{
			name:  "trims spaces",
			title: "  spaced out title  ",
			want:  "spaced out title",
		},
		{
			name:  "multiple spaces - only double becomes single in single pass",
			title: "title  with    double   spaces",
			want:  "title with  double  spaces", // single-pass replacement
		},
		{
			name:  "complex normalization",
			title: "Fullmetal Alchemist: Brotherhood (TV)",
			want:  "fullmetal alchemist brotherhood",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeTitle(tt.title)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExactMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		t1, t2     string
		titleType  string
		wantResult bool
	}{
		{
			name:       "exact same titles",
			t1:         "Neon Genesis Evangelion",
			t2:         "Neon Genesis Evangelion",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "case insensitive match",
			t1:         "Neon Genesis Evangelion",
			t2:         "neon genesis evangelion",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "different titles",
			t1:         "Neon Genesis Evangelion",
			t2:         "Cowboy Bebop",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "empty t1",
			t1:         "",
			t2:         "Some Title",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "empty t2",
			t1:         "Some Title",
			t2:         "",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "both empty",
			t1:         "",
			t2:         "",
			titleType:  "test",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := exactMatch(tt.t1, tt.t2, tt.titleType)
			if got != tt.wantResult {
				t.Errorf("got %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestNormalizedMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		t1, t2     string
		titleType  string
		wantResult bool
	}{
		{
			name:       "exact match",
			t1:         "Neon Genesis Evangelion",
			t2:         "Neon Genesis Evangelion",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "normalized match - punctuation",
			t1:         "Fullmetal Alchemist: Brotherhood",
			t2:         "Fullmetal Alchemist Brotherhood",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "normalized match - parentheses",
			t1:         "Anime (TV)",
			t2:         "Anime",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "normalized match - case",
			t1:         "NEON GENESIS EVANGELION",
			t2:         "neon genesis evangelion",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "different titles",
			t1:         "Neon Genesis Evangelion",
			t2:         "Cowboy Bebop",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "empty t1",
			t1:         "",
			t2:         "Some Title",
			titleType:  "test",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizedMatch(tt.t1, tt.t2, tt.titleType)
			if got != tt.wantResult {
				t.Errorf("got %v, want %v", got, tt.wantResult)
			}
		})
	}
}

// ============================================
// Logging Tests
// ============================================

func TestNewLoggingRoundTripper_NilBase(t *testing.T) {
	t.Parallel()

	rt := newLoggingRoundTripper(nil, false)

	if rt == nil {
		t.Fatal("expected non-nil round tripper")
	}
	lrt, ok := rt.(*loggingRoundTripper)
	if !ok {
		t.Fatal("expected *loggingRoundTripper type")
	}
	if lrt.base == nil {
		t.Error("expected non-nil base transport")
	}
}

func TestNewLoggingRoundTripper_WithBase(t *testing.T) {
	t.Parallel()

	base := &http.Transport{}
	rt := newLoggingRoundTripper(base, true)

	if rt == nil {
		t.Fatal("expected non-nil round tripper")
	}
	lrt, ok := rt.(*loggingRoundTripper)
	if !ok {
		t.Fatal("expected *loggingRoundTripper type")
	}
	if lrt.base != base {
		t.Error("expected base to be the provided transport")
	}
	if !lrt.verbose {
		t.Error("expected verbose to be true")
	}
}

// ============================================
// Anime Utility Tests
// ============================================

func TestBuildDiffString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		pairs []any
		want  string
	}{
		{
			name:  "no differences",
			pairs: []any{"Status", "watching", "watching", "Score", 8, 8},
			want:  "Diff{}",
		},
		{
			name:  "single difference",
			pairs: []any{"Status", "watching", "completed", "Score", 8, 8},
			want:  "Diff{Status: watching -> completed, }",
		},
		{
			name:  "multiple differences",
			pairs: []any{"Status", "watching", "completed", "Score", 8, 9},
			want:  "Diff{Status: watching -> completed, Score: 8 -> 9, }",
		},
		{
			name:  "empty pairs",
			pairs: []any{},
			want:  "Diff{}",
		},
		{
			name:  "invalid params - not multiple of 3",
			pairs: []any{"Status", "watching"},
			want:  "Diff{invalid params}",
		},
		{
			name:  "non-string field name",
			pairs: []any{123, "value1", "value2"},
			want:  "Diff{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildDiffString(tt.pairs...)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertFuzzyDateToTimeOrNow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fd   *verniy.FuzzyDate
		want *time.Time
	}{
		{
			name: "nil fuzzy date",
			fd:   nil,
			want: nil,
		},
		{
			name: "nil year",
			fd:   &verniy.FuzzyDate{Year: nil, Month: intPtr(1), Day: intPtr(1)},
			want: nil,
		},
		{
			name: "nil month",
			fd:   &verniy.FuzzyDate{Year: intPtr(2023), Month: nil, Day: intPtr(1)},
			want: nil,
		},
		{
			name: "nil day",
			fd:   &verniy.FuzzyDate{Year: intPtr(2023), Month: intPtr(1), Day: nil},
			want: nil,
		},
		{
			name: "valid date",
			fd:   &verniy.FuzzyDate{Year: intPtr(2023), Month: intPtr(6), Day: intPtr(15)},
			want: timePtr(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := convertFuzzyDateToTimeOrNow(tt.fd)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Fatal("got nil, want non-nil")
				}
				if !got.Equal(*tt.want) {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestParseDateOrNow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dateStr string
		wantNil bool
	}{
		{
			name:    "empty string",
			dateStr: "",
			wantNil: true,
		},
		{
			name:    "invalid date format",
			dateStr: "not-a-date",
			wantNil: true,
		},
		{
			name:    "valid date",
			dateStr: "2023-06-15",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDateOrNow(tt.dateStr)
			if tt.wantNil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Fatal("got nil, want non-nil")
				}
				// Verify it's truncated to day
				if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
					t.Errorf("expected time to be truncated to day, got %v", got)
				}
			}
		})
	}
}

// ============================================
// Helper Functions
// ============================================

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// ============================================
// Status Error Tests
// ============================================

func TestErrStatusUnknown(t *testing.T) {
	t.Parallel()

	err := errStatusUnknown
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "status unknown" {
		t.Errorf("got %q, want 'status unknown'", err.Error())
	}
}

func TestErrMangaStatusUnknown(t *testing.T) {
	t.Parallel()

	err := errMangaStatusUnknown
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != "manga status unknown" {
		t.Errorf("got %q, want 'manga status unknown'", err.Error())
	}
}
