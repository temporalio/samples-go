package multiagent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"

	"google.golang.org/adk/v2/model"

	"go.temporal.io/sdk/contrib/googleadk"

	multiagent "github.com/temporalio/samples-go/googleadk/multiagent"
)

// scriptedModelFactory returns a ModelFactory yielding a single shared FakeModel
// so its scripted responses advance turn by turn across Activity invocations
// (turn 1 = first response, turn 2 = second, ...).
func scriptedModelFactory(responses ...*model.LLMResponse) googleadk.ModelFactory {
	fm := googleadk.NewFakeModel(responses...)
	return func(context.Context, string) (model.LLM, error) { return fm, nil }
}

// TestMultiAgentWorkflow drives the coordinator/specialist tree through scripted
// FakeModels — no live LLM. The coordinator emits a transfer_to_agent call to the
// weather specialist; the specialist then calls get_weather and answers. The test
// asserts the specialist's answer surfaced and that the get_weather tool ran (as a
// real activity), proving the transfer took effect.
func TestMultiAgentWorkflow(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(multiagent.MultiAgentWorkflow)
	env.RegisterActivityWithOptions(multiagent.GetWeather, activity.RegisterOptions{Name: multiagent.WeatherToolName})

	// Track whether the get_weather activity ran — it only does if the coordinator
	// successfully transferred to the weather specialist and the specialist called
	// the tool.
	var weatherRan bool
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, _ converter.EncodedValues) {
		if info.ActivityType.Name == multiagent.WeatherToolName {
			weatherRan = true
		}
	})

	// Per-agent scripted models keyed by their distinct model names.
	//   coordinator: turn 1 delegates to the weather specialist.
	//   weather specialist: turn 1 calls get_weather, turn 2 answers.
	//   jokes specialist: never invoked in this scenario.
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			multiagent.CoordinatorModelName: scriptedModelFactory(
				googleadk.FunctionCallResponse("t1", "transfer_to_agent", map[string]any{"agent_name": "weather"}),
			),
			multiagent.WeatherModelName: scriptedModelFactory(
				googleadk.FunctionCallResponse("call-1", multiagent.WeatherToolName, map[string]any{"city": "San Francisco"}),
				googleadk.TextResponse("It's sunny, 72°F in San Francisco."),
			),
			multiagent.JokesModelName: scriptedModelFactory(
				googleadk.TextResponse("Why did the cloud break up with the fog? It needed space."),
			),
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	env.ExecuteWorkflow(multiagent.MultiAgentWorkflow, "What's the weather in San Francisco?")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var answer string
	require.NoError(t, env.GetWorkflowResult(&answer))
	// The final answer came from the weather specialist after the transfer.
	assert.Contains(t, answer, "sunny")
	assert.True(t, weatherRan, "the transfer must reach the weather specialist so get_weather runs")
}
