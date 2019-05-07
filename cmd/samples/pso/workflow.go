package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence"

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

// This is registration process where you register all your workflow handlers.
func init() {
	//workflow.Register(FitnessEvaluationWorkflow)
	workflow.Register(PSOWorkflow)

}

func PSOWorkflow(ctx workflow.Context, functionName string) (err error) {
	// step 1: download resource file
	// step 1: download resource file
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Minute * 10,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// fmt.Printf("Optimizing function %s\n", *function)
	settings := PSODefaultSettings()
	// settings.PrintEvery = *printEvery
	switch functionName {
	case "sphere":
		settings.Function = Sphere
	case "rosenbrock":
		settings.Function = Rosenbrock
	case "griewank":
		settings.Function = Griewank
	}

	// Retry
	for i := 1; i < 5; i++ {
		//		err = processFile(ctx, fileID)
		swarm := NewSwarm(ctx, settings)
		result := swarm.Run()
		if result.Position.Fitness < settings.Function.Goal {
			fmt.Printf("Yay! Goal was reached @ step %d (fitness=%.2e) :-)",
				result.Step, result.Position.Fitness)
			workflow.GetLogger(ctx).Info("optimization was successful ")
			// zap.int("WorkflowID", result.Step),
			// zap.float64("RunID", result.Position.Fitness))

			err = nil
		} else {
			fmt.Printf("Goal was not reached after %d steps (fitness=%.2e) :-)",
				result.Step, result.Position.Fitness)

		}

		if err == nil {
			break
		}
	}
	if err != nil {
		workflow.GetLogger(ctx).Error("optimzation failed.")
	} else {
		workflow.GetLogger(ctx).Info("optimization was successful.")
	}
	return err
}

//FitnessEvaluation workflow decider
func FitnessEvaluationWorkflow(ctx workflow.Context, function string) (err error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Minute * 10,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	return err
}

// func EvaluateFitness(ctx workflow.Context, ind individual) (err error) {
// 	// Evaluate fitness of the individual
// 	// For the first try, let's evaluate a hard-coded function

// 	var fInfo *fileInfo
// 	err = workflow.ExecuteActivity(ctx, EvaluateFitnessActivityName, fileID).Get(ctx, &fInfo)
// 	if err != nil {
// 		return err
// 	}

// 	// return err
// }
