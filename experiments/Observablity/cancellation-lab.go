package observablity

import (
	"log"
	"net/http"
	"time"
)

func Observe() {
	mux := http.NewServeMux()

	mux.HandleFunc("/work", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		log.Println("Request started")

		// Simulate a slow downstream operation.
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			log.Println("Work completed")

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("completed\n"))

		case <-r.Context().Done():
			log.Printf(
				"Request canceled after %s: %v",
				time.Since(start),
				r.Context().Err(),
			)

			// The client may already have disconnected.
			// Do not attempt to write a response here.
			return
		}
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Listening on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Referencing context here is unnecessary for the handler itself;
// this declaration is intentionally omitted from the implementation.
