package main

import (
	"net/http"
	"time"
)

// loggingRoundTripper wraps an http.RoundTripper and logs HTTP requests/responses in verbose mode.
type loggingRoundTripper struct {
	base    http.RoundTripper
	verbose bool
}

// newLoggingRoundTripper creates a new logging round tripper.
func newLoggingRoundTripper(base http.RoundTripper, verbose bool) http.RoundTripper {
	if base == nil {
		return &loggingRoundTripper{base: http.DefaultTransport, verbose: verbose}
	}
	return &loggingRoundTripper{base: base, verbose: verbose}
}

// RoundTrip executes a single HTTP transaction and logs the request/response.
func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	if *verbose {
		LogDebugHTTP(ctx, "%s %s", req.Method, req.URL)
		start := time.Now()

		resp, err := l.base.RoundTrip(req)
		elapsed := time.Since(start)

		if err != nil {
			LogDebugHTTP(ctx, "%s %s failed: %v (took %v)", req.Method, req.URL, err, elapsed)
			return nil, err
		}

		LogDebugHTTP(ctx, "%s %s -> %d (took %v)", req.Method, req.URL, resp.StatusCode, elapsed)
		return resp, nil
	}

	return l.base.RoundTrip(req)
}
