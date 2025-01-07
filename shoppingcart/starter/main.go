package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/shoppingcart"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "shopping_cart_1",
		TaskQueue: "shopping_cart",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, shoppingcart.CartWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

}
