package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log/tag"
)

type Provider struct {
	Issuer   string `json:"issuer"`
	JWKSURI  string `json:"jwks_uri,omitempty"`
	Audience string
	mapper   authorization.ClaimMapper
}

func (p *Provider) Authorize(namespace string, r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		logger.Warn("Authorization header not set")
		return false
	}

	authInfo := authorization.AuthInfo{
		AuthToken: token,
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
