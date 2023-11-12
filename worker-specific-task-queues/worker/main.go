package main

import (
	"log"
	"sync"

	worker_specific_task_queues "github.com/temporalio/samples-go/worker-specific-task-queues"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/google/uuid"
)

func main() {
	// The client and worker are heavyweight objects that should generally be created once per process.
	// In this case, we create a single client but two workers since we need to handle Activities on multiple task queues.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	uniqueTaskQueue := worker_specific_task_queues.WorkerSpecificTaskQueue{
		TaskQueue: uuid.New().String(),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w := worker.New(c, "shared-task-queue", worker.Options{})
		w.RegisterWorkflow(worker_specific_task_queues.FileProcessingWorkflow)

		w.RegisterActivityWithOptions(uniqueTaskQueue.GetWorkerSpecificTaskQueue, activity.RegisterOptions{
			Name: "GetWorkerSpecificTaskQueue",
		})
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a new worker listening on the unique queue
		uniqueTaskQueueWorker := worker.New(c, uniqueTaskQueue.TaskQueue, worker.Options{})

		uniqueTaskQueueWorker.RegisterActivity(worker_specific_task_queues.DownloadFile)
		uniqueTaskQueueWorker.RegisterActivity(worker_specific_task_queues.ProcessFile)
		uniqueTaskQueueWorker.RegisterActivity(worker_specific_task_queues.DeleteFile)

		err = uniqueTaskQueueWorker.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()
	// Wait for both workers to close
	wg.Wait()
}
