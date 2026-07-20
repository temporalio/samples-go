package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"

	"go.temporal.io/sdk/contrib/googleadk"

	multiagent "github.com/temporalio/samples-go/googleadk/multiagent"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, multiagent.TaskQueue, worker.Options{})

	w.RegisterWorkflow(multiagent.MultiAgentWorkflow)
	// Register GetWeather under the tool name the ActivityAsTool dispatches, so the
	// weather specialist's get_weather call resolves to this activity.
	w.RegisterActivityWithOptions(multiagent.GetWeather, activity.RegisterOptions{Name: multiagent.WeatherToolName})

	// Register the integration's model Activity. Every agent in the tree uses a
	// distinct model name so they can be scripted independently in tests; here
	// they all resolve to the same real Gemini model. The API key is read
	// worker-side and never crosses into the workflow.
	gemModel := func(ctx context.Context, name string) (model.LLM, error) {
		// nil config reads GEMINI_API_KEY / GOOGLE_API_KEY from the env.
		return gemini.NewModel(ctx, "gemini-2.0-flash", nil)
	}
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			multiagent.CoordinatorModelName: gemModel,
			multiagent.WeatherModelName:     gemModel,
			multiagent.JokesModelName:       gemModel,
		},
	})
	if err != nil {
		log.Fatalln("Unable to build googleadk activities", err)
	}
	acts.Register(w)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
