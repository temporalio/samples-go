package dynamic_workflows

import (
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"testing"
)

func TestDynamicWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	env.RegisterDynamicWorkflow(DynamicWorkflow, workflow.DynamicRegisterOptions{})
	env.RegisterDynamicActivity(DynamicActivity, activity.DynamicRegisterOptions{})
	env.ExecuteWorkflow("dynamic-activity", "Hello", "World")
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
	var result string
	err := env.GetWorkflowResult(&result)
	assert.NoError(t, err)
	assert.Equal(t, "dynamic-activity - Hello - World", result)
}
