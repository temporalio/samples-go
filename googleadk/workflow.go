// Package adk demonstrates running a Google ADK (adk-go) agent durably on
// Temporal with the go.temporal.io/sdk/contrib/googleadk integration. The agent's
// orchestration loop runs inside the workflow; the model call runs as a Temporal
// Activity, and the get_weather tool is an existing Temporal activity exposed to
// the agent via googleadk.ActivityAsTool.
package adk

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"

	"go.temporal.io/sdk/contrib/googleadk"
)

const (
	// TaskQueue is the task queue the worker listens on and the starter targets.
	TaskQueue = "google-adk"

	// ModelName is the Gemini model the agent uses. The workflow only ships this
	// name; the real model is reconstructed worker-side by the ModelFactory the
	// worker registers (see worker/main.go).
	ModelName = "gemini-2.0-flash"

	// WeatherToolName is both the tool name the model sees and the Activity name
	// the tool dispatches, so the worker must register GetWeather under this name.
	WeatherToolName = "get_weather"
)

// GetWeatherInput is the argument schema the model fills in for the weather tool.
type GetWeatherInput struct {
	City string `json:"city"`
}

// GetWeatherOutput is the weather tool's result, handed back to the model.
type GetWeatherOutput struct {
	City       string `json:"city"`
	Conditions string `json:"conditions"`
}

// GetWeather is an ordinary Temporal activity. Exposed to the agent as a tool via
// googleadk.ActivityAsTool, it runs durably worker-side (retried, timed-out,
// visible in the UI) whenever the model calls it — the recommended pattern for a
// tool that does I/O. A real implementation would call a weather API here.
// @@@SNIPSTART googleadk-hello-tool
func GetWeather(ctx context.Context, in GetWeatherInput) (GetWeatherOutput, error) {
	return GetWeatherOutput{City: in.City, Conditions: "sunny, 72°F"}, nil
}

// @@@SNIPEND

// AgentWorkflow runs a native ADK agent durably. The model call inside r.Run is
// dispatched to the InvokeModel Activity by googleadk.NewModel, and the
// get_weather tool call is dispatched to the GetWeather Activity by
// googleadk.ActivityAsTool.
// @@@SNIPSTART googleadk-hello-workflow
func AgentWorkflow(ctx workflow.Context, question string) (string, error) {
	weatherTool, err := googleadk.ActivityAsTool(GetWeather, googleadk.ActivityToolOptions{
		Name:        WeatherToolName,
		Description: "Get the current weather for a city.",
	})
	if err != nil {
		return "", err
	}

	// Build the agent the ordinary ADK way. NewModel is a model.LLM that carries
	// only the model name in-workflow; the real Gemini client lives worker-side.
	root, err := llmagent.New(llmagent.Config{
		Name:        "assistant",
		Description: "a helpful weather assistant",
		Model:       googleadk.NewModel(ModelName),
		Instruction: "Answer the user's question. Use the get_weather tool when asked about the weather.",
		Tools:       []tool.Tool{weatherTool},
	})
	if err != nil {
		return "", err
	}

	r, err := runner.New(runner.Config{
		AppName:           "weather",
		Agent:             root,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		return "", err
	}

	// NewContext bridges the workflow.Context into the context ADK reads its
	// determinism/executor seams from. Pass it straight to Run.
	adkCtx := googleadk.NewContext(ctx)
	msg := genai.NewContentFromText(question, genai.RoleUser)

	var answer string
	for ev, err := range r.Run(adkCtx, "user-1", "session-1", msg, agent.RunConfig{}) {
		if err != nil {
			return "", err
		}
		if ev != nil && ev.Content != nil {
			for _, p := range ev.Content.Parts {
				if p != nil && p.Text != "" {
					answer = p.Text
				}
			}
		}
	}
	return answer, nil
}

// @@@SNIPEND
