package main

import (
	greeting "github.com/temporalio/samples-go/lambda-worker/greeting"

	lambdaworker "go.temporal.io/sdk/contrib/aws/lambdaworker"
	otel "go.temporal.io/sdk/contrib/aws/lambdaworker/otel"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	lambdaworker.RunWorker(worker.WorkerDeploymentVersion{
		DeploymentName: "my-app",
		BuildID:        "build-1",
	}, func(opts *lambdaworker.Options) error {
		opts.TaskQueue = "serverless-task-queue-1"

		if err := otel.ApplyDefaults(opts, &opts.ClientOptions, otel.Options{}); err != nil {
			return err
		}

		opts.RegisterWorkflowWithOptions(greeting.SampleWorkflow, workflow.RegisterOptions{
			VersioningBehavior: workflow.VersioningBehaviorAutoUpgrade,
		})
		opts.RegisterActivity(greeting.HelloActivity)

		return nil
	})
}
