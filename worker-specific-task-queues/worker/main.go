package main

import (
	"log"
	"sync"

	activities_sticky_queues "github.com/temporalio/samples-go/activities-sticky-queues"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/google/uuid"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	stickTaskQueue := activities_sticky_queues.StickyTaskQueue{
		TaskQueue: uuid.New().String(),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w := worker.New(c, "activities-sticky-queues", worker.Options{})
		w.RegisterWorkflow(activities_sticky_queues.FileProcessingWorkflow)

		w.RegisterActivityWithOptions(stickTaskQueue.GetStickyTaskQueue, activity.RegisterOptions{
			Name: "GetStickyTaskQueue",
		})
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a new worker listening on the stick queue
		stickWorker := worker.New(c, stickTaskQueue.TaskQueue, worker.Options{})

		stickWorker.RegisterActivity(activities_sticky_queues.DownloadFile)
		stickWorker.RegisterActivity(activities_sticky_queues.ProcessFile)
		stickWorker.RegisterActivity(activities_sticky_queues.DeleteFile)

		err = stickWorker.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()
	// Wait for both worker to close
	wg.Wait()
}
