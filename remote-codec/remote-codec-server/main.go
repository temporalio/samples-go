package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

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
func newPayloadEncoderOauthHTTPHandler(providers []*Provider, namespace string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, provider := range providers {
			if provider.Authorize(namespace, r) {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

// HTTP handler for codecs.
// This remote codec server example supports URLs like: /{namespace}/encode and /{namespace}/decode
// For example, for the default namespace you would hit /default/encode and /default/decode
// It will also accept URLs: /encode and /decode with the X-Namespace set to indicate the namespace.
func newPayloadCodecNamespacesHTTPHandler(encoders map[string][]converter.PayloadCodec, providers []*Provider) http.Handler {
	mux := http.NewServeMux()

	codecHandlers := make(map[string]http.Handler, len(encoders))
	for namespace, codecChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		handler := converter.NewPayloadCodecHTTPHandler(codecChain...)
		if len(providers) > 0 {
			handler = newPayloadEncoderOauthHTTPHandler(providers, namespace, handler)
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

type providers []*Provider

func (p *providers) String() string {
	urls := []string{}
	for _, provider := range *p {
		urls = append(urls, provider.Issuer)
	}
	return strings.Join(urls, ", ")
}

func (p *providers) Set(value string) error {
	provider, err := newProvider(value)
	if err != nil {
		return err
	}
	*p = append(*p, provider)

	return nil
}

type audience string

func (a *audience) String() string {
	return fmt.Sprint(*a)
}

func (a *audience) Set(value string) error {
	if len(providersFlag) == 0 {
		return fmt.Errorf("the audience flag must come after a provider flag")
	}
	provider := providersFlag[len(providersFlag)-1]
	provider.Audience = value

	return nil
}

var portFlag int
var providersFlag providers
var audienceFlag audience
var webFlag string

func init() {
	logger = log.NewCLILogger()

	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
	flag.Var(&providersFlag, "provider", "OIDC Provider URL. Optional: Enforces oauth authentication. Repeat this flag for each OIDC provider you wish to use.")
	flag.Var(&audienceFlag, "audience", "OIDC Audience. Optional. Must follow a provider flag.")
	flag.StringVar(&webFlag, "web", "", "Temporal Web URL. Optional: enables CORS which is required for access from Temporal Web")
}

func main() {
	flag.Parse()

	// Set codecs per namespace here.
	// Only handle codecs for the default namespace in this example.
	codecs := map[string][]converter.PayloadCodec{
		"default": {remotecodec.NewPayloadCodec()},
	}

	for _, provider := range providersFlag {
		fmt.Printf("oauth enabled for: %s\n", provider.Issuer)
		if provider.Audience != "" {
			fmt.Printf("oauth audience: %s\n", provider.Audience)
		}
	}

	handler := newPayloadCodecNamespacesHTTPHandler(codecs, providersFlag)

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
