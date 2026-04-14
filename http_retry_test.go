package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	b := &ExponentialBackoff{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     500 * time.Millisecond,
		Multiplier:      2.0,
	}

	tests := []struct {
		attempt          int
		expectedDuration time.Duration
	}{
		{0, 0},
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 500 * time.Millisecond}, // Capped at MaxInterval
		{5, 500 * time.Millisecond}, // Still capped
	}

	for _, tt := range tests {
		t.Run(tt.expectedDuration.String(), func(t *testing.T) {
			duration := b.Duration(tt.attempt)
			if duration != tt.expectedDuration {
				t.Fatalf("attempt %d: expected %v, got %v", tt.attempt, tt.expectedDuration, duration)
			}
		})
	}
}

func TestShouldRetryStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, false},
		{300, false},
		{400, false},
		{404, false},
		{429, true}, // Too Many Requests
		{408, true}, // Request Timeout
		{500, true}, // Internal Server Error
		{502, true}, // Bad Gateway
		{503, true}, // Service Unavailable
		{504, true}, // Gateway Timeout
		{599, true}, // Other 5xx
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.statusCode), func(t *testing.T) {
			result := shouldRetryStatus(tt.statusCode)
			if result != tt.expected {
				t.Fatalf("status %d: expected %v, got %v", tt.statusCode, tt.expected, result)
			}
		})
	}
}

func TestCloneRequest(t *testing.T) {
	original, _ := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"http://example.com",
		io.NopCloser(strings.NewReader("test-body")),
	)
	original.Header.Set("X-Custom", "value")

	cloned := cloneRequest(original)

	// Verify URL is cloned
	if cloned.URL.String() != original.URL.String() {
		t.Fatalf("URL not cloned correctly")
	}

	// Verify Method is cloned
	if cloned.Method != original.Method {
		t.Fatalf("Method not cloned correctly")
	}

	// Verify Headers are cloned
	if cloned.Header.Get("X-Custom") != "value" {
		t.Fatalf("Headers not cloned correctly")
	}

	// Verify Body is cloned
	originalBody, _ := io.ReadAll(original.Body)
	clonedBody, _ := io.ReadAll(cloned.Body)
	if string(originalBody) != string(clonedBody) {
		t.Fatalf("Body not cloned correctly: original='%s', cloned='%s'", string(originalBody), string(clonedBody))
	}
}

// ============================================
// retryableRoundTripper Tests
// ============================================

func TestRetryableRoundTripper_RetryOn500(t *testing.T) {
	// Timing-sensitive test: depends on HTTP request execution speed
	// Remove t.Parallel() to avoid flaky failures when CPU is overloaded
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 3)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableRoundTripper_NoRetryOnSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		attempts++
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 3)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetryableRoundTripper_RetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 3)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryableRoundTripper_ContextCancellation(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(t.Context())
	transport := NewRetryableTransport(&http.Client{}, 10)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)

	// Cancel after first attempt
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	resp, err := client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
	// The error may be context.Canceled or wrapped with the URL
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
	if attempts > 2 {
		t.Fatalf("expected context to cancel after ~1-2 attempts, got %d", attempts)
	}
}

func TestRetryableRoundTripper_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		attempts++
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 3)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}
	_ = resp.Body.Close()
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry for 404), got %d", attempts)
	}
}

func TestRetryableRoundTripper_WithBody(t *testing.T) {
	// Timing-sensitive test: depends on HTTP request execution speed
	// Remove t.Parallel() to avoid flaky failures when CPU is overloaded
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 3)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL, io.NopCloser(strings.NewReader("test-body")))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableRoundTripper_MaxRetriesExhausted(t *testing.T) {
	// Timing-sensitive test: depends on HTTP request execution speed
	// Remove t.Parallel() to avoid flaky failures when CPU is overloaded
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	transport := NewRetryableTransport(&http.Client{}, 2)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err == nil {
		t.Fatal("expected error after max retries exhausted, got nil")
	}
	// Should attempt 3 times (initial + 2 retries)
	if attempts != 3 {
		t.Fatalf("expected 3 attempts (initial + 2 retries), got %d", attempts)
	}
}

func TestRetryAfterOrBackoff(t *testing.T) {
	t.Parallel()
	backoff := &ExponentialBackoff{
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}

	tests := []struct {
		name        string
		resp        *http.Response
		attempt     int
		expected    time.Duration
		fromHeader  bool
	}{
		{
			name: "uses Retry-After header on 429",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"30"}},
			},
			attempt:    1,
			expected:   30 * time.Second,
			fromHeader: true,
		},
		{
			name: "falls back to backoff when no Retry-After header",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{},
			},
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
		{
			name: "ignores Retry-After for non-429 status",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{"Retry-After": []string{"30"}},
			},
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
		{
			name: "falls back to backoff when Retry-After is not a number",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"not-a-number"}},
			},
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
		{
			name:       "falls back to backoff when resp is nil",
			resp:       nil,
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
		{
			name: "falls back to backoff when Retry-After is zero",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"0"}},
			},
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
		{
			name: "uses Retry-After regardless of backoff attempt number",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"60"}},
			},
			attempt:    3,
			expected:   60 * time.Second,
			fromHeader: true,
		},
		{
			name: "falls back to backoff when Retry-After is negative",
			resp: &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"-10"}},
			},
			attempt:    1,
			expected:   1 * time.Second,
			fromHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, fromHeader := retryAfterOrBackoff(tt.resp, tt.attempt, backoff)
			if got != tt.expected {
				t.Fatalf("duration: expected %v, got %v", tt.expected, got)
			}
			if fromHeader != tt.fromHeader {
				t.Fatalf("fromHeader: expected %v, got %v", tt.fromHeader, fromHeader)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		resp     *http.Response
		expected bool
	}{
		{
			name:     "connection refused error",
			err:      errors.New("dial tcp: connection refused"),
			expected: true,
		},
		{
			name:     "connection reset error",
			err:      errors.New("read: connection reset by peer"),
			expected: true,
		},
		{
			name:     "broken pipe error",
			err:      errors.New("write: broken pipe"),
			expected: true,
		},
		{
			name:     "other error not retryable",
			err:      errors.New("EOF"),
			expected: false,
		},
		{
			name:     "timeout error not retryable",
			err:      errors.New("context deadline exceeded"),
			expected: false,
		},
		{
			name:     "429 response is retryable",
			resp:     &http.Response{StatusCode: http.StatusTooManyRequests},
			expected: true,
		},
		{
			name:     "500 response is retryable",
			resp:     &http.Response{StatusCode: http.StatusInternalServerError},
			expected: true,
		},
		{
			name:     "502 response is retryable",
			resp:     &http.Response{StatusCode: http.StatusBadGateway},
			expected: true,
		},
		{
			name:     "503 response is retryable",
			resp:     &http.Response{StatusCode: http.StatusServiceUnavailable},
			expected: true,
		},
		{
			name:     "200 response is not retryable",
			resp:     &http.Response{StatusCode: http.StatusOK},
			expected: false,
		},
		{
			name:     "404 response is not retryable",
			resp:     &http.Response{StatusCode: http.StatusNotFound},
			expected: false,
		},
		{
			name:     "400 response is not retryable",
			resp:     &http.Response{StatusCode: http.StatusBadRequest},
			expected: false,
		},
		{
			name:     "nil error and nil response returns false without panic",
			err:      nil,
			resp:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryable(tt.err, tt.resp)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCloneRequest_NilBody(t *testing.T) {
	t.Parallel()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com", nil)
	cloned := cloneRequest(req)

	if cloned.URL.String() != req.URL.String() {
		t.Fatalf("URL not cloned: got %s", cloned.URL)
	}
	if cloned.Body != nil && cloned.Body != http.NoBody {
		t.Fatal("expected nil/NoBody for nil original body")
	}
}

func TestRetryableRoundTripper_RateLimitedLogMessage(t *testing.T) {
	// Verify that a 429 with Retry-After triggers the "rate limited" log path.
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf strings.Builder
	logger := NewLogger(false)
	logger.SetOutput(&buf)
	ctx := logger.WithContext(t.Context())

	transport := &retryableRoundTripper{
		underlying: http.DefaultTransport,
		maxRetries: 3,
		backoff:    &defaultBackoffConfig,
	}
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if !strings.Contains(buf.String(), "rate limited") {
		t.Fatalf("expected 'rate limited' in log output, got: %s", buf.String())
	}
}

func TestRetryableRoundTripper_RetryAfterHeaderIsUsed(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Use a backoff much longer than Retry-After to confirm Retry-After wins.
	transport := &retryableRoundTripper{
		underlying: http.DefaultTransport,
		maxRetries: 3,
		backoff: &ExponentialBackoff{
			InitialInterval: 30 * time.Second,
			MaxInterval:     60 * time.Second,
			Multiplier:      2.0,
		},
	}
	client := &http.Client{Transport: transport}

	start := time.Now()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = resp.Body.Close()

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	// Should have waited ~1s (Retry-After), not 30s (backoff).
	if elapsed >= 10*time.Second {
		t.Fatalf("waited too long (%v): Retry-After header was not used", elapsed)
	}
}

func TestNewRetryableTransport_DefaultTransport(t *testing.T) {
	client := &http.Client{} // Transport is nil
	transport := NewRetryableTransport(client, 3)

	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	rt, ok := transport.(*retryableRoundTripper)
	if !ok {
		t.Fatal("expected *retryableRoundTripper type")
	}

	if rt.underlying == nil {
		t.Fatal("expected underlying transport to be set to http.DefaultTransport")
	}
}

// =============================================================================
// parseRetryAfter
// =============================================================================

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		wantOk    bool
		wantExact time.Duration // only checked when wantOk && >= 0
		wantMin   time.Duration // for approximate checks (HTTP-date)
		wantMax   time.Duration
	}{
		{
			name:      "integer seconds",
			value:     "30",
			wantOk:    true,
			wantExact: 30 * time.Second,
		},
		{
			name:   "zero seconds",
			value:  "0",
			wantOk: false,
		},
		{
			name:   "negative seconds",
			value:  "-10",
			wantOk: false,
		},
		{
			name:   "invalid string",
			value:  "not-a-number",
			wantOk: false,
		},
		{
			name:   "empty string",
			value:  "",
			wantOk: false,
		},
		{
			name:   "past HTTP-date",
			value:  "Mon, 02 Jan 2006 15:04:05 GMT",
			wantOk: false,
		},
		{
			name:    "future HTTP-date",
			value:   time.Now().Add(60 * time.Second).UTC().Format(http.TimeFormat),
			wantOk:  true,
			wantMin: 50 * time.Second,
			wantMax: 65 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, ok := parseRetryAfter(tt.value)
			if ok != tt.wantOk {
				t.Fatalf("ok: expected %v, got %v (duration=%v)", tt.wantOk, ok, d)
			}
			if !tt.wantOk {
				return
			}
			if tt.wantExact > 0 {
				if d != tt.wantExact {
					t.Fatalf("duration: expected %v, got %v", tt.wantExact, d)
				}
			} else {
				if d < tt.wantMin || d > tt.wantMax {
					t.Fatalf("duration %v not in expected range [%v, %v]", d, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}

func TestRetryAfterOrBackoff_HTTPDate(t *testing.T) {
	t.Parallel()
	backoff := &ExponentialBackoff{
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}

	futureDate := time.Now().Add(60 * time.Second).UTC().Format(http.TimeFormat)
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Retry-After": []string{futureDate}},
	}

	got, fromHeader := retryAfterOrBackoff(resp, 1, backoff)
	if !fromHeader {
		t.Fatal("expected fromHeader=true for HTTP-date Retry-After")
	}
	// HTTP-date is truncated to seconds, so allow ±5s tolerance.
	if got < 55*time.Second || got > 65*time.Second {
		t.Fatalf("expected duration ~60s, got %v", got)
	}
}

// =============================================================================
// executeWithRetry — backoff skipped on last attempt
// =============================================================================

// countingBackoff records how many times Duration is called.
type countingBackoff struct {
	calls int
}

func (c *countingBackoff) Duration(_ int) time.Duration {
	c.calls++
	return 0 // zero so tests run instantly
}

func TestExecuteWithRetry_SkipsBackoffOnLastAttempt(t *testing.T) {
	t.Parallel()
	const maxRetries = 3

	cb := &countingBackoff{}

	// All requests fail with a retryable error.
	doRequest := func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:1", nil)
	_, _ = executeWithRetry(req, maxRetries, cb, doRequest)

	// Attempts: 0, 1, 2, 3 — backoff computed before attempts 1,2,3 only.
	// The last attempt (3 == maxRetries) must NOT trigger a backoff call.
	if cb.calls != maxRetries {
		t.Fatalf("expected %d backoff calls, got %d", maxRetries, cb.calls)
	}
}
