package main

import (
	"os"
	"os/signal"

	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	choice_multi "github.com/temporalio/temporal-go-samples/choice-multi"
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

	w := worker.New(c, "choice-multi", worker.Options{
		Logger: logger,
	})
	defer w.Stop()

	w.RegisterWorkflow(choice_multi.MultiChoiceWorkflow)

	orderChoices := []string{
		choice_multi.OrderChoiceApple,
		choice_multi.OrderChoiceBanana,
		choice_multi.OrderChoiceCherry,
		choice_multi.OrderChoiceOrange}
	w.RegisterActivity(&choice_multi.OrderActivities{OrderChoices: orderChoices})

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
