package helloworldmtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}

// ParseClientOptionFlags parses the given arguments into client options. In
// some cases a failure will be returned as an error, in others the process may
// exit with help info.
func ParseClientOptionFlags(args []string) (client.Options, error) {
	// Parse args
	set := flag.NewFlagSet("hello-world-mtls", flag.ExitOnError)
	targetHost := set.String("target-host", "localhost:7233", "Host:port for the server")
	namespace := set.String("namespace", "default", "Namespace for the server")
	serverRootCACert := set.String("server-root-ca-cert", "", "Optional path to root server CA cert")
	clientCert := set.String("client-cert", "", "Required path to client cert")
	clientKey := set.String("client-key", "", "Required path to client key")
	serverName := set.String("server-name", "", "Server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")
	if err := set.Parse(args); err != nil {
		return client.Options{}, fmt.Errorf("failed parsing args: %w", err)
	} else if *clientCert == "" || *clientKey == "" {
		return client.Options{}, fmt.Errorf("-client-cert and -client-key are required")
	}

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

	return client.Options{
		HostPort:  *targetHost,
		Namespace: *namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            serverCAPool,
				ServerName:         *serverName,
				InsecureSkipVerify: *insecureSkipVerify,
			},
		},
	}, nil
}
