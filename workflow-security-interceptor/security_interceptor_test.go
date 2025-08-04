package workflow_security_interceptor_test

import (
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	wsi "github.com/temporalio/samples-go/workflow-security-interceptor"
)

func TestSecurityInterceptorWorkflow(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		Interceptors: []interceptor.WorkerInterceptor{wsi.NewWorkerInterceptor()},
	})
	env.RegisterWorkflow(wsi.Workflow)
	env.RegisterWorkflow(wsi.ChildWorkflow)
	env.RegisterWorkflow(wsi.ProhibitedChildWorkflow)
	env.RegisterActivity(wsi.ValidateChildWorkflowTypeActivity)

	env.ExecuteWorkflow(wsi.Workflow)
	err := env.GetWorkflowError()

	require.Error(t, err, "expected prohibited child workflow to fail")
	require.Contains(t, err.Error(), "Child workflow type \"ProhibitedChildWorkflow\" not allowed",
		"expected error to contain prohibited child workflow type message")
}
