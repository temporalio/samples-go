package helloworld

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, "Temporal").Return("Hello Temporal!", nil)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)
}

func Test_Activity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(Activity)

	val, err := env.ExecuteActivity(Activity, "World")
	require.NoError(t, err)

	var res string
	require.NoError(t, val.Get(&res))
	require.Equal(t, "Hello World!", res)
}

func Test_Using_DevServer(t *testing.T) {
	//"" will let use a random port in local env
	hostPort := ""
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{ClientOptions: &client.Options{HostPort: hostPort}})
	require.NoError(t, err)
	require.NotNil(t, server)
	defer func() { _ = server.Stop() }()

	var (
		c       client.Client
		w       worker.Worker
		wInChan <-chan interface{}
	)

	taskQ := "hello-world"

	ch := make(chan interface{})
	go func() {
		c = server.Client()
		w = worker.New(c, taskQ, worker.Options{})
		wInChan = worker.InterruptCh()

		ch <- struct{}{}

		_ = w.Run(wInChan)
	}()

	<-ch

	require.NotNil(t, c)
	require.NotNil(t, w)
	require.NotNil(t, wInChan)

	// register activity and workflow
	w.RegisterWorkflow(Workflow)
	w.RegisterActivity(Activity)

	// run the workflow application (equivalent to starter/main.go)
	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: taskQ,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, Workflow, "Temporal")
	require.NoError(t, err)
	require.NotNil(t, we)

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	require.NoError(t, err)
	require.Equal(t, "Hello Temporal!", result)

	// stop worker
	w.Stop()

	// stop server
	err = server.Stop()
	require.NoError(t, err)
}
