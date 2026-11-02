package goconceptexp

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "net/http/pprof"
)

func expensiveWork(n int) [32]byte {
	var result [32]byte

	for i := 0; i < n; i++ {
		// Intentionally inefficient for this experiment:
		// format a new string and allocate a byte slice
		// on every iteration.
		input := []byte(fmt.Sprintf("document-chunk-%d", i))

		result = sha256.Sum256(input)
	}

	return result
}

func Performance() {
	// Keep the profiling HTTP server accessible only
	// through the local machine.
	go func() {
		log.Println(
			http.ListenAndServe("127.0.0.1:6060", nil),
		)
	}()

	api := http.NewServeMux()

	api.HandleFunc("GET /work", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		n := 100_000

		if value := r.URL.Query().Get("n"); value != "" {
			parsed, err := strconv.Atoi(value)
			if err != nil || parsed < 1 || parsed > 1_000_000 {
				http.Error(
					w,
					"n must be between 1 and 1000000",
					http.StatusBadRequest,
				)
				return
			}

			n = parsed
		}

		result := expensiveWork(n)

		fmt.Fprintf(
			w,
			"processed=%d checksum=%x\n",
			n,
			result,
		)
	})

	log.Println("API listening on :8080")

	if err := http.ListenAndServe(":8080", api); err != nil {
		log.Fatal(err)
	}
}
