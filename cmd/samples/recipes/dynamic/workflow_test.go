package main

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/cadence/testsuite"
	"testing"
)

func TestDynamicWorkflow(t *testing.T) {
	a := assert.New(t)
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.OnActivity(getGreetingActivity).Return("Greet", nil).Times(1)
	env.OnActivity(getNameActivity).Return("Name", nil).Times(1)
	env.OnActivity(sayGreetingActivity, "Greet", "Name").Return("Greet Name", nil).Times(1)

	env.ExecuteWorkflow(SampleGreetingsWorkflow)

	a.True(env.IsWorkflowCompleted())
	a.NoError(env.GetWorkflowError())
	env.AssertExpectations(t)
}