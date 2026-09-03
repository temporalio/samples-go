package metrics

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.temporal.io/sdk/contrib/googleadk"
)

const (
	TaskQueue = "google-adk-metrics"
	ModelName = "gemini-2.5-flash"
	agentName = "metrics-assistant"
)

func AgentWorkflow(ctx workflow.Context, prompt string) (string, error) {
	handler := workflow.GetMetricsHandler(ctx).WithTags(map[string]string{
		"agent": agentName,
		"model": ModelName,
	})

	root, err := llmagent.New(llmagent.Config{
		Name:                agentName,
		Description:         "an assistant that demonstrates model usage metrics",
		Model:               googleadk.NewModel(ModelName),
		Instruction:         "Answer the user's question concisely.",
		AfterModelCallbacks: []llmagent.AfterModelCallback{afterModelMetricsCallback(handler)},
		OnModelErrorCallbacks: []llmagent.OnModelErrorCallback{
			modelErrorMetricsCallback(handler),
		},
	})
	if err != nil {
		return "", err
	}

	r, err := runner.New(runner.Config{
		AppName:           "metrics",
		Agent:             root,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		return "", err
	}

	message := genai.NewContentFromText(prompt, genai.RoleUser)
	var answer string
	for event, err := range r.Run(googleadk.NewContext(ctx), "user-1", "session-1", message, agent.RunConfig{}) {
		if err != nil {
			return "", err
		}
		if event != nil && event.Content != nil {
			for _, part := range event.Content.Parts {
				if part != nil && part.Text != "" {
					answer = part.Text
				}
			}
		}
	}
	return answer, nil
}

func afterModelMetricsCallback(handler client.MetricsHandler) llmagent.AfterModelCallback {
	return func(_ agent.Context, response *model.LLMResponse, responseErr error) (*model.LLMResponse, error) {
		if responseErr != nil || response == nil {
			return nil, nil
		}
		handler.Counter("google_adk_model_calls").Inc(1)
		if response.UsageMetadata == nil {
			return nil, nil
		}
		usage := response.UsageMetadata
		handler.Counter("google_adk_input_tokens").Inc(int64(usage.PromptTokenCount))
		handler.Counter("google_adk_output_tokens").Inc(int64(usage.CandidatesTokenCount))
		handler.Counter("google_adk_reasoning_tokens").Inc(int64(usage.ThoughtsTokenCount))
		handler.Counter("google_adk_cached_input_tokens").Inc(int64(usage.CachedContentTokenCount))
		handler.Counter("google_adk_total_tokens").Inc(int64(usage.TotalTokenCount))
		return nil, nil
	}
}

func modelErrorMetricsCallback(handler client.MetricsHandler) llmagent.OnModelErrorCallback {
	return func(_ agent.Context, _ *model.LLMRequest, _ error) (*model.LLMResponse, error) {
		handler.Counter("google_adk_model_errors").Inc(1)
		return nil, nil
	}
}
