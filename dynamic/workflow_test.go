package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.temporal.io/sdk/testsuite"
)

func TestDynamicWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(&Activities{})
	var a *Activities
	env.OnActivity(a.GetGreeting).Return("Greet", nil).Times(1)
	env.OnActivity(a.GetName).Return("Name", nil).Times(1)
	env.OnActivity(a.SayGreeting, "Greet", "Name").Return("Greet Name", nil).Times(1)

	env.ExecuteWorkflow(SampleGreetingsWorkflow)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
