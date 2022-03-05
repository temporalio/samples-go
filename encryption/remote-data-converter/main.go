package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/temporalio/samples-go/encryption"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

var logger log.Logger

type Provider struct {
	Issuer   string `json:"issuer"`
	JWKS_URI string `json:"jwks_uri,omitempty"`
	Audience string
	mapper   authorization.ClaimMapper
}

func (p *Provider) Authorize(namespace string, r *http.Request) bool {
	authInfo := authorization.AuthInfo{
		AuthToken: r.Header.Get("Authorization"),
		Audience:  p.Audience,
	}

	claims, err := p.mapper.GetClaims(&authInfo)
	if err != nil {
		logger.Warn("unable to parse claims", tag.NewErrorTag(err))
		return false
	}

	// If they have no role in this namespace they will get RoleUndefined
	role := claims.Namespaces[namespace]

	switch {
	case strings.HasSuffix(r.URL.Path, "/decode"):
		if role >= authorization.RoleReader {
			return true
		}
	case strings.HasSuffix(r.URL.Path, "/encode"):
		if role >= authorization.RoleWriter {
			return true
		}
	}

	return false
}

func newProvider(providerURL string) (*Provider, error) {
	var provider Provider

	res, err := http.Get(strings.TrimSuffix(providerURL, "/") + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&provider)
	if err != nil {
		return nil, err
	}

	provider.mapper = newClaimMapper(provider.JWKS_URI)

	return &provider, nil
}

func newClaimMapper(providerKeysURL string) authorization.ClaimMapper {
	authConfig := config.Authorization{
		JWTKeyProvider: config.JWTKeyProvider{
			KeySourceURIs: []string{providerKeysURL},
		},
		ClaimMapper: "default",
	}

	provider := authorization.NewDefaultTokenKeyProvider(
		&authConfig,
		logger,
	)

	return authorization.NewDefaultJWTClaimMapper(provider, &authConfig, logger)
}

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
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

// HTTP handler for tctl and Temporal Web use
// If a providerConfig is passed then oauth will be required to access the decoder.
// This remote decoder example supports URLs like: /{namespace}/encode and /{namespace}/decode
// For example, for the default namespace you would hit /default/encode and /default/decode
// tctl --remote_data_converter_endpoint flag replaces `{namespace}` in the endpoint URL you pass it
// with the namespace being operated on to easily support this pattern.
// It will also accept URLs: /encode and /decode with the X-Namespace set to indicate the namespace.
// This is useful for web access such as from Temporal Web.
func newPayloadEncoderNamespacesHTTPHandler(encoders map[string][]converter.PayloadEncoder, providers []*Provider) http.Handler {
	mux := http.NewServeMux()

	encoderHandlers := map[string]http.Handler{}
	for namespace, encoderChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		handler := converter.NewPayloadEncoderHTTPHandler(encoderChain...)
		if len(providers) > 0 {
			handler = newPayloadEncoderOauthHTTPHandler(providers, namespace, handler)
		}
		mux.Handle("/"+namespace+"/", handler)

		encoderHandlers[namespace] = handler
	}

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.Header.Get("X-Namespace")
		if namespace != "" {
			if handler, ok := encoderHandlers[namespace]; ok {
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

var providersFlag providers
var audienceFlag audience
var webFlag string

func init() {
	logger = log.NewCLILogger()

	flag.Var(&providersFlag, "provider", "OIDC Provider URL. Optional: Enables and requires oauth")
	flag.Var(&audienceFlag, "audience", "OIDC Audience to enforce for the previous provider. Optional")
	flag.StringVar(&webFlag, "web", "", "Temporal Web URL. Optional: enables CORS which is required for access from Temporal Web")
}

func main() {
	flag.Parse()

	// When supporting multiple namespaces, add them here.
	// If you need to tailor encoding per namespace, you'd do that here.
	// Only handle encoding for the default namespace in this example.
	encoders := map[string][]converter.PayloadEncoder{
		"default": encryption.NewEncoders(encryption.DataConverterOptions{Compress: true}),
	}

	for _, provider := range providersFlag {
		fmt.Printf("oauth enabled for: %s\n", provider.Issuer)
		if provider.Audience != "" {
			fmt.Printf("oauth audience: %s\n", provider.Audience)
		}
	}

	handler := newPayloadEncoderNamespacesHTTPHandler(encoders, providersFlag)

	if webFlag != "" {
		fmt.Printf("CORS enabled for Origin: %s\n", webFlag)
		handler = newCORSHTTPHandler(webFlag, handler)
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:8081",
		Handler: handler,
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
