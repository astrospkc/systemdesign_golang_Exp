# Go Experiments & System Design Lab

A practical, hands-on repository dedicated to exploring core **Go (Golang) concepts**, concurrency internals, observability, and real-world **System Design patterns**.

---

## Motive & Focus Areas

The goal of this repository is to bridge the gap between theoretical software engineering principles and real-world Go implementations through runnable experiments and documentation:

- **Go Internals & Concurrency**: Practical exploration of goroutine scaling, data races, the Go memory model (`happens-before`), synchronization primitives (`sync.Mutex`, atomics, channels), and benchmark profiling.
- **System Design & Distributed Patterns**: Translating system design problems into functional Go code—including rate-limiting algorithms (Token Bucket, Fixed Window), concurrency throttling, circuit breakers, and connection handling.
- **Observability & Resilience**: Understanding production lifecycle management through request tracing, context propagation (`context.Context`), timeout handling, and cancellation ergonomics.
- **System Design Notes**: Clear explanations and architectural breakdowns of high-scale networking, data-intensive architectures, and distributed systems questions.

---

## Upcoming & Future Topics

As this repository evolves, upcoming areas of exploration will include:

- **Distributed Systems & Consensus**:
  - Implementations of Leader Election algorithms and consensus basics (Raft, Paxos concepts).
  - Consistent hashing and distributed partition rings for caching and storage.
- **Resilience & Fault Tolerance Patterns**:
  - Circuit Breakers, Retry policies with Exponential Backoff + Jitter, and Dead Letter Queues.
  - Connection pooling (TCP, database, gRPC) and graceful shutdown mechanics.
- **Advanced Rate Limiting & Traffic Shaping**:
  - Sliding Window Counter, Sliding Window Log, and Leaky Bucket implementations.
  - Distributed rate limiting using shared backends like Redis.
- **In-Memory Storage & Caching**:
  - Thread-safe LRU/LFU cache engines with custom eviction and TTL policies.
  - Write-through, write-behind, and cache-aside patterns.
- **Messaging & Event-Driven Architectures**:
  - Asynchronous event bus patterns in Go.
  - Architectural comparisons and design studies on Kafka, RabbitMQ, and message ordering guarantees.
- **Low-Level Go Performance & Profiling**:
  - Memory allocation optimizations, escape analysis, and zero-allocation techniques.
  - Deep dive into `pprof` (CPU, memory, goroutine blocking, mutex contention).
