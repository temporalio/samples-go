package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"

	"go.temporal.io/sdk/contrib/googleadk"

	humanintheloop "github.com/temporalio/samples-go/googleadk/humanintheloop"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// The plugin registers the integration's model Activity on the worker. The
	// delete_resource tool is an in-workflow function tool (not an
	// ActivityAsTool), so there is no tool activity to register.
	adkPlugin, err := googleadk.NewPlugin(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			humanintheloop.ModelName: func(ctx context.Context, name string) (model.LLM, error) {
				// nil config reads GEMINI_API_KEY / GOOGLE_API_KEY from the env.
				return gemini.NewModel(ctx, name, nil)
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to build googleadk plugin", err)
	}

	w := worker.New(c, humanintheloop.TaskQueue, worker.Options{
		Plugins: []worker.Plugin{adkPlugin},
	})

	w.RegisterWorkflow(humanintheloop.ApprovalWorkflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
