package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/helloworld"
	"github.com/temporalio/samples-go/serverjwtauth"
)

func main() {
	key, jwk, err := serverjwtauth.ReadKey()
	if err != nil {
		log.Fatalln(err)
	}
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HeadersProvider: &serverjwtauth.JWTHeadersProvider{
			Config: serverjwtauth.JWTConfig{
				Key:   key,
				KeyID: jwk.KeyID,
				Permissions: []string{
					"default:read",
					"default:write",
				},
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "server-jwt-auth", worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
