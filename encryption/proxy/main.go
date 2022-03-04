package main

import (
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strings"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.temporal.io/server/common/log"

	"github.com/temporalio/samples-go/encryption"
)

const ServerEndpoint = "localhost:7234"

type ProviderConfig struct {
	Issuer   string `json:"issuer"`
	JWKS_URI string `json:"jwks_uri,omitempty"`
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

type appConfig struct {
	Provider string
	Audience string
}

type audienceGetter struct {
	audience string
}

func (a *audienceGetter) Audience(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) string {
	if a == nil {
		return ""
	}

	return a.audience
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

	flag.StringVar(&config.Provider, "provider", "", "OIDC Provider URL. Optional. Enable and require oauth")
	flag.StringVar(&config.Audience, "audience", "", "OIDC Audience. Optional. Only accept tokens for this audience")
	flag.Parse()

	clientInterceptor, err := converter.NewPayloadEncoderGRPCClientInterceptor(
		converter.PayloadEncoderGRPCClientInterceptorOptions{
			Encoders: encryption.NewEncoders(encryption.DataConverterOptions{Compress: true}),
		},
	)
	if err != nil {
		logger.Fatal("unable to create interceptor: %v", tag.NewErrorTag(err))
	}

	grpcClient, err := grpc.Dial(
		"localhost:7233",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientInterceptor),
	)
	defer func() { _ = grpcClient.Close() }()

	workflowClient := workflowservice.NewWorkflowServiceClient(grpcClient)

	if err != nil {
		logger.Fatal("unable to create client: %v", tag.NewErrorTag(err))
	}

	listener, err := net.Listen("tcp", ServerEndpoint)
	if err != nil {
		logger.Fatal("unable to create listener: %v", tag.NewErrorTag(err))
	}

	serverInterceptors := []grpc.UnaryServerInterceptor{}

	if config.Provider != "" {
		providerConfig, err := discoverProviderConfig(config.Provider)
		if err != nil {
			logger.Fatal("error", tag.NewErrorTag(err))
		}

		serverInterceptors = append(serverInterceptors,
			authorization.NewAuthorizationInterceptor(
				newClaimMapper(providerConfig.JWKS_URI, logger),
				authorization.NewDefaultAuthorizer(),
				metrics.NoopMetricsClient{},
				logger,
				&audienceGetter{audience: config.Audience},
			),
		)
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(serverInterceptors...))
	handler, err := client.NewWorkflowServiceProxyServer(
		client.WorkflowServiceProxyOptions{Client: workflowClient},
	)
	if err != nil {
		logger.Fatal("unable to create service proxy: %v", tag.NewErrorTag(err))
	}

	workflowservice.RegisterWorkflowServiceServer(server, handler)

	err = server.Serve(listener)
	if err != nil {
		logger.Fatal("unable to serve: %v", tag.NewErrorTag(err))
	}
}
