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

	chat "github.com/temporalio/samples-go/googleadk/chat"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// The plugin registers the integration's model Activity on the worker. The
	// chat agent has no tools, so that is the only registration it brings.
	adkPlugin, err := googleadk.NewPlugin(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			chat.ModelName: func(ctx context.Context, name string) (model.LLM, error) {
				// nil config reads GEMINI_API_KEY / GOOGLE_API_KEY from the env.
				return gemini.NewModel(ctx, name, nil)
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to build googleadk plugin", err)
	}

	w := worker.New(c, chat.TaskQueue, worker.Options{
		Plugins: []worker.Plugin{adkPlugin},
	})

	w.RegisterWorkflow(chat.ChatWorkflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
