package main

import (
	"flag"
	"fmt"
	bsdc "github.com/temporalio/samples-go/blob-store-data-converter"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	"go.temporal.io/sdk/converter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

var portFlag int
var web string

func init() {
	flag.IntVar(&portFlag, "port", 8082, "Port to listen on")
	flag.StringVar(&web, "web", "http://localhost:8233", "Temporal UI URL")
}

func main() {
	flag.Parse()

	// This example codec server does not support varying config per namespace,
	// decoding for the Temporal Web UI or oauth.
	// For a more complete example of a codec server please see the codec-server sample at:
	// https://github.com/temporalio/samples-go/tree/main/codec-server
	handler := converter.NewPayloadCodecHTTPHandler(
		//bsdc.NewBaseCodec(blobstore.NewClient()),
		bsdc.NewBlobCodec(blobstore.NewClient(), bsdc.PropagatedValues{}),
	)

	srv := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(portFlag),
		Handler: newCORSHTTPHandler(handler),
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("allowing CORS Headers for %s\n", web)
		fmt.Printf("Listening on http://%s/\n", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		_ = srv.Close()
	case err := <-errCh:
		log.Fatal(err)
	}
}

// newCORSHTTPHandler wraps a HTTP handler with CORS support
func newCORSHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", web)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace,X-CSRF-Token,Caller-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
