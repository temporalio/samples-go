package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/helloworld"
	"github.com/temporalio/samples-go/serverjwtauth"
)

func main() {
	key, jwk, err := serverjwtauth.ReadKey()
	if err != nil {
		log.Fatalln(err)
	}
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HeadersProvider: &serverjwtauth.JWTHeadersProvider{
			Config: serverjwtauth.JWTConfig{
				Key:   key,
				KeyID: jwk.KeyID,
				Permissions: []string{
					"default:read",
					"default:write",
				},
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "server-jwt-auth",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
