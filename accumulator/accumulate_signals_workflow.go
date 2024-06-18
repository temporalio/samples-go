package accumulator

import (
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
)

/**
 * This sample demonstrates how to accumulate many signals (events) over a time period.
 * This sample implements the Accumulator Pattern: collect many meaningful things that
 *   need to be collected and worked on together, such as all payments for an account, or
 *   all account updates by account.
 * This sample models robots being created throughout the time period,
 *   groups them by what color they are, and greets all the robots of a color at the end.
 *
 * A new workflow is created per grouping. Workflows continue as new as needed.
 * A sample activity at the end is given, and you could add an activity to
 *   process individual events in the processGreeting() method.
 *
 * Because Temporal Workflows cannot have an unlimited size, Continue As New is used
 *   to process more signals that may come in.
 * You could create as many groupings as desired, as Temporal Workflows scale out to many workflows without limit.
 * You could vary the time that the workflow waits for other signals, say for a day, or a variable time from first
 *   signal with the GetNextTimeout() function.
 */

// signalToSignalTimeout is them maximum time between signals
const signalToSignalTimeout = 30 * time.Second

// fromFirstSignalTimeout is the maximum time to receive all signals
const fromFirstSignalTimeout = 60 * time.Second

type AccumulateGreeting struct {
	GreetingText string
	Bucket       string
	GreetingKey  string
}

type GreetingsInfo struct {
	BucketKey          string
	GreetingsList      []AccumulateGreeting
	UniqueGreetingKeys map[string]bool
}

// GetNextTimeout returns the maximum time allowed to wait for the next signal.
func (a *AccumulateGreeting) GetNextTimeout(ctx workflow.Context, timeToExit bool, firstSignalTime time.Time) (time.Duration, error) {
	if firstSignalTime.IsZero() {
		firstSignalTime = workflow.Now(ctx)
	}
	total := workflow.Now(ctx).Sub(firstSignalTime)
	totalLeft := fromFirstSignalTimeout - total
	if totalLeft <= 0 {
		return 0, nil
	}
	if signalToSignalTimeout < totalLeft {
		return signalToSignalTimeout, nil
	}
	return totalLeft, nil
}

// AccumulateSignalsWorkflow workflow definition
func AccumulateSignalsWorkflow(ctx workflow.Context, greetings GreetingsInfo) (allGreetings string, err error) {
	log := workflow.GetLogger(ctx)
	var a AccumulateGreeting
	if greetings.GreetingsList == nil {
		greetings.GreetingsList = []AccumulateGreeting{}
	}
	if greetings.UniqueGreetingKeys == nil {
		greetings.UniqueGreetingKeys = make(map[string]bool)
	}
	var unprocessedGreetings []AccumulateGreeting
	var firstSignalTime time.Time
	exitRequested := false

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Listen to signals in a different goroutine
	workflow.Go(ctx, func(gCtx workflow.Context) {
		for {
			log.Debug("In workflow.go signals goroutine for{}")
			selector := workflow.NewSelector(gCtx)
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "greeting"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, &a)
				unprocessedGreetings = append(unprocessedGreetings, a)
				log.Debug("Signal Received with text: " + a.GreetingText + ", more: " + strconv.FormatBool(more))

				a = AccumulateGreeting{}
			})
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "exit"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, nil)
				exitRequested = true
				log.Debug("Exit Signal Received")
			})
			log.Debug("Before select greeting  text: " + a.GreetingText)
			selector.Select(gCtx)
			log.Debug("After select, greeting  text: " + a.GreetingText)
			if firstSignalTime.IsZero() {
				firstSignalTime = workflow.Now(gCtx)
			}
		}
	})

	for !workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		// Wait for Signal or timeout
		timeout, err := a.GetNextTimeout(ctx, exitRequested, firstSignalTime)

		if err != nil {
			log.Warn("Error awaiting signal")
			return "", err
		}
		if timeout <= 0 {
			// do final processing? or just check for signal one more time
			log.Debug("No time left for signals, checking one more time")
		}

		log.Debug("Awaiting for " + timeout.String())
		gotSignalBeforeTimeout, _ := workflow.AwaitWithTimeout(ctx, timeout, func() bool {
			return len(unprocessedGreetings) > 0 || exitRequested
		})

		// timeout happened without a signal coming in, so let's process the greetings and wrap it up!
		if len(unprocessedGreetings) == 0 {
			log.Debug("Into final processing, signal received? " + strconv.FormatBool(gotSignalBeforeTimeout) + ", exit requested: " + strconv.FormatBool(exitRequested))
			// do final processing
			allGreetings = ""
			err := workflow.ExecuteActivity(ctx, ComposeGreeting, greetings.GreetingsList).Get(ctx, &allGreetings)
			if err != nil {
				log.Error("ComposeGreeting activity failed.", "Error", err)
				return allGreetings, err
			}

			return allGreetings, nil
		}

		/* process latest signal
		 * Here is where we can process individual signals as they come in.
		 * It's ok to call activities here.
		 * This also validates an individual greeting:
		 * - check for duplicates
		 * - check for correct bucket
		 */
		toProcess := unprocessedGreetings
		unprocessedGreetings = []AccumulateGreeting{}

		for _, ug := range toProcess {
			if ug.Bucket != greetings.BucketKey {
				log.Warn("Wrong bucket, something is wrong with your signal processing. WF Bucket: [" + greetings.BucketKey + "], greeting bucket: [" + ug.Bucket + "]")
			} else if greetings.UniqueGreetingKeys[ug.GreetingKey] {
				log.Warn("Duplicate Greeting Key. Key: [" + ug.GreetingKey + "]")
			} else {
				greetings.UniqueGreetingKeys[ug.GreetingKey] = true
				greetings.GreetingsList = append(greetings.GreetingsList, ug)
			}
		}

	}

	log.Debug("Accumulate workflow starting new run with " + strconv.Itoa(len(greetings.GreetingsList)) + " greetings.")
	return "Continued As New.", workflow.NewContinueAsNewError(ctx, AccumulateSignalsWorkflow, greetings)
}
