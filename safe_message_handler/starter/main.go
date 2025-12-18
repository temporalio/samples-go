package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	"github.com/google/uuid"
	"github.com/temporalio/samples-go/safe_message_handler"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "ClusterManagerWorkflow-" + uuid.NewString(),
		TaskQueue: "safe-message-handlers-task-queue",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, safe_message_handler.ClusterManagerWorkflow, safe_message_handler.ClusterManagerInput{
		TestContinueAsNew: true,
	})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	delay := time.Second * 10

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), safe_message_handler.StartCluster, nil)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	allocationUpdates := make([]client.WorkflowUpdateHandle, 0, 6)
	for i := range 6 {
		handle, err := c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
			WorkflowID:   we.GetID(),
			RunID:        we.GetRunID(),
			UpdateName:   safe_message_handler.AssignNodesToJobs,
			WaitForStage: client.WorkflowUpdateStageAccepted,
			Args: []interface{}{safe_message_handler.ClusterManagerAssignNodesToJobInput{
				TotalNumNodes: 2,
				JobName:       fmt.Sprintf("task-%d", i),
			}},
		})
		if err != nil {
			log.Fatalln("Unable to update workflow", err)
		}
		allocationUpdates = append(allocationUpdates, handle)
	}

	for _, handle := range allocationUpdates {
		err = handle.Get(context.Background(), nil)
		if err != nil {
			log.Fatalln("Unable to get workflow update result", err)
		}
	}
	log.Println("Sleeping")
	time.Sleep(delay)

	log.Println("Deleting jobs...")
	deletionUpdates := make([]client.WorkflowUpdateHandle, 0, 6)
	for i := range 6 {
		handle, err := c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
			WorkflowID:   we.GetID(),
			RunID:        we.GetRunID(),
			UpdateName:   safe_message_handler.DeleteJob,
			WaitForStage: client.WorkflowUpdateStageAccepted,
			Args: []interface{}{safe_message_handler.ClusterManagerDeleteJobInput{
				JobName: fmt.Sprintf("task-%d", i),
			}},
		})
		if err != nil {
			log.Fatalln("Unable to update workflow", err)
		}
		deletionUpdates = append(deletionUpdates, handle)
	}
	for _, handle := range deletionUpdates {
		err = handle.Get(context.Background(), nil)
		if err != nil {
			log.Fatalln("Unable to get workflow update result", err)
		}
	}

	err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), safe_message_handler.ShutdownCluster, nil)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	// Synchronously wait for the workflow completion.
	var result safe_message_handler.ClusterManagerResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Cluster shut down successfully:", result)
}
