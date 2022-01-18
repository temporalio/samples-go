package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gopkg.in/square/go-jose.v2"

	"github.com/temporalio/samples-go/serverjwtauth"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) == 2 {
		if os.Args[1] == "gen" {
			return serverjwtauth.GenAndWriteKey()
		} else if os.Args[1] == "serve" {
			return serve()
		} else if os.Args[1] == "gen-and-serve" {
			err := serverjwtauth.GenAndWriteKey()
			if err == nil {
				err = serve()
			}
			return err
		} else if os.Args[1] == "tctl-system-token" {
			return tctlSystemToken()
		}
	}
	return fmt.Errorf("only a single argument of 'gen', 'serve', or 'gen-and-serve' is supported")
}

func serve() error {
	log.Print("Starting JWKS server")

	// Load key
	_, jwk, err := serverjwtauth.ReadKey()
	if err != nil {
		return err
	}

	// Marshal set JSON
	jsonBytes, err := json.Marshal(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{*jwk}})
	if err != nil {
		return fmt.Errorf("failed marshalling jwks: %w", err)
	}

	// Serve
	http.HandleFunc("/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonBytes)
	})
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed listening: %w", err)
	}
	var srv http.Server
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(l) }()

	// Wait for error or signal
	log.Printf("Started JWKS server. Endpoint: http://%v/jwks.json. Ctrl+C to exit.", l.Addr())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	select {
	case <-sigCh:
		_ = srv.Close()
		return nil
	case <-errCh:
		return fmt.Errorf("server failed: %w", err)
	}
}

func tctlSystemToken() error {
	// Load key
	key, jwk, err := serverjwtauth.ReadKey()
	if err != nil {
		return err
	}

	// Create token that will last an hour
	config := serverjwtauth.JWTConfig{
		Key:         key,
		KeyID:       jwk.KeyID,
		Permissions: []string{"system:admin"},
		Expiration:  1 * time.Hour,
	}
	token, err := config.GenToken()
	if err != nil {
		return err
	}
	fmt.Printf("TEMPORAL_CLI_AUTHORIZATION_TOKEN=Bearer %v", token)
	return nil
}
