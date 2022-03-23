package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	remotecodec "github.com/temporalio/samples-go/remote-codec"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

var logger log.Logger

// newCORSHTTPHandler wraps a HTTP handler with CORS support
func newCORSHTTPHandler(web string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", web)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// newPayloadEncoderOauthHTTPHandler wraps a HTTP handler with oauth support
func newPayloadEncoderOauthHTTPHandler(provider *Provider, namespace string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if provider.Authorize(namespace, r) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

// HTTP handler for codecs.
// This remote codec server example supports URLs like: /{namespace}/encode and /{namespace}/decode
// For example, for the default namespace you would hit /default/encode and /default/decode
// It will also accept URLs: /encode and /decode with the X-Namespace set to indicate the namespace.
func newPayloadCodecNamespacesHTTPHandler(encoders map[string][]converter.PayloadCodec, provider *Provider) http.Handler {
	mux := http.NewServeMux()

	codecHandlers := make(map[string]http.Handler, len(encoders))
	for namespace, codecChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		handler := converter.NewPayloadCodecHTTPHandler(codecChain...)
		if provider != nil {
			handler = newPayloadEncoderOauthHTTPHandler(provider, namespace, handler)
		}
		mux.Handle("/"+namespace+"/", handler)

		codecHandlers[namespace] = handler
	}

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.Header.Get("X-Namespace")
		if namespace != "" {
			if handler, ok := codecHandlers[namespace]; ok {
				handler.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}))

	return mux
}

var portFlag int
var providerFlag string
var audienceFlag string
var webFlag string

var provider *Provider

func init() {
	logger = log.NewCLILogger()

	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
	flag.StringVar(&providerFlag, "provider", "", "OIDC Provider URL. Optional: Enforces oauth authentication")
	flag.StringVar(&audienceFlag, "audience", "", "OIDC Audience. Optional")
	flag.StringVar(&webFlag, "web", "", "Temporal Web URL. Optional: enables CORS which is required for access from Temporal Web")
}

func main() {
	flag.Parse()

	// Set codecs per namespace here.
	// Only handle codecs for the default namespace in this example.
	codecs := map[string][]converter.PayloadCodec{
		"default": {remotecodec.NewPayloadCodec()},
	}

	if providerFlag != "" {
		p, err := newProvider(providerFlag)
		if err != nil {
			logger.Fatal("error", tag.NewErrorTag(err))
		}
		fmt.Printf("oauth enabled for: %s\n", provider.Issuer)
		provider = p
		if audienceFlag != "" {
			fmt.Printf("oauth audience: %s\n", provider.Audience)
			provider.Audience = audienceFlag
		}
	}

	handler := newPayloadCodecNamespacesHTTPHandler(codecs, provider)

	if webFlag != "" {
		fmt.Printf("CORS enabled for Origin: %s\n", webFlag)
		handler = newCORSHTTPHandler(webFlag, handler)
	}

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
		logger.Fatal("error", tag.NewErrorTag(err))
	}
}
