package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	choice "github.com/temporalio/temporal-go-samples/choice-exclusive"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "choice", worker.Options{})

	w.RegisterWorkflow(choice.ExclusiveChoiceWorkflow)

	orderChoices := []string{
		choice.OrderChoiceApple,
		choice.OrderChoiceBanana,
		choice.OrderChoiceCherry,
		choice.OrderChoiceOrange}
	w.RegisterActivity(&choice.OrderActivities{OrderChoices: orderChoices})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
