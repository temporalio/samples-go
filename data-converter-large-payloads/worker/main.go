package main

import (
	"log"

	dataconverterlargepayloads "github.com/temporalio/samples-go/data-converter-large-payloads"
	pc "github.com/temporalio/samples-go/data-converter-large-payloads/payloadconverter"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

func main() {

	dataconverter := converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pc.NewLargeSizePayloadConverter(),
		// fallback converter for payloads that do not exceed the threshold size
		converter.NewJSONPayloadConverter(),
	)

	c, err := client.Dial(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: dataconverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "data-converter-large-payloads", worker.Options{})

	w.RegisterWorkflow(dataconverterlargepayloads.Workflow)
	w.RegisterActivity(dataconverterlargepayloads.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
