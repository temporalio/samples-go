package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/schedule"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)


func main() {
	ctx := context.Background()
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
// @@@SNIPSTART samples-go-schedule-create-delete
	// This schedule ID can be user business logic identifier as well.
	scheduleID := "schedule_" + uuid.New()
	workflowID := "schedule_workflow_" + uuid.New()
	// Create the schedule, start with no spec so the schedule will not run.
	scheduleHandle, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID:   scheduleID,
		Spec: client.ScheduleSpec{},
		Action: &client.ScheduleWorkflowAction{
			ID:        workflowID,
			Workflow:  schedule.SampleScheduleWorkflow,
			TaskQueue: "schedule",
		},
	})
	if err != nil {
		log.Fatalln("Unable to create schedule", err)
	}
	// Delete the schedule once the sample is done
	defer func() {
		log.Println("Deleting schedule", "ScheduleID", scheduleHandle.GetID())
		err = scheduleHandle.Delete(ctx)
		if err != nil {
			log.Fatalln("Unable to delete schedule", err)
		}
	}()
// @@@SNIPEND
// @@@SNIPSTART samples-go-schedule-trigger
	// Manually trigger the schedule once
	log.Println("Manually triggering schedule", "ScheduleID", scheduleHandle.GetID())

	err = scheduleHandle.Trigger(ctx, client.ScheduleTriggerOptions{
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_ALLOW_ALL,
	})
	if err != nil {
		log.Fatalln("Unable to trigger schedule", err)
	}
// @@@SNIPEND
// @@@SNIPSTART samples-go-schedule-update
	// Update the schedule with a spec so it will run periodically,
	log.Println("Updating schedule", "ScheduleID", scheduleHandle.GetID())
	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{
				// Run the schedule at 5pm on Friday
				Calendars: []client.ScheduleCalendarSpec{
					{
						Hour: []client.ScheduleRange{
							{
								Start: 17,
							},
						},
						DayOfWeek: []client.ScheduleRange{
							{
								Start: 5,
							},
						},
					},
				},
				// Run the schedule every 5s
				Intervals: []client.ScheduleIntervalSpec{
					{
						Every: 5 * time.Second,
					},
				},
			}
			// Start the schedule paused to demonstrate how to unpause a schedule
			schedule.Description.Schedule.State.Paused = true
			schedule.Description.Schedule.State.LimitedActions = true
			schedule.Description.Schedule.State.RemainingActions = 10

			return &client.ScheduleUpdate{
				Schedule: &schedule.Description.Schedule,
			}, nil
		},
	})
	if err != nil {
		log.Fatalln("Unable to update schedule", err)
	}
// @@@SNIPEND
// @@@SNIPSTART samples-go-schedule-unpause-describe
	// Unpause schedule
	log.Println("Unpausing schedule", "ScheduleID", scheduleHandle.GetID())
	err = scheduleHandle.Unpause(ctx, client.ScheduleUnpauseOptions{})
	if err != nil {
		log.Fatalln("Unable to unpause schedule", err)
	}
	// Wait for the schedule to run 10 actions
	log.Println("Waiting for schedule to complete 10 actions", "ScheduleID", scheduleHandle.GetID())

	for {
		description, err := scheduleHandle.Describe(ctx)
		if err != nil {
			log.Fatalln("Unable to describe schedule", err)
		}
		if description.Schedule.State.RemainingActions != 0 {
			log.Println("Schedule has remaining actions", "ScheduleID", scheduleHandle.GetID(), "RemainingActions", description.Schedule.State.RemainingActions)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
}
// @@@SNIPEND
