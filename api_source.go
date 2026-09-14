package main

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// apiSource holds the HTTP client and circuit-breaker state shared by
// third-party ID-mapping clients (Hato, MangaBaka). These are small
// third-party services that go down for long stretches; once one does,
// every lookup costs several retries with backoff, so the breaker stops
// asking for the rest of the run.
type apiSource struct {
	name        string
	baseURL     string
	httpClient  HTTPClient
	maxFailures int

	mu            sync.Mutex
	failureStreak int
	givenUp       bool
}

// newAPISource builds an apiSource with a retrying HTTP client.
func newAPISource(name, baseURL string, timeout time.Duration, maxFailures int) apiSource {
	return apiSource{
		name:    name,
		baseURL: baseURL,
		httpClient: NewRetryableClient(&http.Client{
			Timeout: timeout,
		}, 3),
		maxFailures: maxFailures,
	}
}

// gaveUp reports whether the client stopped talking to this API for this run.
func (s *apiSource) gaveUp() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.givenUp
}

// noteFailure counts a failed request and stops the client once the API looks down.
func (s *apiSource) noteFailure(ctx context.Context, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failureStreak++
	if s.givenUp || s.failureStreak < s.maxFailures {
		return
	}

	s.givenUp = true
	LogWarn(ctx, "%s API failed %d times in a row (%v), skipping it for the rest of this run",
		s.name, s.failureStreak, err)
}

// noteSuccess clears the streak — the service answered, whatever the answer was.
func (s *apiSource) noteSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failureStreak = 0
}
