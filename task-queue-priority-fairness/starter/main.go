package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	task_queue_priority_fairness "github.com/temporalio/samples-go/task-queue-priority-fairness"
)

func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	jobs := task_queue_priority_fairness.BuildJobs()
	workflowID := fmt.Sprintf("task-queue-priority-fairness-%d", time.Now().UnixNano())
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: task_queue_priority_fairness.WorkflowTaskQueue,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, task_queue_priority_fairness.RenderWorkflow, jobs)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow.", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	fmt.Println("If you are running the full backlog demo, start the Activity Worker now:")
	fmt.Println("go run task-queue-priority-fairness/worker/main.go -mode activity")

	var results []task_queue_priority_fairness.RenderResult
	if err := we.Get(context.Background(), &results); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}

	fmt.Println(task_queue_priority_fairness.FormatResults(results))
	summary := task_queue_priority_fairness.SummarizeResults(results)
	fmt.Print(task_queue_priority_fairness.FormatSummary(summary))
}
