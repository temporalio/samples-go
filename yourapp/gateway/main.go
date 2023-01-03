package main

import (
	"context"
	"log"
	"net/http"

	"github.com/temporalio/samples-go/yourapp"

	"go.temporal.io/sdk/client"
)

func main() {
	// Create a Temporal Client to communicate with the Temporal Cluster.
	// A Temporal Client is a heavyweight object that should be created just once per process.
	temporalClient, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal Client", err)
	}
	defer temporalClient.Close()
	// Start an HTTP server and listen on /start
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		startWorkflowHandler(w, r, temporalClient)
	})
	http.ListenAndServe(":8081", nil)
}

func startWorkflowHandler(w http.ResponseWriter, r *http.Request, temporalClient client.Client) {
	// Set the options for the Workflow Execution.
	// A Task Queue must be specified.
	// A custom Workflow Id is highly recommended.
	workflowOptions := client.StartWorkflowOptions{
		ID:        "your-workflow-id",
		TaskQueue: "your-custom-task-queue-name",
	}
	// Use an object as your Workflow Function parameter.
	// Objects enable your Function signature to remain compatible if fields change.
	workflowParams := yourapp.YourWorkflowParam{
		WorkflowParamX: "Hello",
		WorkflowParamY: 0,
	}
	// Make the call to the Temporal Cluster to start the Workflow Execution.
	workflowExecution, err := temporalClient.ExecuteWorkflow(
		context.Background(), 
		workflowOptions, 
		yourapp.YourWorkflowDefinition,
		workflowParams,
	)
	if err != nil {
		log.Fatalln("Unable to execute the Workflow", err)
	}
	log.Println("Started Workflow\n", "WorkflowID:", workflowExecution.GetID(), "\nRunID:", workflowExecution.GetRunID())
}
