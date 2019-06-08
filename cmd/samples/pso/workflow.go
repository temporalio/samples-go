package main

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"

	"github.com/pborman/uuid"
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

// HostID - Use a new uuid just for demo so we can run 2 host specific activity workers on same machine.
// In real world case, you would use a hostname or ip address as HostID.
var HostID = ApplicationName + "_" + uuid.New()

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

var QueryResult string

// This is registration process where you register all your workflow handlers.
func init() {
	//workflow.Register(FitnessEvaluationWorkflow)
	workflow.Register(PSOWorkflow)
	workflow.Register(PSOChildWorkflow)
}

//PSOWorkflow workflow decider
func PSOWorkflow(ctx workflow.Context, functionName string) (err error) {
	logger := workflow.GetLogger(ctx)

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	// Set child workflow options
	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
	childID := fmt.Sprintf("PSO_child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want cadence to generate a unique ID for child execution
		WorkflowID:                   childID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	// setup query handler for query type "state"
	QueryResult = "started"
	err = workflow.SetQueryHandler(ctx, "state", func(input []byte) (string, error) {
		return QueryResult, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	logger.Info(fmt.Sprintf("Optimizing function %s", functionName))

	settings := PSODefaultSettings(functionName)

	// Retry with different random seed
	const NumberOfAttempts = 5
	for i := 1; i < NumberOfAttempts; i++ {
		swarm, err := NewSwarm(ctx, settings)
		if err != nil {
			logger.Error("Optimization failed. ", zap.Error(err))
			return err
		}

		QueryResult = "initialized"

		var goalReached bool
		err = workflow.ExecuteChildWorkflow(ctx, PSOChildWorkflow, *swarm, 1).Get(ctx, &goalReached)
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
func PSOChildWorkflow(ctx workflow.Context, swarm Swarm, startingStep int) (bool, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Child workflow execution started.")

	// Set activity options
	ctx = workflow.WithActivityOptions(ctx, ActivityOptions)

	result, err := swarm.Run(ctx, startingStep)
	if err != nil {
		if err.Error() == "CONTINUEASNEW" {
			QueryResult = "ContinueAsNew issued"
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
