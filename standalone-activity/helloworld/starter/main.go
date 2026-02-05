package main

import (
	"context"
	"github.com/temporalio/samples-go/standalone-activity/helloworld"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"log"
	"time"
)

// This sample is very similar to helloworld. The difference is that whereas in
// helloworld the activity is orchestrated by a workflow, in this sample the activity is
// executed directly by a client ("standalone activity").

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	activityOptions := client.StartActivityOptions{
		ID:        "standalone_activity_helloworld_ActivityID",
		TaskQueue: "standalone-activity-helloworld",
		// at least one of ScheduleToCloseTimeout or StartToCloseTimeout is required.
		ScheduleToCloseTimeout: 10 * time.Second,
	}

	// Normally we would execute a workflow, but in this case we are executing an activity directly.
	handle, err := c.ExecuteActivity(context.Background(), activityOptions, helloworld.Activity, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started standalone activity", "ActivityID", handle.GetID(), "RunID", handle.GetRunID())

	// Synchronously wait for the activity completion.
	var result string
	err = handle.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get standalone activity result", err)
	}
	log.Println("Activity result:", result)
}
