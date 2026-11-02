package goconceptexp

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func RatelimitHttpExp() {
	// Average rate: 5 requests per second, burst: 10 requests.
	limiter := rate.NewLimiter(5, 10)

	// Concurrency limiter: max 3 concurrent operations.
	const maxConcurrency = 3
	concurrencySem := make(chan struct{}, maxConcurrency)

	// Metrics tracked for the UI
	var (
		mu                  sync.Mutex
		activeWorkers       int
		successCount        int
		rateRejectedCount   int
		concurrencyRejected int
	)

	mux := http.NewServeMux()

	// 1. Serve the HTML page at "/"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	})

	// 2. Metrics endpoint for the UI
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		stats := map[string]int{
			"active":              activeWorkers,
			"success":             successCount,
			"rateRejected":        rateRejectedCount,
			"concurrencyRejected": concurrencyRejected,
		}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	})

	// 3. Work endpoint used by the HTML UI
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode") // "both", "rate", "concurrency"

		// Rate limiting check
		if mode == "both" || mode == "rate" {
			if !limiter.Allow() {
				mu.Lock()
				rateRejectedCount++
				mu.Unlock()
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
		}

		// Concurrency limiting check
		if mode == "both" || mode == "concurrency" {
			select {
			case concurrencySem <- struct{}{}:
				defer func() { <-concurrencySem }()
			default:
				mu.Lock()
				concurrencyRejected++
				mu.Unlock()
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		// Track active requests
		mu.Lock()
		activeWorkers++
		mu.Unlock()

		defer func() {
			mu.Lock()
			activeWorkers--
			successCount++
			mu.Unlock()
		}()

		// Simulate 2 seconds of work as described in the UI
		select {
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case <-r.Context().Done():
			return
		}
	})

	// 4. Endpoint for curl testing
	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate_limit_exceeded",
				"message": "Please retry later.",
			})
			return
		}

		select {
		case <-time.After(100 * time.Millisecond):
		case <-r.Context().Done():
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "accepted",
		})
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

// test run  :
// for bulk run
//  seq 1 30 | xargs -P 30 -I {} \
//   curl -s -o /dev/null -w '%{http_code}\n' \
//   http://localhost:8080/generate

// for single run
// curl -i http://localhost:8080/generate

// The exact distribution depends on request timing and token refill.

// With an initially full bucket of 10 tokens, no more than 10 requests can consume
//
//	the initial tokens. Additional requests may succeed as tokens refill during the test.
const page = `<!doctype html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Rate vs Concurrency Lab</title>
<style>
body { font: 16px system-ui; max-width: 760px; margin: 32px auto; padding: 0 16px; }
button, select { padding: 10px; margin: 4px; }
.cards { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.card { padding: 16px; background: #eee; border-radius: 10px; }
.num { font-size: 28px; font-weight: bold; }
#log { white-space: pre-wrap; overflow-wrap: anywhere; }
</style>
</head>
<body>
<h1>Rate Limiting vs Concurrency</h1>
<p>Each accepted request takes 2 seconds to finish.</p>
<label>Test mode:
<select id="mode">
<option value="both">Both limits</option>
<option value="rate">Rate limit only</option>
<option value="concurrency">Concurrency limit only</option>
</select></label>
<button onclick="runTest()">Send 20 requests</button>
<button onclick="refresh()">Refresh metrics</button>
<div class="cards">
<div class="card">Active work<div class="num" id="active">0</div>
<progress id="bar" max="3" value="0"></progress></div>
<div class="card">Successful<div class="num" id="success">0</div></div>
<div class="card">Rate rejected (429)<div class="num" id="rateRejected">0</div></div>
<div class="card">Concurrency rejected (503)<div class="num" id="concurrencyRejected">0</div></div>
</div>
<h3>Last test results</h3>
<pre id="log">Choose a mode and send requests.</pre>
<script>
async function refresh() {
  const s = await fetch('/stats').then(r => r.json());
  for (const k of ['active','success','rateRejected','concurrencyRejected'])
    document.getElementById(k).textContent = s[k];
  document.getElementById('bar').value = s.active;
}
async function runTest() {
  const mode = document.getElementById('mode').value;
  document.getElementById('log').textContent = 'Sending requests...';
  const results = await Promise.all(Array.from({length:20}, async () => {
    const r = await fetch('/work?mode=' + mode);
    return r.status;
  }));
  const counts = {};
  results.forEach(code => counts[code] = (counts[code] || 0) + 1);
  document.getElementById('log').textContent =
    'HTTP status counts: ' + JSON.stringify(counts, null, 2);
  await refresh();
}
setInterval(refresh, 300);
refresh();
</script>
</body>
</html>`
