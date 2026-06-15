package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	endpointName  = "nexus-signal-without-temporalnexus-endpoint"
	serviceName   = "signal-demo-service"
	operationName = "signal-workflow"
	signalName    = "demo-signal"
	taskQueue     = "nexus-signal-without-temporalnexus-task-queue"
)

type SignalInput struct {
	WorkflowID string
	Message    string
}

func SignalReceiverWorkflow(ctx workflow.Context) (string, error) {
	ch := workflow.GetSignalChannel(ctx, signalName)

	var message string
	ch.Receive(ctx, &message)
	return message, nil
}

func NexusCallerWorkflow(ctx workflow.Context, input SignalInput) (string, error) {
	c := workflow.NewNexusClient(endpointName, serviceName)
	fut := c.ExecuteOperation(ctx, operationName, input, workflow.NexusOperationOptions{})

	var result string
	if err := fut.Get(ctx, &result); err != nil {
		return "", err
	}
	return result, nil
}

func main() {
	var namespace string
	var hostPort string
	flag.StringVar(&namespace, "namespace", "default", "Temporal namespace")
	flag.StringVar(&hostPort, "target-host", client.DefaultHostPort, "Temporal frontend host:port")
	flag.Parse()

	ctx := context.Background()
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	signalOperation := nexus.NewSyncOperation(operationName, func(ctx context.Context, input SignalInput, _ nexus.StartOperationOptions) (string, error) {
		if err := c.SignalWorkflow(ctx, input.WorkflowID, "", signalName, input.Message); err != nil {
			return "", err
		}
		return fmt.Sprintf("signaled workflow %q", input.WorkflowID), nil
	})

	service := nexus.NewService(serviceName)
	if err := service.Register(signalOperation); err != nil {
		log.Fatalln("Unable to register Nexus operation", err)
	}

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(SignalReceiverWorkflow)
	w.RegisterWorkflow(NexusCallerWorkflow)
	w.RegisterNexusService(service)

	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	defer w.Stop()

	receiverID := fmt.Sprintf("nexus-signal-receiver-%d", time.Now().UnixNano())
	receiverRun, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        receiverID,
		TaskQueue: taskQueue,
	}, SignalReceiverWorkflow)
	if err != nil {
		log.Fatalln("Unable to start receiver workflow", err)
	}

	callerRun, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        fmt.Sprintf("nexus-signal-caller-%d", time.Now().UnixNano()),
		TaskQueue: taskQueue,
	}, NexusCallerWorkflow, SignalInput{
		WorkflowID: receiverID,
		Message:    "signal sent from a raw Nexus operation",
	})
	if err != nil {
		log.Fatalln("Unable to start caller workflow", err)
	}

	var callerResult string
	if err := callerRun.Get(ctx, &callerResult); err != nil {
		log.Fatalln("Caller workflow failed", err)
	}
	log.Println("Caller workflow result:", callerResult)

	var receiverResult string
	if err := receiverRun.Get(ctx, &receiverResult); err != nil {
		log.Fatalln("Receiver workflow failed", err)
	}
	log.Println("Receiver workflow result:", receiverResult)
}
