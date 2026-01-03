package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// randHTTPParamString generates a random string of length n for HTTP parameters.
// Returns an error if random generation fails (should never happen in practice).
func randHTTPParamString(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid length: %d", n)
	}

	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}
