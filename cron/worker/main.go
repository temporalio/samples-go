package main

import (
	"os"
	"os/signal"

	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/cron"
)

var (
	logger *zap.Logger
)

func main() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The workers are supposed to be long running process that should not exit.
	c, w := startWorker()
	waitCtrlC()
	// Stop worker, close connection, clean up resources.
	w.Stop()
	_ = c.CloseConnection()
}

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorker() (client.Client, worker.Worker) {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	workerOptions := worker.Options{
		Logger: logger,
	}

	w := worker.New(c, "cron-task-list", workerOptions)

	w.RegisterWorkflow(cron.SampleCronWorkflow)
	w.RegisterActivity(cron.SampleCronActivity)

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	return c, w
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
