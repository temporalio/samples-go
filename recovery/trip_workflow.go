package recovery

import (
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type (
	// UserState kept within workflow and passed from one run to another on ContinueAsNew
	UserState struct {
		TripCounter int
	}

	// TripEvent passed in as signal to TripWorkflow
	TripEvent struct {
		ID    string
		Total int
	}
)

const (
	// TripSignalName is the signal name for trip completion event
	TripSignalName = "trip_event"

	// QueryName is the query type name
	QueryName = "counter"
)

// TripWorkflow to keep track of total trip count for a user
// It waits on a TripEvent signal and increments a counter on each signal received by this workflow
// Trip count is managed as workflow state and passed to new run after 10 signals received by each execution
func TripWorkflow(ctx workflow.Context, state UserState) error {
	logger := workflow.GetLogger(ctx)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger.Info("Trip Workflow Started for User.",
		"User", workflowID,
		"TripCounter", state.TripCounter)

	// Register query handler to return trip count
	err := workflow.SetQueryHandler(ctx, QueryName, func(input []byte) (int, error) {
		return state.TripCounter, nil
	})

	if err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	// TripCh to wait on trip completed event signals
	tripCh := workflow.GetSignalChannel(ctx, TripSignalName)
	for i := 0; i < 10; i++ {
		var trip TripEvent
		tripCh.Receive(ctx, &trip)
		logger.Info("Trip complete event received.", "ID", trip.ID, "Total", trip.Total)
		state.TripCounter++
	}

	logger.Info("Starting a new run.", "TripCounter", state.TripCounter)
	return workflow.NewContinueAsNewError(ctx, "TripWorkflow", state)
}

func deserializeUserState(data *commonpb.Payloads) (UserState, error) {
	var state UserState
	if err := converter.GetDefaultDataConverter().FromPayloads(data, &state); err != nil {
		return UserState{}, err
	}

	return state, nil
}

func deserializeTripEvent(data *commonpb.Payloads) (TripEvent, error) {
	var trip TripEvent
	if err := converter.GetDefaultDataConverter().FromPayloads(data, &trip); err != nil {
		return TripEvent{}, err
	}

	return trip, nil
}
