# Concurrency, Data Races, and Synchronization in Go

This document explains the concepts demonstrated in [`race1.go`](race1.go) and [`race1_test.go`](race1_test.go), covering **Data Races**, **Mutex Synchronization**, and **Atomic Operations**.

---

## 1. What is a Data Race?

In Go, a **data race** happens when:
1. Two or more goroutines access the same memory location simultaneously.
2. At least one access is a **write**.
3. There is **no synchronization** (no Mutex, channels, or atomic operations) establishing an ordering relationship (*happens-before*).

### The Problem in `mode=race`
In `race1.go`, the unsafe increment mode does this:
```go
for j := 0; j < *iterations; j++ {
    counter++ // Unsafe concurrent write!
}
```

Although `counter++` looks like a single line of Go code, at the CPU hardware level it consists of **three separate instructions**:
1. **Read**: Fetch current value of `counter` from RAM into CPU register.
2. **Modify**: Add `1` to the register value.
3. **Write**: Store register value back to RAM.

When multiple goroutines run concurrently on different CPU cores, these steps interleave:
- Goroutine A reads `100`.
- Goroutine B reads `100`.
- Goroutine A writes `101`.
- Goroutine B writes `101` (overwriting Goroutine A's work).

**Result:** Increments are silently lost. If 8 workers do 100,000 increments, instead of `800,000`, the final counter will be far lower.

---

## 2. Solution 1: Mutual Exclusion (`sync.Mutex`)

A **Mutex** (mutual exclusion lock) guarantees that only **one goroutine at a time** can execute the code enclosed between `.Lock()` and `.Unlock()`.

```go
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
```

### Why locking both Read and Write matters:
- **Writing (`Increment`)**: Ensures no two goroutines enter the critical section concurrently.
- **Reading (`Value`)**: Reading without acquiring the lock while another goroutine is writing would still be a data race (and could yield torn or stale reads on 32-bit architectures).

### Go Memory Model (*Happens-Before*):
- An `Unlock()` on a mutex **synchronizes-before** any subsequent `Lock()` on that same mutex returns.
- This guarantees all writes made before the unlock are visible to the next goroutine acquiring the lock.

---

## 3. Solution 2: Atomic Operations (`sync/atomic`)

Go 1.19+ provides typed atomics like `atomic.Int64`:

```go
var counter atomic.Int64

for j := 0; j < *iterations; j++ {
    counter.Add(1)
}
...
fmt.Println("Counter: ", counter.Load())
```

### How Atomics Work:
- Atomic operations use low-level CPU hardware instructions (such as `LOCK XADD` on x86/x64).
- The CPU guarantees that reading, modifying, and writing back to memory happens as an **indivisible single step** at the hardware bus level.
- **No lock or OS-level scheduler blocking** is involved.

### Mutex vs. Atomic: When to use which?

| Feature | `sync.Mutex` | `sync/atomic` (`atomic.Int64`) |
| :--- | :--- | :--- |
| **Granularity** | Protects complex state, multiple fields, slices, maps, or blocks of logic. | Protects single primitive variables (integers, booleans, pointers). |
| **Performance** | Incurs slight lock contention and context-switch overhead under high contention. | Extremely fast; direct CPU hardware instructions. |
| **Complexity** | Simple mental model for multi-variable state. | Can get complex if managing invariants across multiple variables. |

---

## 4. Testing Concurrent Code (`race1_test.go`)

In [`race1_test.go`](race1_test.go), we verify the correctness of the `Counter` using a concurrent stress test:

```go
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
```

### Key Takeaways from the Test:
1. **`sync.WaitGroup` coordinates completion**:
   - `wg.Add(workers)` registers active goroutines.
   - `defer wg.Done()` decrements counter when the worker finishes.
   - `wg.Wait()` blocks until all goroutines report completion.
2. **Determinism Verification**:
   - Total expected increments = $\text{workers} \times \text{iterations} = 10 \times 10,000 = 100,000$.
   - If `got != want`, synchronization has failed.

---

## 5. How to Run & Detect Races

### Run the program with different modes:
```bash
# Run unsafe race mode (output will be inaccurate)
go run main.go -mode=race

# Run with Mutex protection (exact count)
go run main.go -mode=mutex

# Run with Atomic operations (exact count, high throughput)
go run main.go -mode=atomic
```

### Run with the Go Race Detector (`-race` flag):
Go includes a built-in ThreadSanitizer-powered race detector:

```bash
# Detect races in tests
go test -v -race ./experiments/go_concept_exp/race/...

# Detect races during application execution
go run -race main.go -mode=race
```

When run against unsafe code, the race detector outputs a detailed traceback showing the conflicting goroutines, exact source lines, and memory addresses where the race occurred.

