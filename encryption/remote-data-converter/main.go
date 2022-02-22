package main

import (
	"net/http"
	"os"
	"os/signal"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"

	"github.com/temporalio/samples-go/encryption"
)

func main() {
	logger := log.NewCLILogger()
	var frontendURL string

	if len(os.Args) > 1 {
		// To allow us to use this decoder for Temporal Web we need it's URL.
		// This allows us to return CORS headers permitting the browser to talk to the decoder.
		frontendURL = os.Args[1]
	}

	endpoint := converter.NewPayloadEncoderHTTPHandler(
		encryption.NewEncoders(encryption.DataConverterOptions{Compress: true})...,
	)

	srv := &http.Server{
		Addr: "0.0.0.0:8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Optional CORS support for Temporal Web
			if frontendURL != "" {
				// Add CORS headers before we hand over to PayloadEncoderHTTPHandler
				w.Header().Set("Access-Control-Allow-Origin", frontendURL)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")

				// If this is just a CORS preflight request we don't need to peform any work.
				// The browser just wants a 200 OK response with CORS header(s) set.
				if r.Method == "OPTIONS" {
					return
				}
			}

			// Delegate to PayloadEncoderHTTPHandler to perform the work.
			endpoint.ServeHTTP(w, r)
		}),
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		srv.Close()
	case err := <-errCh:
		logger.Fatal("error", tag.NewErrorTag(err))
	}
}
