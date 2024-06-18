package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"math/rand"

	accumulator "github.com/temporalio/samples-go/accumulator"
	"go.temporal.io/sdk/client"
)

var WorkflowIDPrefix = "accumulate"

var TaskQueue = "accumulate_greetings"

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// setup which tests to run
	// by default it will run an accumulation with a few (20) signals
	// to a set of 4 buckets with Signal To Start
	triggerContinueAsNewWarning := false

	testSignalEdgeCases := true
	// configure signal edge cases to test
	testSignalAfterWorkflowExit := false
	testSignalAfterExitSignal := !testSignalAfterWorkflowExit
	testDuplicate := true
	testIgnoreBadBucket := true

	// setup to send signals
	bucket := "blue"
	buckets := []string{"red", "blue", "green", "yellow"}
	names := []string{"Genghis Khan", "Missy", "Bill", "Ted", "Rufus", "Abe"}

	maxSignals := 20
	if triggerContinueAsNewWarning {
		maxSignals = 5000
	}

	for i := 0; i < maxSignals; i++ {
		bucketIndex := rand.Intn(len(buckets))
		bucket = buckets[bucketIndex]
		nameIndex := rand.Intn(len(names))

		greeting := accumulator.AccumulateGreeting{
			GreetingText: names[nameIndex],
			Bucket:       bucket,
			GreetingKey:  "key-" + fmt.Sprint(i),
		}

		greetings := accumulator.GreetingsInfo{
			BucketKey:          bucket,
			GreetingsList:      []accumulator.AccumulateGreeting{},
			UniqueGreetingKeys: make(map[string]bool),
		}
		time.Sleep(5 * time.Millisecond)

		WorkflowID := WorkflowIDPrefix + "-" + bucket
		workflowOptions := client.StartWorkflowOptions{
			ID:        WorkflowID,
			TaskQueue: TaskQueue,
		}
		we, err := c.SignalWithStartWorkflow(context.Background(), WorkflowID, "greeting", greeting, workflowOptions, accumulator.AccumulateSignalsWorkflow, greetings)
		if err != nil {
			log.Fatalln("Unable to signal with start workflow", err)
		}
		log.Println("Signaled/Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID(), "signal:", greeting.GreetingText)

	}

	// skip further testing
	if !testSignalEdgeCases {
		return
	}

	// now we will try sending a signals near time of workflow exit
	bucket = "purple"
	WorkflowID := WorkflowIDPrefix + "-" + bucket

	workflowOptions := client.StartWorkflowOptions{
		ID:        WorkflowID,
		TaskQueue: TaskQueue,
	}

	suzieGreeting := new(accumulator.AccumulateGreeting)
	suzieGreeting.GreetingText = "Suzie Robot"
	suzieGreeting.Bucket = bucket
	suzieGreeting.GreetingKey = "11235813"

	greetings := accumulator.GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []accumulator.AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}

	// start the workflow async and then signal it
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, accumulator.AccumulateSignalsWorkflow, greetings)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// After start for AccumulateSignalsWorkflow returns, the workflow is guaranteed to be
	// started, so we can send a signal to it using the workflow ID and Run ID
	// This workflow keeps receiving signals until exit is called or the timer finishes with no signals

	// When the workflow is started the accumulateGreetings will block for the
	// previously defined conditions
	// Send the first workflow signal
	err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "greeting", suzieGreeting)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	log.Println("Sent " + suzieGreeting.GreetingText + " to " + we.GetID())

	// This test signals exit, then waits for the workflow to end
	// signals after this will error, as the workflow execution already completed
	if testSignalAfterWorkflowExit {
		err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "exit", "")
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
		log.Println(we.GetID() + ":Sent exit")
		var exitSignalResults string
		err = we.Get(context.Background(), &exitSignalResults)
		if err != nil {
			log.Fatalln("Unable to get workflow results", err)
		}
		log.Println(we.GetID() + "-" + exitSignalResults + ": execution results: " + exitSignalResults)
	}

	// This test sends an exit signal, does not wait for workflow to exit, then sends a signal
	// this demonstrates Temporal history rollback
	// see https://community.temporal.io/t/continueasnew-signals/1008/7
	if testSignalAfterExitSignal {
		err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "exit", "")
		if err != nil {
			log.Fatalln("Unable to signal workflow "+we.GetID(), err)
		}
		log.Println(we.GetID() + ": Sent exit")
	}

	// Test sending more signals after workflow exit

	janeGreeting := new(accumulator.AccumulateGreeting)
	janeGreeting.GreetingText = "Jane Robot"
	janeGreeting.Bucket = bucket
	janeGreeting.GreetingKey = "112358132134"
	err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "greeting", janeGreeting)
	if err != nil {
		log.Println("Workflow " + we.GetID() + " not found to signal - this is intentional: " + err.Error())
	}
	log.Println("Sent " + janeGreeting.GreetingText + " to " + we.GetID())

	if testIgnoreBadBucket {
		// send a third signal with an incorrect bucket - this will be ignored
		// can use workflow update to validate and reject a request if needed
		badBucketGreeting := new(accumulator.AccumulateGreeting)
		badBucketGreeting.GreetingText = "Ozzy Robot"
		badBucketGreeting.Bucket = "taupe"
		badBucketGreeting.GreetingKey = "112358132134"
		err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "greeting", badBucketGreeting)
		if err != nil {
			log.Println("Workflow " + we.GetID() + " not found to signal - this is intentional: " + err.Error())
		}
		log.Println("Sent " + badBucketGreeting.GreetingText + " to " + we.GetID())
	}

	if testDuplicate {
		// intentionally send a duplicate signal
		err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), "greeting", janeGreeting)
		if err != nil {
			log.Println("Workflow " + we.GetID() + " not found to signal - this is intentional: " + err.Error())
		}
		log.Println("Sent Duplicate " + janeGreeting.GreetingText + " to " + we.GetID())
	}

	if !testSignalAfterWorkflowExit {
		// wait for results if we haven't waited for them yet
		var exitSignalResults string
		err = we.Get(context.Background(), &exitSignalResults)
		if err != nil {
			log.Fatalln("Unable to get workflow results", err)
		}
		log.Println(we.GetID() + ": Execution results: " + exitSignalResults)
	}

}
