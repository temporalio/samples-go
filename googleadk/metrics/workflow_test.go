package metrics

import (
	"context"
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	ubertally "github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/activity"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/testsuite"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"

	"go.temporal.io/sdk/contrib/googleadk"
)

func TestAgentWorkflowRecordsUsageMetrics(t *testing.T) {
	response := googleadk.TextResponse("Temporal makes the agent durable.")
	response.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:        11,
		CandidatesTokenCount:    7,
		ThoughtsTokenCount:      3,
		CachedContentTokenCount: 2,
		TotalTokenCount:         21,
	}
	scope := runWorkflow(t, response, "Temporal makes the agent durable.")

	assertCounters(t, scope.Snapshot(), map[string]int64{
		"google_adk_model_calls+agent=metrics-assistant,model=gemini-2.5-flash":         1,
		"google_adk_input_tokens+agent=metrics-assistant,model=gemini-2.5-flash":        11,
		"google_adk_output_tokens+agent=metrics-assistant,model=gemini-2.5-flash":       7,
		"google_adk_reasoning_tokens+agent=metrics-assistant,model=gemini-2.5-flash":    3,
		"google_adk_cached_input_tokens+agent=metrics-assistant,model=gemini-2.5-flash": 2,
		"google_adk_total_tokens+agent=metrics-assistant,model=gemini-2.5-flash":        21,
	})
}

func TestAgentWorkflowHandlesMissingUsageMetadata(t *testing.T) {
	scope := runWorkflow(t, googleadk.TextResponse("No usage was returned."), "No usage was returned.")

	assertCounters(t, scope.Snapshot(), map[string]int64{
		"google_adk_model_calls+agent=metrics-assistant,model=gemini-2.5-flash": 1,
	})
}

func TestAgentWorkflowRecordsModelErrors(t *testing.T) {
	scope := runWorkflowWithModel(t, errorModel{}, "", true)

	assertCounters(t, scope.Snapshot(), map[string]int64{
		"google_adk_model_errors+agent=metrics-assistant,model=gemini-2.5-flash": 1,
	})
}

func runWorkflow(t *testing.T, response *model.LLMResponse, want string) ubertally.TestScope {
	t.Helper()
	return runWorkflowWithModel(t, googleadk.NewFakeModel(response).WithName(ModelName), want, false)
}

func runWorkflowWithModel(t *testing.T, testModel model.LLM, want string, wantError bool) ubertally.TestScope {
	t.Helper()
	scope := ubertally.NewTestScope("", nil)
	var suite testsuite.WorkflowTestSuite
	suite.SetMetricsHandler(sdktally.NewMetricsHandler(scope))
	env := suite.NewTestWorkflowEnvironment()

	activities, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			ModelName: func(context.Context, string) (model.LLM, error) { return testModel, nil },
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(activities.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	env.ExecuteWorkflow(AgentWorkflow, "Why use Temporal?")
	require.True(t, env.IsWorkflowCompleted())
	if wantError {
		require.Error(t, env.GetWorkflowError())
		return scope
	}
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, want, result)
	return scope
}

func assertCounters(t *testing.T, snapshot ubertally.Snapshot, want map[string]int64) {
	t.Helper()
	got := make(map[string]int64, len(snapshot.Counters()))
	for key, counter := range snapshot.Counters() {
		if !strings.HasPrefix(key, "google_adk_") {
			continue
		}
		got[key] = counter.Value()
	}
	require.Equal(t, want, got)
}

type errorModel struct{}

func (errorModel) Name() string { return ModelName }

func (errorModel) GenerateContent(context.Context, *model.LLMRequest, bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(nil, errors.New("model failed"))
	}
}
