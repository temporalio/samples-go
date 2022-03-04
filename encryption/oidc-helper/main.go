package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

type ProviderConfig struct {
	Issuer                      string `json:"issuer"`
	JWKS_URI                    string `json:"jwks_uri,omitempty"`
	AuthorizationEndpoint       string `json:"authorization_endpoint,omitempty"`
	callbackEndpoint            string
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitempty"`
	TokenEndpoint               string `json:"token_endpoint,omitempty"`
	ClientID                    string `json:"client_id,omitempty"`
	clientSecret                string
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

type appConfig struct {
	Provider     string
	ProxyURL     string
	ClientID     string
	ClientSecret string
}

func packState(state string, redirect string) string {
	return fmt.Sprintf("%s:%s", state, redirect)
}

func unpackState(state string) (string, string) {
	parts := strings.SplitN(state, ":", 2)
	return parts[0], parts[1]
}

func newOIDCHelperHTTPHandler(providerConfig *ProviderConfig, proxy string) http.Handler {
	mux := http.NewServeMux()

	cfg := *providerConfig
	if proxy != "" {
		base := strings.TrimSuffix(proxy, "/")
		cfg.AuthorizationEndpoint = base + "/oauth/authorize"
		cfg.TokenEndpoint = base + "/oauth/token"
		cfg.callbackEndpoint = base + "/oauth/callback"
	}

	mux.Handle("/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(cfg)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}))

	mux.Handle("/oauth/authorize", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		redirect := query.Get("redirect_uri")
		state := query.Get("state")
		query.Set("state", packState(state, redirect))
		query.Set("redirect_uri", cfg.callbackEndpoint)

		http.Redirect(w, r, fmt.Sprintf("%s?%s", providerConfig.AuthorizationEndpoint, query.Encode()), http.StatusFound)
	}))

	mux.Handle("/oauth/token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		values := r.Form
		values.Set("client_secret", cfg.clientSecret)
		values.Set("grant_type", "authorization_code")
		values.Set("redirect_uri", cfg.callbackEndpoint)

		res, err := http.PostForm(providerConfig.TokenEndpoint, values)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", res.Header.Get("Content-Length"))

		io.Copy(w, res.Body)
		res.Body.Close()
	}))

	mux.Handle("/oauth/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		state := query.Get("state")
		state, redirect := unpackState(state)
		query.Set("state", state)

		http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect, query.Encode()), http.StatusFound)
	}))

	return mux
}

func main() {
	logger := log.NewCLILogger()
	config := &appConfig{}

	flag.StringVar(&config.Provider, "provider", "", "OIDC Provider URL")
	flag.StringVar(&config.ClientID, "client-id", "", "OIDC Client ID. Optional. Allows clients to discover the client ID")
	flag.StringVar(&config.ProxyURL, "proxy-url", "", "Base URL to use for proxy endpoints, enables proxying for Authorization flow")
	flag.StringVar(&config.ClientSecret, "client-secret", "", "OIDC Client Secret. Optional. Required for proxying Authorization flow")
	flag.Parse()

	if config.Provider == "" {
		logger.Fatal("provider flag is required")
	}

	providerConfig, err := discoverProviderConfig(config.Provider)
	if err != nil {
		logger.Fatal("error", tag.NewErrorTag(err))
	}
	if config.ClientID != "" {
		providerConfig.ClientID = config.ClientID
	}
	if config.ClientSecret != "" {
		providerConfig.clientSecret = config.ClientSecret
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:8083",
		Handler: newOIDCHelperHTTPHandler(providerConfig, config.ProxyURL),
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
