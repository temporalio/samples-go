// Package multiagent demonstrates a multi-agent Google ADK (adk-go) system
// running durably on Temporal with the go.temporal.io/sdk/contrib/googleadk
// integration. A "coordinator" root agent delegates to one of two specialist
// SubAgents — a weather specialist (which owns the get_weather ActivityAsTool)
// and a jokes specialist — via ADK's built-in transfer_to_agent mechanism.
//
// The entire multi-agent orchestration, including the transfer_to_agent hop, runs
// inside the Workflow; only the model calls (one per agent turn) and the
// get_weather tool run as Temporal Activities.
package multiagent

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
	TaskQueue = "google-adk-multiagent"

	// CoordinatorModelName, WeatherModelName, and JokesModelName are distinct
	// model names so the worker (and the test) can register a separate
	// ModelFactory per agent and script each one independently. In production
	// they all resolve to the same Gemini model behind the InvokeModel Activity.
	CoordinatorModelName = "gemini-2.0-flash-coordinator"
	WeatherModelName     = "gemini-2.0-flash-weather"
	JokesModelName       = "gemini-2.0-flash-jokes"

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

// GetWeather is an ordinary Temporal activity exposed to the weather specialist
// as a tool via googleadk.ActivityAsTool. It runs durably worker-side whenever
// the specialist calls it. A real implementation would call a weather API here.
func GetWeather(ctx context.Context, in GetWeatherInput) (GetWeatherOutput, error) {
	return GetWeatherOutput{City: in.City, Conditions: "sunny, 72°F"}, nil
}

// MultiAgentWorkflow builds a coordinator agent with two specialist SubAgents and
// runs the question through the tree. The coordinator decides which specialist to
// delegate to (emitting a transfer_to_agent call, which ADK resolves in-workflow);
// the chosen specialist then answers. The workflow returns the final text.
// @@@SNIPSTART googleadk-multiagent-workflow
func MultiAgentWorkflow(ctx workflow.Context, question string) (string, error) {
	weatherTool, err := googleadk.ActivityAsTool(GetWeather, googleadk.ActivityToolOptions{
		Name:        WeatherToolName,
		Description: "Get the current weather for a city.",
	})
	if err != nil {
		return "", err
	}

	// The weather specialist owns the get_weather tool.
	weather, err := llmagent.New(llmagent.Config{
		Name:        "weather",
		Description: "answers questions about the current weather in a city",
		Model:       googleadk.NewModel(WeatherModelName),
		Instruction: "You are a weather specialist. Use the get_weather tool to answer weather questions.",
		Tools:       []tool.Tool{weatherTool},
	})
	if err != nil {
		return "", err
	}

	// The jokes specialist just tells jokes.
	jokes, err := llmagent.New(llmagent.Config{
		Name:        "jokes",
		Description: "tells a light-hearted joke",
		Model:       googleadk.NewModel(JokesModelName),
		Instruction: "You are a comedian. Respond with a short, friendly joke.",
	})
	if err != nil {
		return "", err
	}

	// The coordinator delegates to whichever specialist fits the question. ADK
	// wires the parent/child relationship from SubAgents and exposes the built-in
	// transfer_to_agent tool automatically.
	coordinator, err := llmagent.New(llmagent.Config{
		Name:        "coordinator",
		Description: "routes the user's request to the right specialist",
		Model:       googleadk.NewModel(CoordinatorModelName),
		Instruction: "You are a router. Delegate weather questions to the weather agent " +
			"and requests for a joke to the jokes agent. Do not answer directly.",
		SubAgents: []agent.Agent{weather, jokes},
	})
	if err != nil {
		return "", err
	}

	r, err := runner.New(runner.Config{
		AppName:           "multiagent",
		Agent:             coordinator,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		return "", err
	}

	adkCtx := googleadk.NewContext(ctx)
	msg := genai.NewContentFromText(question, genai.RoleUser)

	var answer string
	for ev, err := range r.Run(adkCtx, "user-1", "session-1", msg, agent.RunConfig{}) {
		if err != nil {
			return "", err
		}
		if ev == nil || ev.Content == nil {
			continue
		}
		// Keep the last non-empty text produced by any agent in the tree; after a
		// transfer_to_agent hop this is the specialist's answer.
		for _, p := range ev.Content.Parts {
			if p != nil && p.Text != "" {
				answer = p.Text
			}
		}
	}
	return answer, nil
}

// @@@SNIPEND
