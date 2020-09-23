package main

import (
	"log"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/pso"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: pso.NewJSONDataConverter(),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "pso", worker.Options{
		MaxConcurrentActivityExecutionSize: 1, // Activities are supposed to be CPU intensive, so better limit the concurrency
	})

	w.RegisterWorkflow(pso.PSOWorkflow)
	w.RegisterWorkflow(pso.PSOChildWorkflow)

	w.RegisterActivityWithOptions(pso.InitParticleActivity, activity.RegisterOptions{Name: pso.InitParticleActivityName})
	w.RegisterActivityWithOptions(pso.UpdateParticleActivity, activity.RegisterOptions{Name: pso.UpdateParticleActivityName})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
