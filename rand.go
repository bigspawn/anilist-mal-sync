package main

import (
	"time"

	"golang.org/x/exp/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randHttpParamString(n int) string {
	rand.Seed(uint64(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(letters[rand.Intn(len(letters))])
	}
	return string(b)
}
