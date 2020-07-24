package main

import (
	"context"
	"flag"
	"log"

	"go.temporal.io/sdk/client"
)

func main() {
	var workflowID, queryType string
	flag.StringVar(&workflowID, "w", "query_workflow", "WorkflowID.")
	flag.StringVar(&queryType, "t", "state", "Query type [state|__stack_trace].")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", queryType)
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	log.Println("Received query result", "Result", result)
}
