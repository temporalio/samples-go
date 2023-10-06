//go:build go1.21

package main

import (
	"log"
	"os"

	"log/slog"

	"github.com/temporalio/samples-go/slogadapter"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{
		Logger: tlog.NewStructuredLogger(
			slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}))),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "slog-logger", worker.Options{})

	w.RegisterWorkflow(slogadapter.Workflow)
	w.RegisterActivity(slogadapter.LoggingActivity)
	w.RegisterActivity(slogadapter.LoggingErrorAcctivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
