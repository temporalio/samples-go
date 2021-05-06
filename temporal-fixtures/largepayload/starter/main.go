package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/temporal-fixtures/largepayload"
	"go.temporal.io/sdk/client"
)

var (
	NumberOfWorkflows = 5
	PayloadSize       = 1 * 1024 * 1024
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	id := uuid.New()[0:4]

	memoToken := make([]byte, PayloadSize)
	rand.Read(memoToken)

	i := 1
	for i <= NumberOfWorkflows {
		workflowOptions := client.StartWorkflowOptions{
			ID:        "largepayload_" + id + "_" + strconv.Itoa(i),
			TaskQueue: "largepayload",
			Memo:      map[string]interface{}{"attr1": memoToken},
		}

		we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, largepayload.LargePayloadWorkflow, PayloadSize)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
		i++
	}
}
