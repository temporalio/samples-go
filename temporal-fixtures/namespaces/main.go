package main

import (
	"context"

	// "fmt"
	"log"
	"strconv"

	"github.com/pborman/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

var (
	NumberOfNamespaces = 20
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewNamespaceClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	uuidvar := uuid.New()[:6]
	i := 1
	for i <= NumberOfNamespaces {
		req := &workflowservice.RegisterNamespaceRequest{
			Namespace:                        uuidvar + "_" + strconv.Itoa(i),
			Description:                      "Namespace Description " + strconv.Itoa(i),
			OwnerEmail:                       "owner@mail.com",
			WorkflowExecutionRetentionPeriod: durationpb.New(24 * time.Hour),
		}
		if err = c.Register(context.Background(), req); err != nil {
			log.Fatalln("Unable to register namespace", err)
		}
		i++
	}
}
