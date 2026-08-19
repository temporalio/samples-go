package nexus_standalone_activity_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/nexus-rpc/sdk-go/nexus"

	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-standalone-activity/handler"
	"github.com/temporalio/samples-go/nexus-standalone-activity/service"
)

const (
	taskQueue    = "nexus-standalone-activity-test"
	endpointName = "nexus-standalone-activity-test-endpoint"
)

func Test_NexusOperationStandaloneActivity_Using_DevServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start the dev server with standalone Nexus support and Activity callbacks enabled, which is
	// what lets a Standalone Activity complete a Nexus Operation.
	server, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{
		CachedDownload: testsuite.CachedDownload{
			Version: "v1.7.4-standalone-nexus-operations",
		},
		ExtraArgs: []string{
			"--dynamic-config-value", "activity.enableCallbacks=true",
		},
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

	// Register Nexus operations on the worker.
	w := worker.New(c, taskQueue, worker.Options{})

	svc := nexus.NewService(service.GreetingServiceName)
	require.NoError(t, svc.Register(handler.GreetOperation))
	w.RegisterNexusService(svc)
	w.RegisterActivity(handler.CreateGreetingActivity)
	require.NoError(t, w.Start())
	defer w.Stop()

	// Create a standalone NexusClient.
	nexusClient, err := c.NewNexusClient(client.NexusClientOptions{
		Endpoint: endpointName,
		Service:  service.GreetingServiceName,
	})
	require.NoError(t, err)

	handle, err := nexusClient.ExecuteOperation(ctx, service.GreetOperationName, service.GreetingInput{Name: "Test"}, client.StartNexusOperationOptions{
		ID:                     "greeting-" + uuid.NewString(),
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	require.NotEmpty(t, handle.GetID())

	var result service.GreetingOutput
	err = handle.Get(ctx, &result)
	require.NoError(t, err)
	require.Equal(t, "Hello, Test!", result.Message)
}
