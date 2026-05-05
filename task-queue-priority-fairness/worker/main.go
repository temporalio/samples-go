package main

import (
	"flag"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	task_queue_priority_fairness "github.com/temporalio/samples-go/task-queue-priority-fairness"
)

func main() {
	mode := flag.String("mode", "workflow", "worker mode: workflow or activity")
	flag.Parse()

	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	switch *mode {
	case "workflow":
		workflowWorker := newWorkflowWorker(c)
		if err := workflowWorker.Run(worker.InterruptCh()); err != nil {
			log.Fatalln("Unable to start workflow worker", err)
		}
	case "activity":
		activityWorker := newActivityWorker(c)
		if err := activityWorker.Run(worker.InterruptCh()); err != nil {
			log.Fatalln("Unable to start activity worker", err)
		}
	default:
		log.Fatalf("Unknown mode %q. Use workflow or activity.", *mode)
	}
}

func newWorkflowWorker(c client.Client) worker.Worker {
	w := worker.New(c, task_queue_priority_fairness.WorkflowTaskQueue, worker.Options{})
	w.RegisterWorkflow(task_queue_priority_fairness.RenderWorkflow)
	return w
}

func newActivityWorker(c client.Client) worker.Worker {
	w := worker.New(c, task_queue_priority_fairness.ActivityTaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
	})
	w.RegisterActivity(task_queue_priority_fairness.ProcessRenderJob)
	return w
}
