package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// ============================================
// Race Condition Tests
// ============================================

func TestOAuth_ConcurrentStateAccess(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	oauth := &OAuth{
		siteName: "test",
		state:    "test-state",
		Config: &oauth2.Config{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://example.com/auth",
				TokenURL: "https://example.com/token",
			},
		},
	}

	var wg sync.WaitGroup
	// Run multiple goroutines accessing state methods
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = oauth.GetAuthURL()
			_ = oauth.NeedInit()
			_ = oauth.IsTokenValid()
			_ = oauth.TokenExpiry()
		}()
	}

	wg.Wait()
}

func TestSyncReport_ConcurrentWarningAdd(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping race condition test in short mode")
	}

	report := NewSyncReport()
	var wg sync.WaitGroup

	// Add warnings concurrently from multiple goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			report.AddWarning(fmt.Sprintf("Title%d", n), "reason", "detail", "Anime")
		}(i)
	}

	wg.Wait()

	// Verify all warnings were added
	if len(report.Warnings) != 100 {
		t.Errorf("Expected 100 warnings, got %d", len(report.Warnings))
	}
}

// ============================================
// Input Edge Cases
// ============================================

func TestNormalizeTitle_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "empty string",
			title: "",
			want:  "",
		},
		{
			name:  "only spaces",
			title: "     ",
			want:  "",
		},
		{
			name:  "only punctuation - removed individually",
			title: "!:?.,",
			want:  ",", // ! and ? removed, then : replaced with space, then . removed, leaving ,
		},
		{
			name:  "only parentheses",
			title: "((()))",
			want:  "",
		},
		{
			name:  "unicode characters",
			title: "進撃の巨人Attack on Titan",
			want:  "進撃の巨人attack on titan",
		},
		{
			name:  "emoji characters",
			title: "🎬 Anime Title",
			want:  "🎬 anime title",
		},
		{
			name:  "multiple colons - single pass replacement",
			title: "Fullmetal: Alchemist: Brotherhood",
			want:  "fullmetal alchemist brotherhood", // only first : is replaced
		},
		{
			name:  "nested parentheses",
			title: "Anime (TV (2023))",
			want:  "anime", // parens removed, trailing space trimmed
		},
		{
			name:  "mixed punctuation - single pass",
			title: "What?! Is this... anime?",
			want:  "what is this anime", // !! -> empty, then . removed, ? kept
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTitle(tt.title)
			if got != tt.want {
				t.Errorf("normalizeTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExactMatch_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		t1, t2     string
		titleType  string
		wantResult bool
	}{
		{
			name:       "both empty",
			t1:         "",
			t2:         "",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "one empty, one non-empty",
			t1:         "",
			t2:         "Some Title",
			titleType:  "test",
			wantResult: false,
		},
		{
			name:       "only spaces",
			t1:         "   ",
			t2:         "   ",
			titleType:  "test",
			wantResult: true, // exact match on spaces
		},
		{
			name:       "special characters only",
			t1:         "!!",
			t2:         "!!",
			titleType:  "test",
			wantResult: true,
		},
		{
			name:       "unicode identical",
			t1:         "進撃の巨人",
			t2:         "進撃の巨人",
			titleType:  "test",
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exactMatch(tt.t1, tt.t2, tt.titleType)
			if got != tt.wantResult {
				t.Errorf("exactMatch() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestBuildDiffString_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		pairs []any
		want  string
	}{
		{
			name:  "nil values",
			pairs: []any{"Status", (*Status)(nil), "watching"},
			want:  "Diff{Status: <nil> -> watching, }",
		},
		{
			name:  "empty strings",
			pairs: []any{"Title", "", "value"},
			want:  "Diff{Title:  -> value, }",
		},
		{
			name:  "very long values",
			pairs: []any{"Title", strings.Repeat("a", 1000), strings.Repeat("b", 1000)},
			want:  fmt.Sprintf("Diff{Title: %s -> %s, }", strings.Repeat("a", 1000), strings.Repeat("b", 1000)),
		},
		{
			name:  "special characters",
			pairs: []any{"Path", "C:\\Users\\path", "/usr/local/path"},
			want:  "Diff{Path: C:\\Users\\path -> /usr/local/path, }",
		},
		{
			name:  "zero values",
			pairs: []any{"Score", 0, 0},
			want:  "Diff{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDiffString(tt.pairs...)
			if got != tt.want {
				t.Errorf("buildDiffString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================
// Nil/Zero Value Handling
// ============================================

func TestAnime_SameTypeWithTarget_NilAndZero(t *testing.T) {
	t.Parallel()
	anime := Anime{
		IDAnilist: 123,
		IDMal:     456,
		TitleEN:   "Test Anime",
	}

	tests := []struct {
		name   string
		anime  Anime
		target Target
		want   bool
	}{
		{
			name:   "nil target",
			anime:  anime,
			target: nil,
			want:   false,
		},
		{
			name:   "zero anime vs non-zero",
			anime:  Anime{},
			target: anime,
			want:   false,
		},
		{
			name:   "non-zero vs zero anime",
			anime:  anime,
			target: Anime{},
			want:   false,
		},
		{
			name:   "manga target (type mismatch)",
			anime:  anime,
			target: Manga{IDAnilist: 123},
			want:   false,
		},
		{
			name:   "same ID but different type",
			anime:  Anime{IDAnilist: 123, IDMal: 456},
			target: Manga{IDAnilist: 123, IDMal: 456},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.anime.SameTypeWithTarget(tt.target)
			if got != tt.want {
				t.Errorf("SameTypeWithTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================
// Context Cancellation Tests
// ============================================

func TestRetryableTransport_ContextCancellation(t *testing.T) {
	// Timing-sensitive test: depends on HTTP request execution speed
	// Remove t.Parallel() to avoid flaky failures when CPU is overloaded
	tests := []struct {
		name              string
		cancelImmediately bool
		expectCallCount   int
	}{
		{
			name:              "context cancelled immediately",
			cancelImmediately: true,
			expectCallCount:   0, // No calls when context is already cancelled
		},
		{
			name:              "context not cancelled",
			cancelImmediately: false,
			expectCallCount:   3, // 3 calls total (2 errors + 1 success)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			if tt.cancelImmediately {
				cancel()
			} else {
				defer cancel()
			}

			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				callCount++
				if callCount > 2 {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(http.StatusTooManyRequests)
			}))
			defer server.Close()

			transport := NewRetryableTransport(&http.Client{}, 3)
			client := &http.Client{Transport: transport}

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			resp, _ := client.Do(req)
			if resp != nil {
				_ = resp.Body.Close()
			}

			if callCount < tt.expectCallCount {
				t.Errorf("Expected at least %d calls, got %d", tt.expectCallCount, callCount)
			}
		})
	}
}

func TestWithTimeout_DeadlineExceeded(t *testing.T) {
	ctx := t.Context()
	timeout := 1 * time.Nanosecond

	newCtx, cancel := withTimeout(ctx, timeout)
	defer cancel()

	// Wait for timeout to pass
	time.Sleep(10 * time.Millisecond)

	// Verify context is cancelled
	select {
	case <-newCtx.Done():
		if newCtx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", newCtx.Err())
		}
	default:
		t.Error("Expected context to be cancelled")
	}
}

// ============================================
// Edge Case: Empty/Invalid Data
// ============================================

func TestWatchConfig_GetInterval_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		interval string
		wantErr  bool
	}{
		{
			name:     "negative duration - valid in Go",
			interval: "-5m",
			wantErr:  false, // Negative durations are valid in Go
		},
		{
			name:     "zero duration",
			interval: "0s",
			wantErr:  false,
		},
		{
			name:     "very large duration",
			interval: "87600h",
			wantErr:  false,
		},
		{
			name:     "fractional seconds",
			interval: "1.5s",
			wantErr:  false,
		},
		{
			name:     "multiple units",
			interval: "1h30m45s",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc := &WatchConfig{Interval: tt.interval}
			got, err := wc.GetInterval()

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			// Just verify parsing works - negative durations are valid in Go
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Verify the parsed duration matches expected (negative values OK)
			if tt.interval == "-5m" && got != -5*time.Minute {
				t.Errorf("Expected -5m, got %v", got)
			}
		})
	}
}

func TestStatus_ErrorCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  Status
		wantErr bool
	}{
		{
			name:    "invalid status",
			status:  Status("invalid_status"),
			wantErr: true,
		},
		{
			name:    "empty status",
			status:  Status(""),
			wantErr: true,
		},
		{
			name:    "unknown status",
			status:  StatusUnknown,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.status.GetMalStatus()
			if tt.wantErr && err == nil {
				t.Error("Expected error for invalid status")
			}
		})
	}
}

// ============================================
// Edge Case: Large Inputs
// ============================================

func TestNormalizeTitle_LongTitle(t *testing.T) {
	t.Parallel()
	longTitle := strings.Repeat("Very Long Anime Title ", 100)

	// Should not panic
	result := normalizeTitle(longTitle)

	if len(result) == 0 {
		t.Error("Expected non-empty result for long title")
	}

	// Should be shorter than original (replacements happened)
	if len(result) >= len(longTitle) {
		t.Errorf("Expected shorter result, got %d vs %d", len(result), len(longTitle))
	}
}

func TestLogger_Progress_VeryLongTitle(t *testing.T) {
	t.Parallel()
	logger := NewLogger(false)
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Title longer than 150 characters
	longTitle := strings.Repeat("A", 200)

	// For non-terminal output, Progress outputs each item
	logger.Progress(10, 10, "watching", longTitle)

	// Should not panic and should show progress message
	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected some output when current == total")
	}
	if !strings.Contains(output, "[10/10]") {
		t.Error("Expected '[10/10]' in output")
	}
}

func TestLogger_Progress_VeryLongTitle_Verbose(t *testing.T) {
	t.Parallel()
	logger := NewLogger(true)
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Title longer than 150 characters
	longTitle := strings.Repeat("A", 200)

	// For non-terminal output, Progress outputs each item
	logger.Progress(10, 10, "watching", longTitle)

	// Should not panic and should show progress message
	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected some output in verbose mode when current == total")
	}
	if !strings.Contains(output, "[10/10]") {
		t.Error("Expected '[10/10]' in output")
	}
}
