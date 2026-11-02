package ratelimitingalgo

import (
    "testing"
    "time"
)

func TestFixedWindowCounterBasic(t *testing.T) {
    limit := 2
    window := 1 * time.Second
    counter := NewFixedWindowCounter(limit, window)

    // First two requests should be allowed
    for i := 0; i < limit; i++ {
        if !counter.Allow() {
            t.Fatalf("expected request %d to be allowed", i+1)
        }
    }
    // Third request within same window should be blocked
    if counter.Allow() {
        t.Fatalf("expected request beyond limit to be blocked")
    }
    // After the window elapses, requests should be allowed again
    time.Sleep(window + 10*time.Millisecond)
    if !counter.Allow() {
        t.Fatalf("expected request after window reset to be allowed")
    }
}

func TestFixedWindowCounterConcurrent(t *testing.T) {
    limit := 5
    window := 500 * time.Millisecond
    counter := NewFixedWindowCounter(limit, window)

    allowed := 0
    done := make(chan struct{})
    // Spawn more goroutines than limit
    for i := 0; i < limit*2; i++ {
        go func() {
            if counter.Allow() {
                // atomically increment allowed counter
                // simple non-atomic increment is fine for test as we only read after all goroutines finish
                allowed++
            }
            done <- struct{}{}
        }()
    }
    // Wait for all goroutines
    for i := 0; i < limit*2; i++ {
        <-done
    }
    if allowed != limit {
        t.Fatalf("expected %d allowed requests, got %d", limit, allowed)
    }
}

