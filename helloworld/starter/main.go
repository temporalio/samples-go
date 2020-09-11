// @@@START go-helloworld-sample-workflow-starter
package main

import (
  "context"
  "log"

  "go.temporal.io/sdk/client"

  "github.com/temporalio/temporal-go-samples/helloworld"
)

func main() {
  // Create a Temporal Go Client
  c, err := client.NewClient(client.Options{})
  if err != nil {
    log.Fatalln("unable to create client", err)
  }
  defer c.Close()
  // Task Queue that the Workflow and Activity Tasks will be sent to
  // Must be the same name as the Task Queue the Worker is listening to
  taskQueue := "hello-world-task-queue"
  // Create Workflow options
  workflowOptions := client.StartWorkflowOptions{
    TaskQueue: taskQueue,
  }
  // This is the name we are feeding to the Workflow
  // Which will in turn, be fed to the Activity
  // And will be appended to "Hello "
  name := "World"
  // Execute the Workflow
  wrkflw, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.HelloWorldWorkflow, name)
  if err != nil {
    log.Fatalln("unable to execute Workflow", err)
  }
  // Get the result of the Workflow
  var result string
  err = wrkflw.Get(context.Background(), &result)
  if err != nil {
    log.Fatalln("unable to get Workflow result", err)
  }
  // Print the Workflow result to the console
  log.Println("Workflow result: ", result)
}
// @@@END
