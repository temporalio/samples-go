package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	streams "github.com/temporalio/samples-go/workflowstreams"
)

// The LLM scenario runs on its own worker and task queue to keep its OpenAI
// dependency isolated. Set OPENAI_API_KEY in the environment before running it.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, streams.LLMTaskQueue, worker.Options{})

	w.RegisterWorkflow(streams.LLMWorkflow)
	w.RegisterActivity(streams.StreamCompletion)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
