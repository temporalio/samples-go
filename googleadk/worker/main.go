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

	adk "github.com/temporalio/samples-go/googleadk"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// @@@SNIPSTART googleadk-hello-worker
	w := worker.New(c, adk.TaskQueue, worker.Options{})

	w.RegisterWorkflow(adk.AgentWorkflow)
	// Register GetWeather under the tool name the ActivityAsTool dispatches, so the
	// agent's get_weather call resolves to this activity.
	w.RegisterActivityWithOptions(adk.GetWeather, activity.RegisterOptions{Name: adk.WeatherToolName})

	// Register the integration's model Activity. The real Gemini model lives here,
	// behind the Activity boundary; the API key is read worker-side from the env
	// and never crosses into the workflow. Disable the model SDK's own retries so
	// Temporal's RetryPolicy is the single source of truth.
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			adk.ModelName: func(ctx context.Context, name string) (model.LLM, error) {
				// nil config reads GEMINI_API_KEY / GOOGLE_API_KEY from the env.
				return gemini.NewModel(ctx, name, nil)
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to build googleadk activities", err)
	}
	acts.Register(w)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	// @@@SNIPEND
}
