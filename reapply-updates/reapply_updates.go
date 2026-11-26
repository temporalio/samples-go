package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var WorkflowId = fmt.Sprintf("reapply-updates-%s", uuid.New().String())

func Workflow(ctx workflow.Context) ([]string, error) {
	var updateArgs []string
	workflow.SetUpdateHandler(ctx, "my-update", func(arg string) []string {
		updateArgs = append(updateArgs, arg)
		return updateArgs
	})
	workflow.Await(ctx, func() bool { return len(updateArgs) > 0 })
	return updateArgs, nil
}

func Starter(c client.Client) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                    WorkflowId,
		TaskQueue:             "reapply-updates",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx := context.Background()

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, Workflow)
	if err != nil {
		log.Fatalln("Failed to start workflow", err)
	}

	_, err = c.UpdateWorkflow(ctx, WorkflowId, we.GetRunID(), "my-update", "arg1")
	if err != nil {
		log.Fatalln("Failed to execute update")
	}

	var result1 []string
	err = we.Get(ctx, &result1)
	if err != nil {
		log.Fatalln("Failed to obtain workflow result", err)
	}
	fmt.Println("workflow result", result1)

	newRunId, err := resetWorkflow(we.GetRunID(), 4, c)
	if err != nil {
		log.Fatalln("Failed to reset workflow", err)
	}
	fmt.Printf("did reset: http://localhost:8080/namespaces/default/workflows/%s/%s", WorkflowId, newRunId)

	newHandle := c.GetWorkflow(ctx, WorkflowId, newRunId)
	var result2 []string
	err = newHandle.Get(ctx, &result2)
	if err != nil {
		log.Fatalln("Failed to obtain workflow result after reset", err)
	}
	fmt.Println("workflow result after reset", result2)

}

func resetWorkflow(runId string, eventId int64, client client.Client) (string, error) {
	resp, err := client.ResetWorkflowExecution(context.Background(), &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: "default",
		WorkflowExecution: &common.WorkflowExecution{
			WorkflowId: WorkflowId,
			RunId:      runId,
		},
		Reason:                    "Reset to test update reapply",
		RequestId:                 "1",
		ResetReapplyType:          enums.RESET_REAPPLY_TYPE_UNSPECIFIED, // TODO
		WorkflowTaskFinishEventId: eventId,                              // First WFTCompleted
	})
	if err != nil {
		log.Fatalln("Failed to reset workflow", err)
	}
	return resp.RunId, nil
}

func Worker(c client.Client) {
	w := worker.New(c, "reapply-updates", worker.Options{})
	w.RegisterWorkflow(Workflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func main() {
	c, err := client.Dial(client.Options{Logger: noopLogger{}})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	go Worker(c)
	Starter(c)
}

type noopLogger struct{}

func (l noopLogger) Debug(string, ...interface{}) {}
func (l noopLogger) Info(string, ...interface{})  {}
func (l noopLogger) Warn(string, ...interface{})  {}
func (l noopLogger) Error(string, ...interface{}) {}
