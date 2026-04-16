package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingRoundTripper_NonVerbose_DoesNotLog(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf strings.Builder
	logger := NewLogger(true) // verbose logger
	logger.SetOutput(&buf)

	rt := &loggingRoundTripper{base: http.DefaultTransport, verbose: false}
	client := &http.Client{Transport: rt}

	ctx := logger.WithContext(t.Context())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req) //nolint:gosec
	assert.NoError(t, err)
	_ = resp.Body.Close()

	// Non-verbose mode produces no log output at all.
	assert.Empty(t, buf.String(), "non-verbose mode should produce no log output")
}

func TestLoggingRoundTripper_Verbose_LogsRequest(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf strings.Builder
	logger := NewLogger(true)
	logger.SetOutput(&buf)
	logger.level = LogLevelDebug

	rt := &loggingRoundTripper{base: http.DefaultTransport, verbose: true}
	client := &http.Client{Transport: rt}

	ctx := logger.WithContext(t.Context())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req) //nolint:gosec
	assert.NoError(t, err)
	_ = resp.Body.Close()

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "200")
}

func TestLoggingRoundTripper_Verbose_LogsError(t *testing.T) {
	t.Parallel()
	errTransport := &errorTransport{err: errors.New("connection refused")}

	var buf strings.Builder
	logger := NewLogger(true)
	logger.SetOutput(&buf)
	logger.level = LogLevelDebug

	rt := &loggingRoundTripper{base: errTransport, verbose: true}
	client := &http.Client{Transport: rt}

	ctx := logger.WithContext(t.Context())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:1", nil)
	resp, err := client.Do(req) //nolint:gosec
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "failed")
}

func TestLoggingRoundTripper_NonVerbose_PassesThrough(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	rt := &loggingRoundTripper{base: http.DefaultTransport, verbose: false}
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req) //nolint:gosec
	assert.NoError(t, err)
	_ = resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestLoggingRoundTripper_NonVerbose_PropagatesError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("transport error")
	rt := &loggingRoundTripper{
		base:    &errorTransport{err: wantErr},
		verbose: false,
	}
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:1", nil)
	resp, err := client.Do(req) //nolint:gosec
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.ErrorIs(t, err, wantErr)
}

// errorTransport is a test double that always returns an error.
type errorTransport struct {
	err error
}

func (e *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, e.err
}
