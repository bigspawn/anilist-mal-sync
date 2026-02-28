package main

import (
	"context"
	"testing"
)

// setTestDirection returns a context with the given direction embedded.
func setTestDirection(t *testing.T, dir SyncDirection) context.Context {
	t.Helper()
	return WithDirection(t.Context(), dir)
}
