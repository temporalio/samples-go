// s3-mock runs an in-memory S3-compatible HTTP server backed by gofakes3, so
// this sample can run locally without an AWS account or Docker.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	externalstorage "github.com/temporalio/samples-go/external-storage"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

// statusRecorder captures the response status code so the logging middleware
// can include it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// newLoggingHTTPHandler prints a one-line summary of each request: method,
// path, response status, and how long the handler took.
func newLoggingHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s status=%d duration=%s",
			r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// newHandler builds an in-memory S3 handler with a single bucket pre-created.
// Callers (main, tests) typically wrap the result with newLoggingHTTPHandler.
func newHandler(bucket string) (http.Handler, error) {
	backend := s3mem.New()
	if err := backend.CreateBucket(bucket); err != nil {
		return nil, err
	}
	return gofakes3.New(backend).Server(), nil
}

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "Port to listen on")
	flag.Parse()

	handler, err := newHandler(externalstorage.S3Bucket)
	if err != nil {
		log.Fatalf("new handler: %v", err)
	}

	srv := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(port),
		Handler: newLoggingHTTPHandler(handler),
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	fmt.Printf("Mock S3 server running at http://%s\n", srv.Addr)
	fmt.Printf("Bucket %q created. Press ctrl+c to stop.\n", externalstorage.S3Bucket)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	select {
	case <-sigCh:
		_ = srv.Close()
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}
}
