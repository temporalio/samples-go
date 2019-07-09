package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence"

	"go.uber.org/cadence/workflow"
)

type WorkflowResult struct {
	Msg     string // Uppercase the members otherwise serialization won't work!
	Success bool
}

// ApplicationName is the task list for this sample
const ApplicationName = "PSO"

// ActivityOptions can be reused
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
	workflow.Register(PSOWorkflow)
	workflow.Register(PSOChildWorkflow)
}

//PSOWorkflow workflow decider
func PSOWorkflow(ctx workflow.Context, functionName string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Optimizing function %s", functionName))

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	// Setup query handler for query type "child"
	var childWorkflowID string
	err := workflow.SetQueryHandler(ctx, "child", func(input []byte) (string, error) {
		return childWorkflowID, nil
	})
	if err != nil {
		msg := fmt.Sprintf("SetQueryHandler failed: " + err.Error())
		logger.Error(msg)
		return msg, err
	}

	// Retry with different random seed
	settings := PSODefaultSettings(functionName)
	const NumberOfAttempts = 5
	for i := 1; i < NumberOfAttempts; i++ {
		logger.Info(fmt.Sprintf("Attempt #%d", i))

		swarm, err := NewSwarm(ctx, settings)
		if err != nil {
			msg := fmt.Sprintf("Optimization failed. " + err.Error())
			logger.Error(msg)
			return msg, err
		}

		// Set child workflow options
		// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   "PSO_Child_" + uuid.New(),
			ExecutionStartToCloseTimeout: time.Minute,
		}
		ctx = workflow.WithChildOptions(ctx, cwo)

		childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, PSOChildWorkflow, *swarm, 1)
		var childWE workflow.Execution
		childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE)
		childWorkflowID = childWE.ID
		var result WorkflowResult
		err = childWorkflowFuture.Get(ctx, &result) // This blocking until the child workflow has finished
		if err != nil {
			msg := fmt.Sprintf("Parent execution received child execution failure. " + err.Error())
			logger.Error(msg)
			return msg, err
		}
		if result.Success {
			msg := fmt.Sprintf("Optimization was successful at attempt #%d. %s", i, result.Msg)
			logger.Info(msg)
			return msg, nil
		}
	}

	msg := fmt.Sprintf("Unable to reach goal after %d attempts", NumberOfAttempts)
	logger.Info(msg)
	return msg, nil
}

// PSOChildWorkflow workflow decider
// Returns true if the optimization has converged
func PSOChildWorkflow(ctx workflow.Context, swarm Swarm, startingStep int) (WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Child workflow execution started.")

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	// Run real optimization loop
	result, err := swarm.Run(ctx, startingStep)
	if err != nil {
		if err.Error() == ContinueAsNewStr {
			return WorkflowResult{"NewContinueAsNewError", false}, workflow.NewContinueAsNewError(ctx, PSOChildWorkflow, swarm, result.Step+1)
		}

		msg := fmt.Sprintf("Error in swarm loop: " + err.Error())
		logger.Error(msg)
		return WorkflowResult{msg, false}, errors.New("Error in swarm loop")
	}
	if result.Position.Fitness < swarm.Settings.function.Goal {
		msg := fmt.Sprintf("Yay! Goal was reached @ step %d (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
		logger.Info(msg)
		return WorkflowResult{msg, true}, nil
	}

	msg := fmt.Sprintf("Goal was not reached after %d steps (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
	logger.Info(msg)
	return WorkflowResult{msg, false}, nil
}
