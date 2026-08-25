package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	metrics "github.com/temporalio/samples-go/googleadk/metrics"
)

func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "google-adk-metrics-workflow",
		TaskQueue: metrics.TaskQueue,
	}, metrics.AgentWorkflow, "In one sentence, what is a durable execution?")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	var answer string
	if err := run.Get(context.Background(), &answer); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Agent answer:", answer)
}
