package options

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"crypto/tls"
	"crypto/x509"

	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type CommonFlagsParser interface {
	ClientOptions() (client.Options, error)
}

type clientFlags struct {
	targetHost         *string
	namespace          *string
	serverRootCACert   *string
	clientCert         *string
	clientKey          *string
	serverName         *string
	insecureSkipVerify *bool
	apiKey             *string

	isParsed func() bool
}

// Utility to parse common flags via a supplied FlagSet.
// This allows users to plug in custom flags while using
// [CommonFlagsParser] to parse options from common flags
func NewClientFlagParser(set *flag.FlagSet) CommonFlagsParser {
	return &clientFlags{
		targetHost:         set.String("target-host", "localhost:7233", "Host:port for the Temporal service"),
		namespace:          set.String("namespace", "default", "Namespace to connect to"),
		serverRootCACert:   set.String("server-root-ca-cert", "", "Optional path to root server CA cert"),
		clientCert:         set.String("client-cert", "", "Optional path to client cert, mutually exclusive with API key"),
		clientKey:          set.String("client-key", "", "Optional path to client key, mutually exclusive with API key"),
		serverName:         set.String("server-name", "", "Server name to use for verifying the server's certificate"),
		insecureSkipVerify: set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name"),
		apiKey:             set.String("api-key", "", "Optional API key, mutually exclusive with cert/key"),
		isParsed:           set.Parsed,
	}
}

func (f *clientFlags) ClientOptions() (client.Options, error) {
	if !f.isParsed() {
		return client.Options{}, errors.New("flags not yet parsed")
	}
	if *f.clientCert != "" && *f.clientKey == "" || *f.clientCert == "" && *f.clientKey != "" {
		return client.Options{}, fmt.Errorf("either both or neither of -client-key and -client-cert are required")
	}
	if *f.clientCert != "" && *f.apiKey != "" {
		return client.Options{}, fmt.Errorf("either -client-cert and -client-key or -api-key are required, not both")
	}

	var connectionOptions client.ConnectionOptions
	var credentials client.Credentials
	if *f.clientCert != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(*f.clientCert, *f.clientKey)
		if err != nil {
			return client.Options{}, fmt.Errorf("failed loading client cert and key: %w", err)
		}
		// Load server CA if given
		var serverCAPool *x509.CertPool
		if *f.serverRootCACert != "" {
			serverCAPool = x509.NewCertPool()
			b, err := os.ReadFile(*f.serverRootCACert)
			if err != nil {
				return client.Options{}, fmt.Errorf("failed reading server CA: %w", err)
			} else if !serverCAPool.AppendCertsFromPEM(b) {
				return client.Options{}, fmt.Errorf("server CA PEM file invalid")
			}
		}
		connectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				// In production use GetClientCertificate to allow loading new certs upon expiration
				// without requiring application restart.
				Certificates:       []tls.Certificate{cert},
				RootCAs:            serverCAPool,
				ServerName:         *f.serverName,
				InsecureSkipVerify: *f.insecureSkipVerify,
			},
		}
	} else if *f.apiKey != "" {
		connectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{},
			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", *f.namespace),
							method,
							req,
							reply,
							cc,
							opts...,
						)
					},
				),
			},
		}
		credentials = client.NewAPIKeyStaticCredentials(*f.apiKey)
	}

	return client.Options{
		HostPort:          *f.targetHost,
		Namespace:         *f.namespace,
		ConnectionOptions: connectionOptions,
		Credentials:       credentials,
	}, nil
}
