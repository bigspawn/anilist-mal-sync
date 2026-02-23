package main

import (
	"sync"
)

// testReverseMutex protects reverseDirection in parallel tests
var testReverseMutex sync.Mutex

// setReverseDirectionForTest safely sets reverseDirection for testing and returns cleanup function
func setReverseDirectionForTest(reverse bool) func() {
	testReverseMutex.Lock()
	origReverse := reverseDirection
	reverseDirection = &reverse
	return func() {
		reverseDirection = origReverse
		testReverseMutex.Unlock()
	}
}

// setReverseDirectionForTestNoLock sets reverseDirection without locking (caller must hold lock)
// Used in tests that need to change direction multiple times
func setReverseDirectionForTestNoLock(reverse bool) {
	reverseDirection = &reverse
}

// lockReverseDirection locks the reverseDirection mutex for manual control
func lockReverseDirection() func() {
	testReverseMutex.Lock()
	return func() {
		testReverseMutex.Unlock()
	}
}
