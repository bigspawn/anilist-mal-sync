package main

import (
	"crypto/rand"
	"math/big"
)

// letters must contain only ASCII characters (no runes outside 0-127) because
// the byte conversion below depends on it — converting a non-ASCII rune to byte would truncate.
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randHTTPParamString(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		b[i] = byte(letters[num.Int64()]) // #nosec G115 -- letters is ASCII-only, so the byte conversion cannot truncate
	}
	return string(b)
}
