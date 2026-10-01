// @@@SNIPSTART go-cloud-run-id
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	greeting "github.com/temporalio/samples-go/gcp/cloudrun/id/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/gcp/cloudrun/id"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx := context.Background()

	// The plugin reads Cloud Run instance metadata during Dial and sets the client identity to
	// "<instanceID>@<revision>". It requires the GCP metadata server, so the worker must run on
	// Cloud Run rather than locally.
	plugin := id.NewCloudRunIDPlugin()
	c, err := client.Dial(client.Options{
		HostPort:  getenv("TEMPORAL_ADDRESS", client.DefaultHostPort),
		Namespace: getenv("TEMPORAL_NAMESPACE", client.DefaultNamespace),
		Plugins:   []client.Plugin{plugin},
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client (is this running on Cloud Run?): %v", err)
	}
	defer c.Close()

	log.Printf("Client identity: %s", plugin.Metadata().Identity())

	taskQueue := getenv("TEMPORAL_TASK_QUEUE", "cloud-run-task-queue")
	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(greeting.SampleWorkflow)
	w.RegisterActivity(greeting.HelloActivity)

	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	log.Printf("Worker started on task queue %q", taskQueue)

	// Cloud Run sends SIGTERM and allows roughly ten seconds before SIGKILL; stop gracefully so
	// in-flight tasks are released back to the queue.
	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	log.Println("Shutdown signal received, stopping worker")
	w.Stop()
	log.Println("Worker stopped")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// @@@SNIPEND
