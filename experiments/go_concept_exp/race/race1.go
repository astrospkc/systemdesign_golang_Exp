package goconceptexp

import (
	"flag"
	"fmt"
	"sync"
	"sync/atomic"
)

type Counter struct {
	mu    sync.Mutex
	count int64
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func Race() {
	mode := flag.String("mode", "mutex", "race , mutex , or atomic")
	workers := flag.Int("workers", 8, "number of goroutines")
	iterations := flag.Int("iterations", 100000, "increments per goroutine")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(*workers)

	switch *mode {
	case "race":
		var counter int64

		for i := 0; i < *workers; i++ {
			go func() {
				defer wg.Done()

				for j := 0; j < *iterations; j++ {
					counter++
				}
			}()

		}
		wg.Wait()
		fmt.Println("Counter: ", counter)

	case "mutex":
		var counter Counter

		for i := 0; i < *workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < *iterations; j++ {
					counter.Increment()
				}
			}()
		}
		wg.Wait()
		fmt.Println("Counter: ", counter.Value())

	case "atomic":
		var counter atomic.Int64

		for i := 0; i < *workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < *iterations; j++ {
					counter.Add(1)
				}
			}()
		}
		wg.Wait()
		fmt.Println("Counter: ", counter.Load())

	default:
		panic("unknown mode")
	}
}
