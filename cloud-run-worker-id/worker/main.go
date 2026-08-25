// @@@SNIPSTART go-cloud-run-worker-id
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	greeting "github.com/temporalio/samples-go/cloud-run-worker-id/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/gcp/cloudrun"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx := context.Background()

	// Read the Cloud Run instance metadata once, at worker startup. FetchMetadata reads the
	// deployment name and revision from environment variables and fetches the unique instance ID
	// from the GCP metadata server. It performs a network request, so it must never be called from
	// workflow code. It fails when the process is not running on a Cloud Run worker pool or service
	// (for example, when running locally), because the metadata server is unreachable.
	md, err := cloudrun.FetchMetadata(ctx)
	if err != nil {
		log.Fatalf("fetching Cloud Run metadata (is this running on a Cloud Run worker pool or service?): %v", err)
	}
	log.Printf("Cloud Run metadata: name=%q revision=%q instanceID=%q", md.Name, md.Revision, md.InstanceID)

	// Derive the worker identity ("<instanceID>@<revision>") from the metadata and apply it to the
	// client options. A user-provided identity always wins, so applying it is safe. The Temporal
	// connection is read from the standard TEMPORAL_* environment variables; this sample uses a
	// plaintext connection for simplicity.
	clientOptions := client.Options{
		HostPort:  getenv("TEMPORAL_ADDRESS", client.DefaultHostPort),
		Namespace: getenv("TEMPORAL_NAMESPACE", client.DefaultNamespace),
	}
	md.ApplyToClientOptions(&clientOptions)
	log.Printf("Worker identity: %s", clientOptions.Identity)

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Opt the worker into Worker Deployment Versioning using the Cloud Run deployment name and
	// revision (as the build ID), pinning workflows to this version by default. Cloud Run creates a
	// new revision for each deployment, which makes it a natural worker build ID.
	workerOptions := worker.Options{}
	if err := md.ApplyToWorkerOptions(&workerOptions); err != nil {
		log.Fatalf("configuring worker deployment versioning: %v", err)
	}

	taskQueue := getenv("TEMPORAL_TASK_QUEUE", "cloud-run-task-queue")
	w := worker.New(c, taskQueue, workerOptions)
	w.RegisterWorkflow(greeting.SampleWorkflow)
	w.RegisterActivity(greeting.HelloActivity)

	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	log.Printf("Worker started on task queue %q (deployment=%s build=%s, pinned)", taskQueue, md.Name, md.Revision)

	// Cloud Run sends SIGTERM and allows roughly ten seconds before SIGKILL. Stop the worker
	// gracefully so in-flight tasks are released back to the queue.
	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	log.Println("Shutdown signal received, stopping worker")
	w.Stop()
	log.Println("Worker stopped")
}

// getenv returns the value of the environment variable named by key, or fallback if it is unset or
// empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// @@@SNIPEND
