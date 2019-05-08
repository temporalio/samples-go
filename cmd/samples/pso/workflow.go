package main

import (
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

// This is registration process where you register all your workflow handlers.
func init() {
	//workflow.Register(FitnessEvaluationWorkflow)
	workflow.Register(PSOWorkflow)
}

//PSOWorkflow workflow decider
func PSOWorkflow(ctx workflow.Context, functionName string) (err error) {
	ao := workflow.ActivityOptions{
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
	ctx = workflow.WithActivityOptions(ctx, ao)

	workflow.GetLogger(ctx).Info(fmt.Sprintf("Optimizing function %s", functionName))
	settings := PSODefaultSettings(functionName)

	// Retry
	for i := 1; i < 5; i++ {
		swarm := NewSwarm(ctx, settings)
		result, err := swarm.Run()
		if err != nil {
			break
		}
		if result.Position.Fitness < settings.Function.Goal {
			msg := fmt.Sprintf("Yay! Goal was reached @ step %d (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
			workflow.GetLogger(ctx).Info(msg)
			break
		} else {
			msg := fmt.Sprintf("Goal was not reached after %d steps (fitness=%.2e) :-)", result.Step, result.Position.Fitness)
			workflow.GetLogger(ctx).Info(msg)
		}
	}

	if err != nil {
		workflow.GetLogger(ctx).Error("Optimzation failed. ", zap.Error(err))
	} else {
		workflow.GetLogger(ctx).Info("Optimization was successful")
	}
	return err
}

//FitnessEvaluationWorkflow workflow decider
// func FitnessEvaluationWorkflow(ctx workflow.Context, function string) (err error) {
// 	ao := workflow.ActivityOptions{
// 		ScheduleToStartTimeout: time.Second * 5,
// 		StartToCloseTimeout:    time.Minute,
// 		HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
// 		RetryPolicy: &cadence.RetryPolicy{
// 			InitialInterval:          time.Second,
// 			BackoffCoefficient:       2.0,
// 			MaximumInterval:          time.Minute,
// 			ExpirationInterval:       time.Minute * 10,
// 			NonRetriableErrorReasons: []string{"bad-error"},
// 		},
// 	}
// 	ctx = workflow.WithActivityOptions(ctx, ao)

// 	return err
// }

// func EvaluateFitness(ctx workflow.Context, ind individual) (err error) {
// 	// Evaluate fitness of the individual
// 	// For the first try, let's evaluate a hard-coded function

// 	var fInfo *fileInfo
// 	err = workflow.ExecuteActivity(ctx, evaluateFitnessActivityName, fileID).Get(ctx, &fInfo)
// 	if err != nil {
// 		return err
// 	}

// 	// return err
// }
