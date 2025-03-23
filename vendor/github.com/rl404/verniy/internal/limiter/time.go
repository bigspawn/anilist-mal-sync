package limiter

import (
	"time"
)

// newClock returns an instance of a real-time clock.
// Taken from `github.com/andres-erbsen/clock`.
func newClock() *clock {
	return &clock{}
}

// clock implements a real-time clock by simply wrapping the time package functions.
type clock struct{}

// Now to get time now.
func (c *clock) Now() time.Time {
	return time.Now()
}

// Sleep to sleep.
func (c *clock) Sleep(d time.Duration) {
	time.Sleep(d)
}
