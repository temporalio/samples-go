package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	asyncactivity "github.com/temporalio/samples-go/async-activity"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Useful events to look for: timestamp of ActivityTaskScheduled,
	// ActivityTaskStarted and ActivityTaskCompleted (note that they
	// may not be in the correct timestamp order in the event history).
	w := worker.New(c, "async-activity", worker.Options{
		// Set this to 1 to make the activities run one after the other (note
		// how both are scheduled at the same time, but ActivityTaskStarted
		// differs).
		MaxConcurrentActivityExecutionSize: 2,
		// Set this to 0.5 to create some delay between when activities are
		// started. Note that in this case, the started time does not differ.
		// Only the completed time is different.
		WorkerActivitiesPerSecond: 2,
	})

	w.RegisterWorkflow(asyncactivity.AsyncActivityWorkflow)
	w.RegisterActivity(asyncactivity.HelloActivity)
	w.RegisterActivity(asyncactivity.ByeActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
