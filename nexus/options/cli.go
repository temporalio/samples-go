// @@@SNIPSTART samples-go-nexus-cli
package options

import (
	"context"
	"flag"
	"fmt"
	"os"

	"crypto/tls"
	"crypto/x509"

	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ParseClientOptionFlags parses the given arguments into client options. In
// some cases a failure will be returned as an error, in others the process may
// exit with help info.
func ParseClientOptionFlags(args []string) (client.Options, error) {
	// Parse args
	set := flag.NewFlagSet("nexus-sample", flag.ExitOnError)
	targetHost := set.String("target-host", "localhost:7233", "Host:port for the Temporal service")
	namespace := set.String("namespace", "default", "Namespace to connect to")
	serverRootCACert := set.String("server-root-ca-cert", "", "Optional path to root server CA cert")
	clientCert := set.String("client-cert", "", "Optional path to client cert, mutually exclusive with API key")
	clientKey := set.String("client-key", "", "Optional path to client key, mutually exclusive with API key")
	serverName := set.String("server-name", "", "Server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")
	apiKey := set.String("api-key", "", "Optional API key, mutually exclusive with cert/key")

	if err := set.Parse(args); err != nil {
		return client.Options{}, fmt.Errorf("failed parsing args: %w", err)
	}
	if *clientCert != "" && *clientKey == "" || *clientCert == "" && *clientKey != "" {
		return client.Options{}, fmt.Errorf("either both or neither of -client-key and -client-cert are required")
	}
	if *clientCert != "" && *apiKey != "" {
		return client.Options{}, fmt.Errorf("either -client-cert and -client-key or -api-key are required, not both")
	}

	var connectionOptions client.ConnectionOptions
	var credentials client.Credentials
	if *clientCert != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(*clientCert, *clientKey)
		if err != nil {
			return client.Options{}, fmt.Errorf("failed loading client cert and key: %w", err)
		}

		// Load server CA if given
		var serverCAPool *x509.CertPool
		if *serverRootCACert != "" {
			serverCAPool = x509.NewCertPool()
			b, err := os.ReadFile(*serverRootCACert)
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
				ServerName:         *serverName,
				InsecureSkipVerify: *insecureSkipVerify,
			},
		}
	} else if *apiKey != "" {
		connectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{},
			DialOptions: []grpc.DialOption{
				grpc.WithUnaryInterceptor(
					func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
						return invoker(
							metadata.AppendToOutgoingContext(ctx, "temporal-namespace", *namespace),
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
		credentials = client.NewAPIKeyStaticCredentials(*apiKey)
	}

	return client.Options{
		HostPort:          *targetHost,
		Namespace:         *namespace,
		ConnectionOptions: connectionOptions,
		Credentials:       credentials,
	}, nil
}

// @@@SNIPEND
