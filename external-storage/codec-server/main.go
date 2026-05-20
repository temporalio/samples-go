// codec-server hosts the payload HTTP handler from the Temporal Go SDK so the
// Web UI and CLI can transform external-storage payloads on demand.
//
// Deliberately left out for sample simplicity: authentication (slot a
// middleware between the CORS handler and the dispatcher) and configurable
// listen address. For an example of enabling authentication in a codec
// server, look at ../../codec-server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	externalstorage "github.com/temporalio/samples-go/external-storage"
	"go.temporal.io/sdk/converter"
)

const webUIOrigin = "http://localhost:8233"

// newPayloadNamespacesHTTPHandler returns an http.Handler that dispatches each
// request to a per-namespace handler chosen by the X-Namespace header. The
// Temporal Web UI and CLI send that header on every codec server request, so one
// process can host different codec/storage configurations per namespace
// without per-namespace URL prefixes.
func newPayloadNamespacesHTTPHandler(handlers map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.Header.Get("X-Namespace")
		h, ok := handlers[namespace]
		if !ok {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// statusRecorder captures the response status code so the logging middleware
// can include it. WriteHeader is only called once per request by the SDK's
// payload handler; subsequent writes go through ResponseWriter directly.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// newLoggingHTTPHandler prints a one-line summary of each request: method,
// path, namespace, response status, and how long the handler took.
func newLoggingHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s namespace=%q status=%d duration=%s",
			r.Method, r.URL.Path, r.Header.Get("X-Namespace"), rec.status, time.Since(start))
	})
}

// newCORSHTTPHandler lets the Temporal Web UI call the codec server from its own
// origin. The X-Namespace header is allowlisted so the dispatcher can read it.
func newCORSHTTPHandler(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Namespace")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "Port to listen on")
	flag.Parse()

	ctx := context.Background()
	driver, err := externalstorage.NewS3Driver(ctx)
	if err != nil {
		log.Fatalf("new s3 driver: %v", err)
	}

	// Build the payload handler with the same codec + external storage that
	// the worker and starter use. PreStorageCodecs runs before storage on
	// encode and after retrieval on decode, mirroring what the client-side
	// DataConverter does.
	defaultNamespaceHandler, err := converter.NewPayloadHTTPHandler(converter.PayloadHTTPHandlerOptions{
		PreStorageCodecs: []converter.PayloadCodec{
			converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}),
		},
		ExternalStorage: converter.ExternalStorage{
			Drivers: []converter.StorageDriver{driver},
		},
	})
	if err != nil {
		log.Fatalf("new payload handler: %v", err)
	}

	// Per-namespace map: extend this to host additional namespaces with their
	// own codec chain and/or storage backend.
	handler := newPayloadNamespacesHTTPHandler(map[string]http.Handler{
		"default": defaultNamespaceHandler,
	})
	handler = newCORSHTTPHandler(webUIOrigin, handler)
	handler = newLoggingHTTPHandler(handler)

	srv := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	fmt.Printf("Codec server running at http://%s, ctrl+c to exit\n", srv.Addr)

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
