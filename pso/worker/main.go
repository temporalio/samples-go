package main

import (
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/pso"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: pso.NewJSONDataConverter(),
		Logger:        logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	w := worker.New(c, "pso", worker.Options{
		MaxConcurrentActivityExecutionSize: 1, // Activities are supposed to be CPU intensive, so better limit the concurrency
	})

	w.RegisterWorkflow(pso.PSOWorkflow)
	w.RegisterWorkflow(pso.PSOChildWorkflow)

	w.RegisterActivityWithOptions(pso.InitParticleActivity, activity.RegisterOptions{Name: pso.InitParticleActivityName})
	w.RegisterActivityWithOptions(pso.UpdateParticleActivity, activity.RegisterOptions{Name: pso.UpdateParticleActivityName})

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
