package main

import (
	"context"
	bsdc "github.com/temporalio/samples-go/blob-store-data-converter"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func main() {
	ctx := context.Background()

	bsClient := blobstore.NewClient()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		DataConverter: bsdc.NewDataConverter(
			converter.GetDefaultDataConverter(),
			bsClient,
		),
		// Use a ContextPropagator so that the KeyID value set in the workflow context is
		// also available in the context for activities.
		ContextPropagators: []workflow.ContextPropagator{
			bsdc.NewContextPropagator(),
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx = context.WithValue(ctx, bsdc.PropagatedValuesKey, bsdc.PropagatedValues{
		TenantID:       "tenant12",
		BlobNamePrefix: []string{"starter"},
	})

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "blobstore_codec",
		TaskQueue:                "blobstore_codec",
		WorkflowTaskTimeout:      10 * time.Second, // encoding/decoding time counts towards this timeout
		WorkflowIDConflictPolicy: enums.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}

	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		bsdc.Workflow,
		"StarterSays: big big blob",
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
