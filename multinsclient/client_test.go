package multinsclient_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/multinsclient"
	"github.com/temporalio/temporalite"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	serverlog "go.temporal.io/server/common/log"
)

func TestListWorkflow(t *testing.T) {
	ctx := context.TODO()
	const taskQueue = "tq1"

	// Start server with 4 namespaces
	namespaces := []string{"ns1", "ns2", "ns3", "ns4"}
	s, err := temporalite.NewServer(
		temporalite.WithNamespaces(namespaces...),
		temporalite.WithPersistenceDisabled(),
		temporalite.WithDynamicPorts(),
		temporalite.WithLogger(serverlog.NewNoopLogger()),
	)
	require.NoError(t, err)
	defer s.Stop()
	require.NoError(t, s.Start())

	var multiClientOptions multinsclient.Options
	namespaceClients := map[string]client.Client{}
	var expectedIDs []string
	for _, namespace := range namespaces {
		// Create client for namespace
		namespaceClients[namespace], err = s.NewClient(ctx, namespace)
		require.NoError(t, err)
		defer namespaceClients[namespace].Close()

		// Start worker for namespace
		worker := worker.New(namespaceClients[namespace], taskQueue, worker.Options{})
		worker.RegisterWorkflow(MyWorkflow)
		require.NoError(t, worker.Start())
		defer worker.Stop()

		// Add namespace to options
		multiClientOptions.Namespaces = append(multiClientOptions.Namespaces, multinsclient.Namespace{
			ClientOptions: client.Options{
				HostPort:  s.FrontendHostPort(),
				Namespace: namespace,
			},
		})

		// Run 3 workflows
		for _, workflowID := range []string{"wf1", "wf2", "wf3"} {
			expectedIDs = append(expectedIDs, namespace+"-"+workflowID)
			run, err := namespaceClients[namespace].ExecuteWorkflow(
				ctx, client.StartWorkflowOptions{ID: namespace + "-" + workflowID, TaskQueue: taskQueue}, MyWorkflow)
			require.NoError(t, err)
			require.NoError(t, run.Get(ctx, nil))
		}
	}

	// Create a multi client
	multiClient, err := multinsclient.New(multiClientOptions)
	require.NoError(t, err)
	defer multiClient.Close()

	// List with page size 2
	req := &workflowservice.ListWorkflowExecutionsRequest{PageSize: 2}
	var ids []string
	for {
		resp, err := multiClient.ListWorkflow(ctx, req)
		require.NoError(t, err)
		require.LessOrEqual(t, len(resp.Executions), 2)
		for _, exec := range resp.Executions {
			ids = append(ids, exec.Execution.WorkflowId)
		}
		if len(resp.NextPageToken) == 0 {
			break
		}
		req.NextPageToken = resp.NextPageToken
	}
	// Check IDs
	sort.Strings(ids)
	require.Equal(t, expectedIDs, ids)
}

func MyWorkflow(ctx workflow.Context) error {
	return nil
}
