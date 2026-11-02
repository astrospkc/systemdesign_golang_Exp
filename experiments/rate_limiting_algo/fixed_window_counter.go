package ratelimitingalgo

import (
    "sync"
    "time"
)

// FixedWindowCounter implements a simple fixed window rate limiter.
// It allows up to `limit` requests within each `windowSize` duration.
// The window is aligned to the first request after the previous window expires.
// This implementation is safe for concurrent use.
type FixedWindowCounter struct {
    limit      int
    windowSize time.Duration
    mu         sync.Mutex
    // start of the current window
    windowStart time.Time
    // number of requests seen in the current window
    count int
}

// NewFixedWindowCounter creates a new FixedWindowCounter.
func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
    return &FixedWindowCounter{
        limit:      limit,
        windowSize: windowSize,
        // windowStart is zero value; first Allow will set it.
    }
}

// Allow reports whether a request is permitted at the current time.
func (f *FixedWindowCounter) Allow() bool {
    f.mu.Lock()
    defer f.mu.Unlock()
    now := time.Now()
    // Initialize window on first call
    if f.windowStart.IsZero() {
        f.windowStart = now
    }
    // If the current window has elapsed, reset.
    if now.Sub(f.windowStart) >= f.windowSize {
        f.windowStart = now
        f.count = 0
    }
    if f.count < f.limit {
        f.count++
        return true
    }
    return false
}
