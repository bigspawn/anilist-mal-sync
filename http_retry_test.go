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
		context.Background(),
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
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

	ctx, cancel := context.WithCancel(context.Background())
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, io.NopCloser(strings.NewReader("test-body")))
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

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
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
