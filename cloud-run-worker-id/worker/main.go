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

	// Register the Cloud Run plugin on the client. When the client connects, the plugin reads the
	// Cloud Run instance metadata once — the deployment name and revision from environment
	// variables and the unique instance ID from the GCP metadata server — then sets the derived
	// worker identity ("<instanceID>@<revision>", unless one is already set) on the client and opts
	// every worker created from the client into Worker Deployment Versioning, pinning workflows to
	// this version by default. The metadata fetch performs a network request, so it never runs from
	// workflow code. It fails fast when the process is not running on a Cloud Run worker pool or
	// service (for example, when running locally), because the metadata server is unreachable.
	//
	// The Temporal connection is read from the standard TEMPORAL_* environment variables; this
	// sample uses a plaintext connection for simplicity.
	plugin := cloudrun.NewPlugin(cloudrun.PluginOptions{})
	clientOptions := client.Options{
		HostPort:  getenv("TEMPORAL_ADDRESS", client.DefaultHostPort),
		Namespace: getenv("TEMPORAL_NAMESPACE", client.DefaultNamespace),
		Plugins:   []client.Plugin{plugin},
	}

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("Unable to create Temporal client (is this running on a Cloud Run worker pool or service?): %v", err)
	}
	defer c.Close()

	// The plugin fetched and cached the metadata during Dial. Read it back for logging.
	md := plugin.Metadata()
	log.Printf("Cloud Run metadata: name=%q revision=%q instanceID=%q", md.Name, md.Revision, md.InstanceID)
	log.Printf("Worker identity: %s", md.WorkerIdentity())

	// The plugin sets the worker's DeploymentOptions (deployment name + revision as the build ID,
	// pinned) automatically, so a zero worker.Options is all that is needed here. Cloud Run creates
	// a new revision for each deployment, which makes it a natural worker build ID.
	taskQueue := getenv("TEMPORAL_TASK_QUEUE", "cloud-run-task-queue")
	w := worker.New(c, taskQueue, worker.Options{})
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
