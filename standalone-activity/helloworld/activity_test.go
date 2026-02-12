package helloworld

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

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

func Test_StandaloneActivity_Using_DevServer(t *testing.T) {
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		CachedDownload: testsuite.CachedDownload{
			Version: "v1.5.2-standalone-activity-server",
		},
		ExtraArgs: []string{
			"--dynamic-config-value", "activity.enableStandalone=true",
			"--dynamic-config-value", "history.enableChasm=true",
			"--dynamic-config-value", "history.enableTransitionHistory=true",
		},
	})
	require.NoError(t, err)
	defer func() { _ = server.Stop() }()

	c := server.Client()
	taskQueue := "standalone-activity-test"

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterActivity(Activity)
	require.NoError(t, w.Start())
	defer w.Stop()

	activityOptions := client.StartActivityOptions{
		ID:                     "test-activity-id",
		TaskQueue:              taskQueue,
		ScheduleToCloseTimeout: 10 * time.Second,
	}

	handle, err := c.ExecuteActivity(context.Background(), activityOptions, Activity, "Temporal")
	require.NoError(t, err)
	require.NotEmpty(t, handle.GetID())
	require.NotEmpty(t, handle.GetRunID())

	var result string
	err = handle.Get(context.Background(), &result)
	require.NoError(t, err)
	require.Equal(t, "Hello Temporal!", result)
}
