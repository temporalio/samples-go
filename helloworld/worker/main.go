// @@@START go-helloworld-sample-worker
package main

import (
  "log"
  // Import the Temporal Go SDK client package
  "go.temporal.io/sdk/client"
  // Import the SDK Worker pacakge
  "go.temporal.io/sdk/worker"
  // Import the helloworld package
  "github.com/temporalio/temporal-go-samples/helloworld"
)

func main() {
  // Create a Temporal Go Client
  c, err := client.NewClient(client.Options{})
  if err != nil {
    log.Fatalln("unable to create client", err)
  }
  defer c.Close()
  // Task Queue that the Worker will listen to
  // Must be the same name as the Task Queue the Workflow is sent to
  taskQueue := "helloworld-task-queue"
  // Create a Worker that is listening to the taskQueue
  wrkr := worker.New(c, taskQueue, worker.Options{})
  // Register Workflow with the Worker
  wrkr.RegisterWorkflow(helloworld.HelloWorldWorkflow)
  // Register Activity with the Worker
  wrkr.RegisterActivity(helloworld.HelloWorldActivity)
  // Run the Worker
  if err = wrkr.Run(worker.InterruptCh()); err != nil {
    log.Fatalln("unable to run Worker", err)
  }
}
// @@@END
