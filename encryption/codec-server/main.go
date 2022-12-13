package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/temporalio/samples-go/encryption"

	"go.temporal.io/sdk/converter"
)

var portFlag int

func init() {
	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
}

func main() {
	flag.Parse()

	// This example codec server does not support varying config per namespace,
	// decoding for the Temporal Web UI or oauth.
	// For a more complete example of a codec server please see the codec-server sample at:
	// ../../codec-server.
	handler := converter.NewPayloadCodecHTTPHandler(&encryption.Codec{}, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))

	srv := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(portFlag),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		_ = srv.Close()
	case err := <-errCh:
		log.Fatal(err)
	}
}
