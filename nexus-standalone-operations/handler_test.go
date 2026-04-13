package nexus_standalone_operations_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/nexus-rpc/sdk-go/nexus"

	nexuspb "go.temporal.io/api/nexus/v1"
	operatorservice "go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus/handler"
	"github.com/temporalio/samples-go/nexus/service"
)

const (
	taskQueue    = "nexus-standalone-operations-test"
	endpointName = "nexus-standalone-operations-test-endpoint"
)

// devServerExtraArgs returns the CLI flags needed to enable standalone Nexus
// operations on the dev server. These mirror the flags used in the SDK's own
// integration tests (sdk-go/internal/cmd/build/main.go:152-162).
func devServerExtraArgs() []string {
	return []string{
		"--http-port", "7243",
		// TODO Use correct DC values once the dev server build supporting standalone Nexus is released.
		"--dynamic-config-value", `component.callbacks.allowedAddresses=[{"Pattern":"*","AllowInsecure":true}]`,
		"--dynamic-config-value", `system.refreshNexusEndpointsMinWait="0s"`,
		"--dynamic-config-value", "component.nexusoperations.recordCancelRequestCompletionEvents=true",
		"--dynamic-config-value", "history.enableRequestIdRefLinks=true",
		"--dynamic-config-value", "history.enableChasm=true",
		"--dynamic-config-value", "component.nexusoperations.useSystemCallbackURL=false",
		"--dynamic-config-value", `component.nexusoperations.callback.endpoint.template="http://localhost:7243/namespaces/{{.NamespaceName}}/nexus/callback"`,
	}
}

func Test_StandaloneNexusOperations_Using_DevServer(t *testing.T) {
	// Skip until a dev server version supporting standalone Nexus operations is available.
	// Set ENABLE_STANDALONE_NEXUS_TESTS=1 to run this test once a compatible server exists.
	if os.Getenv("ENABLE_STANDALONE_NEXUS_TESTS") == "" {
		t.Skip("Skipping: standalone Nexus operations not yet supported by dev server. Set ENABLE_STANDALONE_NEXUS_TESTS=1 to enable.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start the dev server with standalone Nexus support.
	// TODO: Update the Version once a dev server build supporting standalone Nexus is released.
	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{
		CachedDownload: testsuite.CachedDownload{
			Version: "v1.5.2-standalone-activity-server",
		},
		ExtraArgs: devServerExtraArgs(),
	})
	require.NoError(t, err)
	defer func() { _ = server.Stop() }()

	c := server.Client()

	// Create a Nexus endpoint targeting our task queue.
	_, err = c.OperatorService().CreateNexusEndpoint(ctx, &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name: endpointName,
			Target: &nexuspb.EndpointTarget{
				Variant: &nexuspb.EndpointTarget_Worker_{
					Worker: &nexuspb.EndpointTarget_Worker{
						Namespace: "default",
						TaskQueue: taskQueue,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Register Nexus operations on the worker, reusing the handler from the nexus sample.
	w := worker.New(c, taskQueue, worker.Options{})

	svc := nexus.NewService(service.HelloServiceName)
	require.NoError(t, svc.Register(handler.EchoOperation, handler.HelloOperation))
	w.RegisterNexusService(svc)
	w.RegisterWorkflow(handler.HelloHandlerWorkflow)
	require.NoError(t, w.Start())
	defer w.Stop()

	// Create a standalone NexusClient.
	nexusClient, err := c.NewNexusClient(client.NexusClientOptions{
		Endpoint: endpointName,
		Service:  service.HelloServiceName,
	})
	require.NoError(t, err)

	// executeWithRetry retries ExecuteOperation until the endpoint has propagated.
	// The endpoint registry is eventually consistent.
	executeWithRetry := func(
		t *testing.T,
		opName string,
		input any,
		options client.StartNexusOperationOptions,
	) client.NexusOperationHandle {
		t.Helper()
		var handle client.NexusOperationHandle
		require.Eventually(t, func() bool {
			var execErr error
			handle, execErr = nexusClient.ExecuteOperation(ctx, opName, input, options)
			return execErr == nil
		}, 10*time.Second, 100*time.Millisecond, "timed out waiting for endpoint to propagate")
		return handle
	}

	// Test sync operation (Echo).
	t.Run("Echo sync operation", func(t *testing.T) {
		input := service.EchoInput{Message: "hello-nexus"}
		handle := executeWithRetry(t, service.EchoOperationName, input, client.StartNexusOperationOptions{
			ID:                     uuid.NewString(),
			ScheduleToCloseTimeout: 10 * time.Second,
		})
		require.NotEmpty(t, handle.GetID())

		var result service.EchoOutput
		err := handle.Get(ctx, &result)
		require.NoError(t, err)
		require.Equal(t, "hello-nexus", result.Message)
	})

	// Test async operation (Hello).
	t.Run("Hello async operation", func(t *testing.T) {
		input := service.HelloInput{Name: "Temporal", Language: service.EN}
		handle := executeWithRetry(t, service.HelloOperationName, input, client.StartNexusOperationOptions{
			ID:                     uuid.NewString(),
			ScheduleToCloseTimeout: 10 * time.Second,
		})
		require.NotEmpty(t, handle.GetID())

		var result service.HelloOutput
		err := handle.Get(ctx, &result)
		require.NoError(t, err)
		require.Equal(t, "Hello Temporal 👋", result.Message)
	})

	// Test ListNexusOperations (on client.Client, not NexusClient).
	t.Run("List operations", func(t *testing.T) {
		require.Eventually(t, func() bool {
			resp, listErr := c.ListNexusOperations(ctx, client.ListNexusOperationsOptions{
				Query: fmt.Sprintf("Endpoint = '%s'", endpointName),
			})
			if listErr != nil {
				return false
			}
			count := 0
			for metadata, iterErr := range resp.Results {
				if iterErr != nil {
					return false
				}
				if metadata.OperationID == "" || metadata.Endpoint != endpointName {
					return false
				}
				count++
			}
			return count > 0
		}, 10*time.Second, 500*time.Millisecond, "timed out waiting for operations to appear in list")
	})

	// Test CountNexusOperations (on client.Client, not NexusClient).
	t.Run("Count operations", func(t *testing.T) {
		require.Eventually(t, func() bool {
			resp, countErr := c.CountNexusOperations(ctx, client.CountNexusOperationsOptions{
				Query: fmt.Sprintf("Endpoint = '%s'", endpointName),
			})
			return countErr == nil && resp.Count > 0
		}, 10*time.Second, 500*time.Millisecond, "timed out waiting for count to reflect operations")
	})
}
