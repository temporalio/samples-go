package main

import (
	"context"
	"flag"
	"log"

	"github.com/temporalio/samples-go/reqresp"
	"go.temporal.io/sdk/client"
)

func main() {
	var opts reqresp.RequesterOptions
	flag.StringVar(&opts.TargetWorkflowID, "w", "reqresp_workflow", "WorkflowID")
	flag.BoolVar(&opts.UseActivityResponse, "activity-based", false, "Use activity-based responses")
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatalln("At least one argument to uppercase is required")
	}

	var err error
	opts.Client, err = client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer opts.Client.Close()

	// Create requester
	req, err := reqresp.NewRequester(opts)
	if err != nil {
		log.Fatalln("Unable to create requester", err)
	}
	defer req.Close()

	// Send arguments to be uppercased one at a time
	log.Printf("Requesting %v value(s) be uppercased", len(flag.Args()))
	for _, arg := range flag.Args() {
		res, err := req.RequestUppercase(context.Background(), arg)
		if err != nil {
			log.Fatalln("Unable to uppercase request", err)
		}
		log.Printf("Uppercased - from %v to %v", arg, res)
	}
}
