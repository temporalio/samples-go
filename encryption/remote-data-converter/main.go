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

type ProviderConfig struct {
	Issuer   string `json:"issuer"`
	JWKS_URI string `json:"jwks_uri,omitempty"`
	Audience string
	mapper   authorization.ClaimMapper
}

func discoverProviderConfig(providerURL string) (*ProviderConfig, error) {
	var providerConfig ProviderConfig

	res, err := http.Get(strings.TrimSuffix(providerURL, "/") + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&providerConfig)
	if err != nil {
		return nil, err
	}

	return &providerConfig, nil
}

func newClaimMapper(providerKeysURL string, logger log.Logger) authorization.ClaimMapper {
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
func newPayloadEncoderOauthHTTPHandler(providerConfig *ProviderConfig, namespace string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authInfo := authorization.AuthInfo{
			AuthToken: r.Header.Get("Authorization"),
			Audience:  providerConfig.Audience,
		}

		claims, err := providerConfig.mapper.GetClaims(&authInfo)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// If they have no role in this namespace they will get RoleUndefined
		role := claims.Namespaces[namespace]

		authorized := false

		switch {
		case strings.HasSuffix(r.URL.Path, "/decode"):
			if role >= authorization.RoleReader {
				authorized = true
			}
		case strings.HasSuffix(r.URL.Path, "/encode"):
			if role >= authorization.RoleWriter {
				authorized = true
			}
		}

		if authorized {
			next.ServeHTTP(w, r)
			return
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
func newPayloadEncoderNamespacesHTTPHandler(encoders map[string][]converter.PayloadEncoder, providerConfig *ProviderConfig, logger log.Logger) http.Handler {
	mux := http.NewServeMux()

	encoderHandlers := map[string]http.Handler{}
	for namespace, encoderChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		handler := converter.NewPayloadEncoderHTTPHandler(encoderChain...)
		if providerConfig != nil {
			handler = newPayloadEncoderOauthHTTPHandler(providerConfig, namespace, handler)
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

type appConfig struct {
	Provider string
	Audience string
	Web      string
}

func main() {
	logger := log.NewCLILogger()
	config := &appConfig{}

	flag.StringVar(&config.Provider, "provider", "", "OIDC Provider URL. Optional: Enables and requires oauth")
	flag.StringVar(&config.Audience, "audience", "", "OIDC Audience to enforce. Optional")
	flag.StringVar(&config.Web, "web", "", "Temporal Web URL. Optional: enables CORS which is required for access from Temporal Web")
	flag.Parse()

	// When supporting multiple namespaces, add them here.
	// If you need to tailor encoding per namespace, you'd do that here.
	// Only handle encoding for the default namespace in this example.
	encoders := map[string][]converter.PayloadEncoder{
		"default": encryption.NewEncoders(encryption.DataConverterOptions{Compress: true}),
	}

	var providerConfig *ProviderConfig

	if config.Provider != "" {
		fmt.Printf("oauth support is enabled using: %s\n", config.Provider)
		if config.Audience != "" {
			fmt.Printf("enforcing oauth audience: %s\n", config.Audience)
		}

		var err error
		providerConfig, err = discoverProviderConfig(config.Provider)
		if err != nil {
			logger.Fatal("unable to discover provider config", tag.NewErrorTag(err))
		}
		providerConfig.mapper = newClaimMapper(providerConfig.JWKS_URI, logger)
	} else {
		fmt.Printf("oauth support is disabled\n")
	}

	handler := newPayloadEncoderNamespacesHTTPHandler(encoders, providerConfig, logger)

	if config.Web != "" {
		fmt.Printf("CORS support is enabled for Origin: %s\n", config.Web)
		handler = newCORSHTTPHandler(config.Web, handler)
	} else {
		fmt.Printf("CORS support is disabled\n")
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
