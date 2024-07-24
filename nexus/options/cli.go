package options

import (
	"flag"
	"fmt"
	"os"

	"crypto/tls"
	"crypto/x509"

	"go.temporal.io/sdk/client"
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
	clientCert := set.String("client-cert", "", "Optional path to client cert")
	clientKey := set.String("client-key", "", "Optional path to client key")
	serverName := set.String("server-name", "", "Server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")

	if err := set.Parse(args); err != nil {
		return client.Options{}, fmt.Errorf("failed parsing args: %w", err)
	}
	if *clientCert != "" && *clientKey == "" || *clientCert == "" && *clientKey != "" {
		return client.Options{}, fmt.Errorf("either both or neither of -client-key and -client-cert are required")
	}

	var connectionOptions client.ConnectionOptions

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
	}

	return client.Options{
		HostPort:          *targetHost,
		Namespace:         *namespace,
		ConnectionOptions: connectionOptions,
	}, nil
}
