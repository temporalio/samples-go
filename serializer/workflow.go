package serializer

import (
	"context"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

type (
	NextRunParams struct {
		PreviousEventID int
		Items           map[int]*ResourceEvent
	}

	ResourceEvent struct {
		EventID int
	}
)

const (
	Resource_Event_Signal_Name = "resource_event"
	Task_Queue_Name            = "event_processor_tq"
)

/**
ResourceWorkflow is created for each resource in the system.  Use business identifier as workflowID when creating
an instance of ResourceWorkflow to guarantee only one workflow execution is running for a resource at any given point
in time.
Once the workflow is created it stays open and keeps on listening for more events on signal channel until closeWorkflow
timer fires.  Once the timer fires it closes the workflow execution and if there are still unprocessed events which
were received out of order it instead calls ContinueAsNew and passes the events to next run of workflow execution.
**/
func ResourceWorkflow(ctx workflow.Context, params NextRunParams) error {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Channel used for delivering events to workflow
	eventCh := workflow.GetSignalChannel(ctx, Resource_Event_Signal_Name)

	closeWorkflow := false
	// Selector used for message pump and close workflow timer
	selector := workflow.NewSelector(ctx)

	// Timer for closing the workflow
	selector.AddFuture(workflow.NewTimer(ctx, time.Hour),
		func(f workflow.Future) {
			// Workflow close timer fired.  Close workflow.
			closeWorkflow = true
		},
	)

	// Create a channel to communicate new events to event processor
	// Shutdown is also communicated by closing the channel
	ch := workflow.NewChannel(ctx)

	// State of resource to initialize the event processor
	// If the resource is stored to some long term storage when closing the workflow then you can run an activity
	// to load from the permament store so it can be correctly initialized before we start the event processor
	state := &EventProcessorState{
		PreviousEventID: params.PreviousEventID,
		Items:           params.Items,
	}

	// Start event processor which implements the pump for processing events
	processor := newEventProcessor(ch, state)
	processorShutdownCh := processor.start(ctx)

	// Selector branch for receiving the event from signal channel and sending it to event processor channel
	selector.AddReceive(eventCh,
		func(c workflow.ReceiveChannel, more bool) {
			// Processing when a new signal is received
			var newEvent ResourceEvent
			c.Receive(ctx, &newEvent)

			// Send new event to the processor
			ch.Send(ctx, newEvent)
		},
	)

	// Pump to continue processing events until the close timer fires
	for !closeWorkflow {
		selector.Select(ctx)
	}

	// Prepare to close this workflow execution

	// First drain all unprocessed signals
	var newEvent ResourceEvent
	for eventCh.ReceiveAsync(&newEvent) {
		// Send new event to the processor
		ch.Send(ctx, newEvent)
	}

	// Close event processor
	ch.Close()

	// Wait for processor to completely shutdown and return state
	var processorState *EventProcessorState
	processorShutdownCh.Receive(ctx, &processorState)

	// This is a good place to store the Resource State back to long term storage

	// Drain all signals which came during processor shutdown
	for eventCh.ReceiveAsync(&newEvent) {
		processorState.addEvent(&newEvent, logger)
	}

	// Check if state has unprocessed events so we can start new run
	if len(processorState.Items) > 0 {
		p := NextRunParams{
			PreviousEventID: processorState.PreviousEventID,
			Items:           processorState.Items,
		}

		return workflow.NewContinueAsNewError(ctx, ResourceWorkflow, p)
	}

	// Complete workflow
	return nil
}

// ProcessEvent is activity which simulates processing of the event
func ProcessEvent(ctx context.Context, event ResourceEvent) error {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)

	logger.Info("Starting to process event",
		"ResourceID", info.WorkflowExecution.ID,
		"EventID", event.EventID)

	sleep := time.Duration(rand.Intn(10))
	time.Sleep(sleep * time.Second)

	logger.Info("Finished processing of event",
		"ResourceID", info.WorkflowExecution.ID,
		"EventID", event.EventID)

	return nil
}
