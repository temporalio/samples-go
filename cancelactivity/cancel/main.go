package main

import (
	"context"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var workflowID string
	flag.StringVar(&workflowID, "w", "workflowID-to-cancel", "w is the workflowID of the workflow to be canceled.")
	flag.Parse()

	if workflowID == "" {
		flag.PrintDefaults()
		return
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	err = c.CancelWorkflow(context.Background(), workflowID, "")
	if err != nil {
		logger.Fatal("Unable to cancel workflow", zap.Error(err))
	} else {
		logger.Info("Workflow cancelled", zap.String("WorkflowID", workflowID))
	}

	// Close connection, clean up resources.
	_ = c.CloseConnection()
}
