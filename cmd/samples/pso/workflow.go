package main

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"

	"go.uber.org/cadence/workflow"
)

type (
	individual struct {
		IndividualID string
		Variable     []float64
		Fitness      []float64
	}
)

// ApplicationName is the task list for this sample
const ApplicationName = "PSO"

var ActivityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Second * 5,
	StartToCloseTimeout:    time.Minute * 10,
	HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
	RetryPolicy: &cadence.RetryPolicy{
		InitialInterval:          time.Second,
		BackoffCoefficient:       2.0,
		MaximumInterval:          time.Minute,
		ExpirationInterval:       time.Minute * 10,
		MaximumAttempts:          5,
		NonRetriableErrorReasons: []string{"bad-error"},
	},
}

const ContinueAsNewStr = "CONTINUEASNEW"

const QueryResultName = "QueryResult"

// This is registration process where you register all your workflow handlers.
func init() {
	//workflow.Register(FitnessEvaluationWorkflow)
	workflow.Register(PSOWorkflow)
	workflow.Register(PSOChildWorkflow)
}

//PSOWorkflow workflow decider
func PSOWorkflow(ctx workflow.Context, functionName string) (err error) {
	logger := workflow.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Optimizing function %s", functionName))

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	// Set child workflow options
	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
	childWorkflowID := fmt.Sprintf("PSO_child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want cadence to generate a unique ID for child execution
		WorkflowID:                   childWorkflowID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	// Setup query handler for query type "child"
	ctx = workflow.WithValue(ctx, QueryResultName, childWorkflowID) // Don't need the runID when querying. Use only workflowID and it is going to query the latest one
	err = workflow.SetQueryHandler(ctx, "child", func(input []byte) (string, error) {
		return ctx.Value(QueryResultName).(string), nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	// Retry with different random seed
	settings := PSODefaultSettings(functionName)
	const NumberOfAttempts = 5
	for i := 1; i < NumberOfAttempts; i++ {
		logger.Info(fmt.Sprintf("Attempt #%d", i))

		swarm, err := NewSwarm(ctx, settings)
		if err != nil {
			logger.Error("Optimization failed. ", zap.Error(err))
			return err
		}

		var goalReached bool
		childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, PSOChildWorkflow, *swarm, 1)
		// var childWE workflow.Execution
		// childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE)
		// ctx = workflow.WithValue(ctx, QueryResultName, childWE.RunID)
		err = childWorkflowFuture.Get(ctx, &goalReached) // This blocking until the child workflow has finished
		if err != nil {
			logger.Error("Parent execution received child execution failure.", zap.Error(err))
			return err
		}
		if goalReached {
			logger.Info("Optimization was successful")
			return nil
		}
	}

	logger.Info("Unable to reach goal after " + string(NumberOfAttempts) + " attempts")
	return nil
}

// PSOChildWorkflow workflow decider
// Returns true if the optimization has converged
func PSOChildWorkflow(ctx workflow.Context, swarm Swarm, startingStep int) (bool, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Child workflow execution started.")

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	// Setup query handler for query type "iteration"
	ctx = workflow.WithValue(ctx, QueryResultName, "started")
	err := workflow.SetQueryHandler(ctx, "iteration", func(input []byte) (string, error) {
		return ctx.Value(QueryResultName).(string), nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return false, err
	}

	// Run real optimization loop
	result, err := swarm.Run(ctx, startingStep)
	if err != nil {
		if err.Error() == ContinueAsNewStr {
			ctx = workflow.WithValue(ctx, QueryResultName, "ContinueAsNew issued")
			return false, workflow.NewContinueAsNewError(ctx, PSOChildWorkflow, swarm, result.Step+1)
		}

		logger.Error("Error in swarm loop: ", zap.Error(err))
		return false, errors.New("Error in swarm loop")
	}
	if result.Position.Fitness < swarm.Settings.Function.Goal {
		msg := fmt.Sprintf("Yay! Goal was reached @ step %d (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
		logger.Info(msg)
		return true, nil
	}

	msg := fmt.Sprintf("Goal was not reached after %d steps (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
	logger.Info(msg)
	return false, nil
}
