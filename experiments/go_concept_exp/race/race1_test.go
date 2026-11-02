package goconceptexp

import (
	"sync"
	"testing"
)

func TestCounterConcurrent(t *testing.T) {
	var counter Counter
	const workers = 10
	const iterations = 10000

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				counter.Increment()
			}
		}()
	}

	wg.Wait()

	want := int64(workers * iterations)
	if got := counter.Value(); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}
