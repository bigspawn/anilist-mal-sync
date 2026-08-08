package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/nstratos/go-myanimelist/mal"
	"github.com/stretchr/testify/assert"
)

// newStalledMALClient points a MAL client at a server that never answers, so a
// call can only return by hitting its own timeout.
func newStalledMALClient(t *testing.T, timeout time.Duration) *MyAnimeListClient {
	t.Helper()

	// Released at cleanup: a handler parked on the request context alone would
	// keep Close waiting on a connection the client has already abandoned.
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() {
		close(release)
		server.Close()
	})

	base, err := url.Parse(server.URL + "/")
	assert.NoError(t, err)

	client := mal.NewClient(nil)
	client.BaseURL = base

	// username is unused by the update calls under test.
	return &MyAnimeListClient{c: client, httpTimeout: timeout}
}

// A stalled write used to hang forever: the update calls were the only ones
// that never wrapped the context in a timeout.
func TestMyAnimeListClient_UpdateAnimeByIDAndOptions_TimesOut(t *testing.T) {
	t.Parallel()

	client := newStalledMALClient(t, 100*time.Millisecond)

	start := time.Now()
	err := client.UpdateAnimeByIDAndOptions(t.Context(), 164, []mal.UpdateMyAnimeListStatusOption{
		mal.AnimeStatusCompleted,
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second, "must give up on its own timeout")
}

func TestMyAnimeListClient_UpdateMangaByIDAndOptions_TimesOut(t *testing.T) {
	t.Parallel()

	client := newStalledMALClient(t, 100*time.Millisecond)

	start := time.Now()
	err := client.UpdateMangaByIDAndOptions(t.Context(), 164, []mal.UpdateMyMangaListStatusOption{
		mal.MangaStatusCompleted,
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second, "must give up on its own timeout")
}
