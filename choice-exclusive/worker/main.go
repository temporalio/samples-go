package main

import (
	"os"
	"os/signal"

	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	choice "github.com/temporalio/temporal-go-samples/choice-exclusive"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer func() { _ = c.CloseConnection() }()

	w := worker.New(c, "choice", worker.Options{
		Logger: logger,
	})
	defer w.Stop()

	w.RegisterWorkflow(choice.ExclusiveChoiceWorkflow)

	orderChoices := []string{
		choice.OrderChoiceApple,
		choice.OrderChoiceBanana,
		choice.OrderChoiceCherry,
		choice.OrderChoiceOrange}
	w.RegisterActivity(&choice.OrderActivities{OrderChoices: orderChoices})

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	// The workers are supposed to be long running process that should not exit.
	waitCtrlC()
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
