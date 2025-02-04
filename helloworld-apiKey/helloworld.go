package helloworldapiKey

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	set := flag.NewFlagSet("hello-world-api-key", flag.ExitOnError)
	targetHost := set.String("target-host", "localhost:7233", "Host:port for the server")
	namespace := set.String("namespace", "default", "Namespace for the server")
	apiKey := set.String("api-key", "", "Optional API key, mutually exclusive with cert/key")

	if err := set.Parse(args); err != nil {
		return client.Options{}, fmt.Errorf("failed parsing args: %w", err)
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("TEMPORAL_CLIENT_API_KEY")
	}
	if *apiKey == "" {
		return client.Options{}, fmt.Errorf("-api-key or TEMPORAL_CLIENT_API_KEY env is required required")
	}

	connectionOptions := client.ConnectionOptions{
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
	credentials := client.NewAPIKeyStaticCredentials(*apiKey)

	return client.Options{
		HostPort:          *targetHost,
		Namespace:         *namespace,
		ConnectionOptions: connectionOptions,
		Credentials:       credentials,
	}, nil
}
