package observablity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func Trace() {
	propagateID := os.Getenv("PROPAGATE_ID") != "false"

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	metadataMux := http.NewServeMux()

	metadataMux.HandleFunc("/metadata", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "missing"
		}

		log.Printf(
			"service=metadata event=start request_id=%s",
			requestID,
		)

		// Simulate metadata processing.
		select {
		case <-time.After(100 * time.Millisecond):
		case <-r.Context().Done():
			log.Printf(
				"service=metadata event=canceled request_id=%s",
				requestID,
			)
			return
		}

		log.Printf(
			"service=metadata event=complete request_id=%s duration_ms=%d",
			requestID,
			time.Since(start).Milliseconds(),
		)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"metadata":"video metadata"}`)
	})

	go func() {
		server := &http.Server{
			Addr:              ":8081",
			Handler:           metadataMux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		log.Println("Metadata service listening on :8081")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("metadata server error: %v", err)
		}
	}()

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/videos", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		requestID := newRequestID()
		start := time.Now()

		w.Header().Set("X-Request-ID", requestID)

		log.Printf(
			"service=api event=start request_id=%s",
			requestID,
		)

		ctx, cancel := context.WithTimeout(
			r.Context(),
			1500*time.Millisecond,
		)
		defer cancel()

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://localhost:8081/metadata",
			nil,
		)
		if err != nil {
			http.Error(
				w,
				"Unable to create request",
				http.StatusInternalServerError,
			)
			return
		}

		if propagateID {
			req.Header.Set("X-Request-ID", requestID)
		}

		dependencyStart := time.Now()

		resp, err := client.Do(req)
		if err != nil {
			log.Printf(
				"service=api event=dependency_error request_id=%s error=%q",
				requestID,
				err.Error(),
			)

			http.Error(
				w,
				"Metadata service unavailable",
				http.StatusBadGateway,
			)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(
				w,
				"Unable to read metadata response",
				http.StatusBadGateway,
			)
			return
		}

		log.Printf(
			"service=api event=dependency_complete request_id=%s status=%d duration_ms=%d",
			requestID,
			resp.StatusCode,
			time.Since(dependencyStart).Milliseconds(),
		)

		if resp.StatusCode != http.StatusOK {
			http.Error(
				w,
				"Metadata service returned an error",
				http.StatusBadGateway,
			)
			return
		}

		log.Printf(
			"service=api event=complete request_id=%s duration_ms=%d",
			requestID,
			time.Since(start).Milliseconds(),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	apiServer := &http.Server{
		Addr:              ":8080",
		Handler:           apiMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on :8080; propagation=%v", propagateID)

	log.Fatal(apiServer.ListenAndServe())
}

func newRequestID() string {
	var value [16]byte

	if _, err := rand.Read(value[:]); err != nil {
		log.Printf("request ID generation failed: %v", err)
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(value[:])
}
