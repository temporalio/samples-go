package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"google.golang.org/grpc"
)

type Provider struct {
	Issuer   string `json:"issuer"`
	JWKSURI  string `json:"jwks_uri,omitempty"`
	audience string
	mapper   authorization.ClaimMapper
}

func (a *Provider) Audience(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) string {
	if a == nil {
		return ""
	}

	return a.audience
}

func newProvider(providerURL string) (*Provider, error) {
	var provider Provider

	res, err := http.Get(strings.TrimSuffix(providerURL, "/") + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	err = json.NewDecoder(res.Body).Decode(&provider)
	if err != nil {
		return nil, err
	}

	provider.mapper = newClaimMapper(provider.JWKSURI)

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
