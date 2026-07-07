package adk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	"google.golang.org/adk/v2/model"

	"go.temporal.io/sdk/contrib/googleadk"

	adk "github.com/temporalio/samples-go/googleadk"
)

// TestAgentWorkflow drives AgentWorkflow through a scripted FakeModel — no live
// LLM. The model first calls the get_weather tool, then answers with text; the
// test asserts the tool ran (as a real activity) and its result reached the
// answer.
func TestAgentWorkflow(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(adk.AgentWorkflow)
	env.RegisterActivityWithOptions(adk.GetWeather, activity.RegisterOptions{Name: adk.WeatherToolName})

	// Scripted model: turn 1 requests get_weather, turn 2 produces the final text.
	fm := googleadk.NewFakeModel(
		googleadk.FunctionCallResponse("call-1", adk.WeatherToolName, map[string]any{"city": "San Francisco"}),
		googleadk.TextResponse("It's sunny, 72°F in San Francisco."),
	)
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			adk.ModelName: func(context.Context, string) (model.LLM, error) { return fm, nil },
		},
	})
	require.NoError(t, err)
	// Register the integration's InvokeModel activity by its stable name.
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	env.ExecuteWorkflow(adk.AgentWorkflow, "What's the weather in San Francisco?")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var answer string
	require.NoError(t, env.GetWorkflowResult(&answer))
	assert.Contains(t, answer, "sunny")
}
