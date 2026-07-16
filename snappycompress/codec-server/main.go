package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/temporalio/samples-go/snappycompress"
	"go.temporal.io/sdk/converter"
)

var portFlag int

func init() {
	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
}

func main() {
	flag.Parse()

	handler := converter.NewPayloadCodecHTTPHandler(
		&snappycompress.Codec{Options: snappycompress.Options{AlwaysEncode: true}},
	)

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
