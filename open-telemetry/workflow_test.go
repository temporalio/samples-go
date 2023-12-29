package open_telemetry_test

import (
	"context"
	"testing"
	"time"

	commonpb "go.temporal.io/api/common/v1"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	open_telemetry "github.com/temporalio/samples-go/open-telemetry"
)

func Test_Workflow_Integration(t *testing.T) {
	ctx := context.Background()
	tracerProvider, err := open_telemetry.CreateTraceProvider(ctx)
	require.NoError(t, err)
	tracer, err := opentelemetry.NewTracer(opentelemetry.TracerOptions{
		HeaderKey: "span-key",
		Tracer:    tracerProvider.Tracer("go-sdk"),
	})
	require.NoError(t, err)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		Interceptors: []interceptor.WorkerInterceptor{interceptor.NewTracingInterceptor(tracer)},
	})

	span, err := tracer.StartSpan(&interceptor.TracerStartSpanOptions{
		Operation: "StartWorkflow",
		Name:      "Workflow",
		Tags:      map[string]string{"temporalWorkflowID": "TestWorkflowID"},
		ToHeader:  true,
		Time:      time.Now(),
	})
	require.NoError(t, err)
	defer span.Finish(&interceptor.TracerFinishSpanOptions{})

	s, err := tracer.MarshalSpan(span)
	require.NoError(t, err)

	payload, err := converter.GetDefaultDataConverter().ToPayload(s)
	require.NoError(t, err)

	// Put on header
	headers := make(map[string]*commonpb.Payload)
	headers["span-key"] = payload
	testSuite.SetHeader(&commonpb.Header{Fields: headers})
	env.RegisterWorkflow(open_telemetry.ChildWorkflow)
	env.RegisterActivity(open_telemetry.Activity)
	env.ExecuteWorkflow(open_telemetry.Workflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.NoError(t, env.GetWorkflowResult(nil))
}
