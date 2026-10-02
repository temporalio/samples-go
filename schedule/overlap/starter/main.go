package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/temporalio/samples-go/schedule/overlap"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

const (
	interval             = 2 * time.Second
	defaultCatchupSeconds = 10
)

type policyChoice struct {
	name  string
	value enumspb.ScheduleOverlapPolicy
}

var policies = map[string]policyChoice{
	"skip":            {"skip", enumspb.SCHEDULE_OVERLAP_POLICY_SKIP},
	"buffer-one":      {"buffer-one", enumspb.SCHEDULE_OVERLAP_POLICY_BUFFER_ONE},
	"buffer-all":      {"buffer-all", enumspb.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL},
	"cancel-other":    {"cancel-other", enumspb.SCHEDULE_OVERLAP_POLICY_CANCEL_OTHER},
	"terminate-other": {"terminate-other", enumspb.SCHEDULE_OVERLAP_POLICY_TERMINATE_OTHER},
	"allow-all":       {"allow-all", enumspb.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL},
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		usage()
	}
	name := strings.ToLower(strings.ReplaceAll(os.Args[1], "_", "-"))
	choice, ok := policies[name]
	if !ok {
		usage()
	}
	catchupSeconds := defaultCatchupSeconds
	if len(os.Args) == 3 {
		var err error
		catchupSeconds, err = strconv.Atoi(os.Args[2])
		if err != nil || catchupSeconds < 10 {
			log.Fatalln("catchup-seconds must be an integer of at least 10")
		}
	}
	catchupWindow := time.Duration(catchupSeconds) * time.Second

	ctx := context.Background()
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	scheduleID := "schedule-overlap-" + choice.name
	handle, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{Intervals: []client.ScheduleIntervalSpec{{
			Every: interval,
		}}},
		Action: &client.ScheduleWorkflowAction{
			ID:        scheduleID + "-workflow",
			Workflow:  overlap.PolicyWorkflow,
			Args:      []interface{}{choice.name},
			TaskQueue: overlap.TaskQueue,
		},
		Overlap:       choice.value,
		CatchupWindow: catchupWindow,
		Note:          "Overlap Policy demo: " + choice.name,
	})
	if err != nil {
		log.Fatalf("Unable to create %q: %v\nDelete an existing demo with: temporal schedule delete --schedule-id %s", scheduleID, err, scheduleID)
	}

	fmt.Printf("Created %q with policy %s (interval %s, Workflow duration %s, Catchup Window %s).\n", handle.GetID(), choice.value, interval, overlap.RunDuration, catchupWindow)
	fmt.Printf("Describe it: temporal schedule describe --schedule-id %s\n", scheduleID)
	fmt.Printf("Delete it:   temporal schedule delete --schedule-id %s\n", scheduleID)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go run ./schedule/overlap/starter <skip|buffer-one|buffer-all|cancel-other|terminate-other|allow-all> [catchup-seconds]\n")
	os.Exit(2)
}
