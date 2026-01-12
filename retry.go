package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/nstratos/go-myanimelist/mal"
)

// createBackoffPolicy creates a configured exponential backoff policy for retrying transient errors
func createBackoffPolicy() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 2 * time.Minute
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.1
	return b
}

// isRetryableError checks if the error should trigger a retry
// Supports both AniList (string-based) and MAL (*mal.ErrorResponse) errors
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for MAL ErrorResponse with HTTP status code
	var malErr *mal.ErrorResponse
	if errors.As(err, &malErr) && malErr.Response != nil {
		code := malErr.Response.StatusCode
		// Retry on: 429 (rate limit), 503 (service unavailable), 502 (bad gateway), 500 (internal server error)
		return code == 429 || code == 503 || code == 502 || code == 500
	}

	// Fallback to string-based check for AniList and other errors
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "500")
}

// retryWithBackoff wraps an operation with exponential backoff for retrying transient errors
func retryWithBackoff(ctx context.Context, operation func() error, operationName string, prefix ...string) error {
	b := createBackoffPolicy()

	var attemptCount int
	retryableOperation := func() error {
		err := operation()
		if err != nil && !isRetryableError(err) {
			// Don't retry non-transient errors
			return backoff.Permanent(err)
		}
		return err
	}

	return backoff.RetryNotify(
		retryableOperation,
		backoff.WithContext(b, ctx),
		func(err error, duration time.Duration) {
			if isRetryableError(err) {
				attemptCount++
				if len(prefix) > 0 {
					// Log retry attempt - VISIBLE to user (not verbose only)
					log.Printf("[%s] Retry attempt %d for %s (waiting %v)...", prefix[0], attemptCount, operationName, duration)
				} else {
					log.Printf("Retry attempt %d for %s (waiting %v)...", attemptCount, operationName, duration)
				}
			}
		},
	)
}

// withTimeout adds a 10-second timeout to the context for API calls
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}
