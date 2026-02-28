package main

import (
	"context"
	"testing"
)

// setTestDirection sets both the context direction and the global reverseDirection variable.
// This is needed because GetTargetID() and GetSourceID() use the global variable, not the context.
func setTestDirection(t *testing.T, dir SyncDirection) context.Context {
	t.Helper()
	*reverseDirection = (dir == SyncDirectionReverse)
	return WithDirection(t.Context(), dir)
}

// setTestDirectionFromCtx sets both the context direction and the global reverseDirection variable
// from an existing context. This is useful when you need to add direction to a context that already
// has other values (like a logger).
func setTestDirectionFromCtx(ctx context.Context, dir SyncDirection) context.Context {
	*reverseDirection = (dir == SyncDirectionReverse)
	return WithDirection(ctx, dir)
}
