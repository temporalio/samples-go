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

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "Port to listen on")
	flag.Parse()

	backend := s3mem.New()
	if err := backend.CreateBucket(externalstorage.S3Bucket); err != nil {
		log.Fatalf("create bucket: %v", err)
	}
	faker := gofakes3.New(backend)

	srv := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(port),
		Handler: newLoggingHTTPHandler(faker.Server()),
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
