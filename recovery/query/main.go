package main

import (
	"context"
	"flag"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/recovery"
)

func main() {
	var workflowID string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", recovery.QueryName)
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	log.Println("Received query result", "Result", result)
}
