package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	w := worker.New(c, "choice-multi", worker.Options{})

	w.RegisterWorkflow(choice_multi.MultiChoiceWorkflow)

	orderChoices := []string{
		choice_multi.OrderChoiceApple,
		choice_multi.OrderChoiceBanana,
		choice_multi.OrderChoiceCherry,
		choice_multi.OrderChoiceOrange}
	w.RegisterActivity(&choice_multi.OrderActivities{OrderChoices: orderChoices})

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
