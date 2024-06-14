package accumulator

import (
	"fmt"
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

// SignalToSignalTimeout is them maximum time between signals
var SignalToSignalTimeout = 30 * time.Second

// FromFirstSignalTimeout is the maximum time to receive all signals
var FromFirstSignalTimeout = 60 * time.Second

type AccumulateGreeting struct {
	GreetingText    string
	Bucket          string
	GreetingKey     string
}


// GetNextTimeout returns the maximum time allowed to wait for the next signal.
func (a *AccumulateGreeting) GetNextTimeout(ctx workflow.Context, timeToExit bool, firstSignalTime time.Time ) (time.Duration, error) {
	if firstSignalTime.IsZero() {
		firstSignalTime = workflow.Now(ctx)
	}
	total := workflow.Now(ctx).Sub(firstSignalTime)
	totalLeft := FromFirstSignalTimeout - total
	if totalLeft <= 0 {
		return 0, nil
	}
	if SignalToSignalTimeout < totalLeft {
		return SignalToSignalTimeout, nil
	}
	return totalLeft, nil
}

// AccumulateSignalsWorkflow workflow definition
func AccumulateSignalsWorkflow(ctx workflow.Context, bucketKey string, greetingsCAN []AccumulateGreeting, uniqueGreetingKeysMapCAN map[string]bool) (allGreetings string, err error) {
	log := workflow.GetLogger(ctx)
	var a AccumulateGreeting
	greetings := []AccumulateGreeting{}
	greetings = append(greetings, greetingsCAN...)
	unprocessedGreetings := []AccumulateGreeting{}
	uniqueGreetingKeysMap := make(map[string]bool)
	for k, v := range uniqueGreetingKeysMapCAN {
		uniqueGreetingKeysMap[k] = v
	}
	var FirstSignalTime time.Time
	ExitRequested := false

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Listen to signals in a different goroutine
	workflow.Go(ctx, func(gCtx workflow.Context) {
		for {
			log.Debug("in workflow.go signals goroutine for{} \n")
			selector := workflow.NewSelector(gCtx)
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "greeting"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, &a)
				unprocessedGreetings = append(unprocessedGreetings, a)
				log.Debug("Signal Received with text: " + a.GreetingText + ", more: " + strconv.FormatBool(more) + "\n")
				
				a = AccumulateGreeting{} 
			})
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "exit"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, nil)
				ExitRequested = true
				log.Debug("Exit Signal Received\n")
			})
			log.Debug("before select, greeting  text: " + a.GreetingText + "\n")
			selector.Select(gCtx)
			log.Debug("after select, greeting  text: " + a.GreetingText + "\n")
			if FirstSignalTime.IsZero() {
				FirstSignalTime = workflow.Now(gCtx)
			}
		}
	})

	for ; !workflow.GetInfo(ctx).GetContinueAsNewSuggested() ;{
		// Wait for Signal or timeout
		timeout, err := a.GetNextTimeout(ctx, ExitRequested, FirstSignalTime)

		if err != nil {
			log.Warn("error awaiting signal\n")
			return "", err
		}
		if timeout <= 0 {
			// do final processing? or just check for signal one more time
			log.Debug("No time left for signals, checking one more time\n")
		}

		log.Debug("Awaiting for " + timeout.String() + "\n")
		gotSignalBeforeTimeout, err := workflow.AwaitWithTimeout(ctx, timeout, func() bool {
			return len(unprocessedGreetings) > 0 || ExitRequested
		})

		// timeout happened without a signal coming in, so let's process the greetings and wrap it up!
		if len(unprocessedGreetings) == 0 {
			log.Debug("Into final processing, signal received? " + strconv.FormatBool(gotSignalBeforeTimeout)  + ", exit requested: " + strconv.FormatBool(ExitRequested) +"\n")
			// do final processing
			allGreetings = ""
			err := workflow.ExecuteActivity(ctx, ComposeGreeting, greetings).Get(ctx, &allGreetings)
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
		for len(unprocessedGreetings) > 0 {
			ug := unprocessedGreetings[0]
			unprocessedGreetings = unprocessedGreetings[1:] 
			
			if ug.Bucket != bucketKey {
				log.Warn("Wrong bucket, something is wrong with your signal processing. WF Bucket: [" + bucketKey +"], greeting bucket: [" + ug.Bucket + "]");
			} else if(uniqueGreetingKeysMap[ug.GreetingKey]) {
				log.Warn("Duplicate Greeting Key. Key: [" + ug.GreetingKey +"]");
			} else {
				uniqueGreetingKeysMap[ug.GreetingKey] = true
				greetings = append(greetings, ug)
			}
		}
		
	}

	log.Debug("Accumulate workflow starting new run with " + strconv.Itoa(len(greetings)) + " greetings.")
	return "Continued As New.", workflow.NewContinueAsNewError(ctx, AccumulateSignalsWorkflow, bucketKey, greetings, uniqueGreetingKeysMap)
}

func printGreetings(s []AccumulateGreeting) {
	fmt.Printf("greetings slice info: len=%d cap=%d %v\n", len(s), cap(s), s)
}