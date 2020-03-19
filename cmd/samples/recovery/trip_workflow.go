package main

import (
	"encoding/json"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
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

	// ApplicationName is the task list for this sample
	ApplicationName = "recoveryGroup"

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
		zap.String("User", workflowID),
		zap.Int("TripCounter", state.TripCounter))

	// Register query handler to return trip count
	err := workflow.SetQueryHandler(ctx, QueryName, func(input []byte) (int, error) {
		return state.TripCounter, nil
	})

	if err != nil {
		logger.Info("SetQueryHandler failed.", zap.Error(err))
		return err
	}

	// TripCh to wait on trip completed event signals
	tripCh := workflow.GetSignalChannel(ctx, TripSignalName)
	for i := 0; i < 10; i++ {
		var trip TripEvent
		tripCh.Receive(ctx, &trip)
		logger.Info("Trip complete event received.", zap.String("ID", trip.ID), zap.Int("Total", trip.Total))
		state.TripCounter++
	}

	logger.Info("Starting a new run.", zap.Int("TripCounter", state.TripCounter))
	return workflow.NewContinueAsNewError(ctx, "TripWorkflow", state)
}

func deserializeUserState(data []byte) (UserState, error) {
	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return UserState{}, err
	}

	return state, nil
}

func deserializeTripEvent(data []byte) (TripEvent, error) {
	var trip TripEvent
	if err := json.Unmarshal(data, &trip); err != nil {
		return TripEvent{}, err
	}

	return trip, nil
}
