package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/serializer"
)

type (
	ResourceUpdates struct {
		ID      string
		Updates []SingleUpdate
	}

	SingleUpdate struct {
		EventID int
	}
)

var (
	NumberOfResouces   int = 1
	MaxNumberOfUpdates int = 10
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})

	if err != nil {
		log.Fatal("Unable to create client")
	}
	defer c.Close()

	all := generateResourceUpdates()
	for len(all) > 0 {
		index := rand.Intn(len(all))
		r := all[index]
		lastItem := r.Updates[len(r.Updates)-1]
		r.Updates = r.Updates[:len(r.Updates)-1]
		if len(r.Updates) == 0 {
			all = removeResource(all, index)
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        r.ID,
			TaskQueue: serializer.Task_Queue_Name,
		}
		run, err := c.SignalWithStartWorkflow(context.Background(), r.ID, serializer.Resource_Event_Signal_Name,
			serializer.ResourceEvent{EventID: lastItem.EventID}, workflowOptions, serializer.ResourceWorkflow, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Resource signaled.  ID: %v\n", run.GetID())
	}
}

func generateResourceUpdates() []*ResourceUpdates {
	allResources := []*ResourceUpdates{}
	for i := 1; i <= NumberOfResouces; i++ {
		name := fmt.Sprintf("Resource_%v", i)
		updates := &ResourceUpdates{
			ID: name,
		}

		numUpdates := rand.Intn(MaxNumberOfUpdates) + 1
		for j := 1; j <= numUpdates; j++ {
			updates.Updates = append(updates.Updates, SingleUpdate{
				EventID: j,
			})
		}

		rand.Shuffle(len(updates.Updates), func(i, j int) {
			updates.Updates[i], updates.Updates[j] = updates.Updates[j], updates.Updates[i]
		})

		allResources = append(allResources, updates)
	}

	return allResources
}

func removeResource(u []*ResourceUpdates, index int) []*ResourceUpdates {
	u[len(u)-1], u[index] = u[index], u[len(u)-1]
	return u[:len(u)-1]
}
