package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/pborman/uuid"
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

	// Setup query handler for query type "child"
	var childWorkflowID string
	err = workflow.SetQueryHandler(ctx, "child", func(input []byte) (string, error) {
		return childWorkflowID, nil
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

		// Set child workflow options
		// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   "PSO_Child_" + uuid.New(),
			ExecutionStartToCloseTimeout: time.Minute,
		}
		ctx = workflow.WithChildOptions(ctx, cwo)

		var goalReached bool
		childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, PSOChildWorkflow, *swarm, 1)
		var childWE workflow.Execution
		childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE)
		childWorkflowID = childWE.ID
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

	// Run real optimization loop
	result, err := swarm.Run(ctx, startingStep)
	if err != nil {
		if err.Error() == ContinueAsNewStr {
			return false, workflow.NewContinueAsNewError(ctx, PSOChildWorkflow, swarm, result.Step+1)
		}

		logger.Error("Error in swarm loop: ", zap.Error(err))
		return false, errors.New("Error in swarm loop")
	}
	if result.Position.Fitness < swarm.Settings.function.Goal {
		msg := fmt.Sprintf("Yay! Goal was reached @ step %d (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
		logger.Info(msg)
		return true, nil
	}

	msg := fmt.Sprintf("Goal was not reached after %d steps (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
	logger.Info(msg)
	return false, nil
}
