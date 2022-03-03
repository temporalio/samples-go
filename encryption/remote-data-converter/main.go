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
	Issuer                      string `json:"issuer"`
	JWKS_URI                    string `json:"jwks_uri,omitempty"`
	TokenEndpoint               string `json:"token_endpoint,omitempty"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitempty"`
	Audience                    string `json:"audience,omitempty"`
	ClientID                    string `json:"client_id,omitempty"`
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

// Simple HTTP + CORS handler for tctl and Temporal Web use, without oauth
// This remote decoder example supports URLs like: /{namespace}/encode and /{namespace}/decode
// For example, for the default namespace you would hit /default/encode and /default/decode
// tctl --remote_data_converter_endpoint flag replaces `{namespace}` in the endpoint URL you pass it
// with the namespace being operated on to easily support this pattern.
func newPayloadEncoderHTTPHandler(web string, encoders map[string][]converter.PayloadEncoder) http.Handler {
	mux := http.NewServeMux()

	for namespace, encoderChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		encoderHandler := converter.NewPayloadEncoderHTTPHandler(encoderChain...)

		mux.Handle("/"+namespace+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if web != "" {
				w.Header().Set("Access-Control-Allow-Origin", web)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace")

				if r.Method == "OPTIONS" {
					return
				}
			}

			encoderHandler.ServeHTTP(w, r)
		}))
	}

	return mux
}

// oauth + CORS handler for tctl and Temporal Web use with oauth
// This handler supports the same URLs as the simple handler above.
// It also provides an additional URL /.well-known/openid-configuration
// This can be used by tctl to discover the correct oidc configuration for
// the remote decoder rather than having to pass the provider, client_id and audience
// as options to tctl when logging in.
// It is not required for remote decoder implementations to support this endpoint, it is
// possible to pass all required values to tctl directly if preferred.
func newOauthPayloadEncoderHTTPHandler(providerConfig ProviderConfig, web string, encoders map[string][]converter.PayloadEncoder, logger log.Logger) http.Handler {
	mux := http.NewServeMux()

	mapper := newClaimMapper(providerConfig.JWKS_URI, logger)

	mux.Handle("/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(providerConfig)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}))

	for namespace, encoderChain := range encoders {
		fmt.Printf("Handling namespace: %s\n", namespace)

		encoderHandler := converter.NewPayloadEncoderHTTPHandler(encoderChain...)
		n := namespace

		mux.Handle("/"+namespace+"/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if web != "" {
				w.Header().Set("Access-Control-Allow-Origin", web)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace")

				if r.Method == "OPTIONS" {
					return
				}
			}

			authInfo := authorization.AuthInfo{
				AuthToken: r.Header.Get("Authorization"),
				Audience:  providerConfig.Audience,
			}

			claims, err := mapper.GetClaims(&authInfo)
			if err != nil {
				logger.Warn("unable to parse claims")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// If they have no role in this namespace they will get RoleUndefined
			role := claims.Namespaces[n]

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
				encoderHandler.ServeHTTP(w, r)
				return
			}

			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}))
	}

	return mux
}

type appConfig struct {
	Provider string
	Audience string
	Web      string
	ClientID string
}

func discoverProviderConfig(providerURL string) (ProviderConfig, error) {
	var providerConfig ProviderConfig

	res, err := http.Get(strings.TrimSuffix(providerURL, "/") + "/.well-known/openid-configuration")
	if err != nil {
		return providerConfig, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&providerConfig)
	if err != nil {
		return providerConfig, err
	}

	return providerConfig, nil
}

func main() {
	logger := log.NewCLILogger()
	config := &appConfig{}

	flag.StringVar(&config.Provider, "provider", "", "OIDC Provider URL (optional, enables and requires oauth)")
	flag.StringVar(&config.Audience, "audience", "", "OIDC Audience to enforce (optional)")
	flag.StringVar(&config.ClientID, "client_id", "", "OIDC Client ID to advertise to users (optional)")
	flag.StringVar(&config.Web, "web", "", "Temporal Web URL (optional)")
	flag.Parse()

	// When supporting multiple namespaces, add them here.
	// If you need to tailor encoding per namespace, you'd do that here.
	// Only handle encoding for the default namespace in this example.
	encoders := map[string][]converter.PayloadEncoder{
		"default": encryption.NewEncoders(encryption.DataConverterOptions{Compress: true}),
	}

	if config.Web != "" {
		fmt.Printf("CORS support is enabled for Origin: %s\n", config.Web)
	} else {
		fmt.Printf("CORS support is disabled\n")
	}

	var handler http.Handler

	if config.Provider != "" {
		fmt.Printf("oauth support is enabled using: %s\n", config.Provider)
		if config.Audience != "" {
			fmt.Printf("enforcing oauth audience: %s\n", config.Audience)
		}

		providerConfig, err := discoverProviderConfig(config.Provider)
		if err != nil {
			logger.Fatal("error", tag.NewErrorTag(err))
		}
		providerConfig.ClientID = config.ClientID
		providerConfig.Audience = config.Audience

		handler = newOauthPayloadEncoderHTTPHandler(providerConfig, config.Web, encoders, logger)
	} else {
		fmt.Printf("oauth support is disabled\n")

		handler = newPayloadEncoderHTTPHandler(config.Web, encoders)
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
