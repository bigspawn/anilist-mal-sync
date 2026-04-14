package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// defaultBackoffConfig defines the default exponential backoff settings.
var defaultBackoffConfig = ExponentialBackoff{
	InitialInterval: 1 * time.Second,
	MaxInterval:     30 * time.Second,
	Multiplier:      2.0,
}

// BackoffStrategy defines retry delay behavior.
type BackoffStrategy interface {
	Duration(attempt int) time.Duration
}

// ExponentialBackoff implements exponential backoff.
type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

func (b *ExponentialBackoff) Duration(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}
	delay := float64(b.InitialInterval) * math.Pow(b.Multiplier, float64(attempt-1))
	if delay > float64(b.MaxInterval) {
		return b.MaxInterval
	}
	return time.Duration(delay)
}

// shouldRetryStatus determines if a response status code should trigger a retry.
func shouldRetryStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		(statusCode >= 500 && statusCode < 600)
}

// isRetryable determines if an error or response should trigger a retry.
func isRetryable(err error, resp *http.Response) bool {
	if err != nil {
		errStr := err.Error()
		return strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "connection reset") ||
			strings.Contains(errStr, "broken pipe")
	}
	return shouldRetryStatus(resp.StatusCode)
}

// cloneRequest creates a copy of an HTTP request, including its body.
func cloneRequest(req *http.Request) *http.Request {
	r := req.Clone(req.Context())
	if req.Body != nil && req.Body != http.NoBody {
		body, _ := io.ReadAll(req.Body)
		_ = req.Body.Close()
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		req.Body = io.NopCloser(strings.NewReader(string(body)))
	}
	return r
}

// retryableRoundTripper implements http.RoundTripper with retry logic.
type retryableRoundTripper struct {
	underlying http.RoundTripper
	maxRetries int
	backoff    BackoffStrategy
}

// NewRetryableTransport wraps an http.Client's Transport with retry logic.
func NewRetryableTransport(baseClient *http.Client, maxRetries int) http.RoundTripper {
	underlying := baseClient.Transport
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	return &retryableRoundTripper{
		underlying: underlying,
		maxRetries: maxRetries,
		backoff:    &defaultBackoffConfig,
	}
}

// RoundTrip executes a single HTTP transaction with retry logic.
func (t *retryableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return executeWithRetry(req, t.maxRetries, t.backoff, func(req *http.Request) (*http.Response, error) {
		return t.underlying.RoundTrip(req)
	})
}

// executeWithRetry executes an HTTP request function with retry logic.
// This is a shared implementation used by both RoundTripper and RetryableClient.
func executeWithRetry(
	req *http.Request,
	maxRetries int,
	backoff BackoffStrategy,
	doRequest func(*http.Request) (*http.Response, error),
) (*http.Response, error) {
	var lastErr error
	var nextWait time.Duration
	var rateLimited bool

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if rateLimited {
				LogWarn(req.Context(), "[HTTP RETRY] Attempt %d/%d for %s (rate limited, waiting %v)",
					attempt, maxRetries, req.URL, nextWait)
			} else {
				LogWarn(req.Context(), "[HTTP RETRY] Attempt %d/%d for %s (waiting %v)",
					attempt, maxRetries, req.URL, nextWait)
			}

			select {
			case <-time.After(nextWait):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		reqClone := cloneRequest(req)
		resp, err := doRequest(reqClone)

		if err == nil && !shouldRetryStatus(resp.StatusCode) {
			return resp, nil
		}

		// Compute the wait for the next retry before closing the response
		// so we can read the Retry-After header if present.
		nextWait, rateLimited = retryAfterOrBackoff(resp, attempt+1, backoff)

		if resp != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			lastErr = err
		}

		if !isRetryable(err, resp) {
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("max retries (%d) exhausted", maxRetries)
}

// retryAfterOrBackoff returns the wait duration before the next retry and
// whether the duration came from a Retry-After header (true) or backoff (false).
// For 429 Too Many Requests responses with a Retry-After header, it uses
// that value to wait precisely until the rate limit window resets.
// Falls back to exponential backoff otherwise.
func retryAfterOrBackoff(resp *http.Response, attempt int, backoff BackoffStrategy) (time.Duration, bool) {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if v := resp.Header.Get("Retry-After"); v != "" {
			if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second, true
			}
		}
	}
	return backoff.Duration(attempt), false
}

// withTimeout adds a timeout to the context for API calls.
func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// HTTPClient interface for flexibility and testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetryableClient wraps HTTPClient with retry logic.
type RetryableClient struct {
	client     HTTPClient
	maxRetries int
	backoff    BackoffStrategy
}

// NewRetryableClient creates a new RetryableClient with default backoff strategy.
func NewRetryableClient(baseClient *http.Client, maxRetries int) *RetryableClient {
	return &RetryableClient{
		client:     baseClient,
		maxRetries: maxRetries,
		backoff:    &defaultBackoffConfig,
	}
}

// Do executes the HTTP request with retry logic.
func (r *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	return executeWithRetry(req, r.maxRetries, r.backoff, r.client.Do)
}
