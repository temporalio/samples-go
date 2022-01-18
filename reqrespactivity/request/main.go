package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/temporalio/samples-go/reqrespactivity"
	"go.temporal.io/sdk/client"
)

func main() {
	var opts reqrespactivity.RequesterOptions
	flag.StringVar(&opts.TargetWorkflowID, "w", "reqrespactivity_workflow", "WorkflowID")
	flag.Parse()

	// Create client
	var err error
	opts.Client, err = client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer opts.Client.Close()

	// Create requester
	req, err := reqrespactivity.NewRequester(opts)
	if err != nil {
		log.Fatalln("Unable to create requester", err)
	}
	defer req.Close()

	// Run until ctrl+c
	log.Printf("Requesting every second until ctrl+c")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()

	// Request every second
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for i := 0; ; i++ {
		str := "foo" + strconv.Itoa(i)
		log.Printf("Requesting %q be uppercased", str)
		if val, err := req.RequestUppercase(ctx, str); err != nil {
			log.Printf("  Failed: %v", err)
		} else {
			log.Printf("  Result: %q", val)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
