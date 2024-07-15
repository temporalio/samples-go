package accumulator

import (
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
)

/**
 * This sample demonstrates how to accumulate many signals (business events) over a time period.
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

// fromStartTimeout is the maximum time to receive all signals
const fromStartTimeout = 60 * time.Second

// exitTimeout is the time to wait after exit is requested to catch any last few signals
const exitTimeout = 1 * time.Second

type AccumulateGreeting struct {
	GreetingText string
	Bucket       string
	GreetingKey  string
}

type GreetingsInfo struct {
	BucketKey          string
	GreetingsList      []AccumulateGreeting
	UniqueGreetingKeys map[string]bool
	startTime          time.Time
}

// GetNextTimeout returns the maximum time for a workflow to wait for the next signal.
// This waits for the greater of the remaining fromStartTimeout and  signalToSignalTimeout
// fromStartTimeout and signalToSignalTimeout can be adjusted to wait for the right amount of time as desired
// This resets with Continue As New
func (a *AccumulateGreeting) GetNextTimeout(ctx workflow.Context, startTime time.Time, exitRequested bool) (time.Duration, error) {
	if exitRequested {
		return exitTimeout, nil
	}
	if startTime.IsZero() {
		startTime = workflow.GetInfo(ctx).WorkflowStartTime // if you want to start from the time of the first signal, customize this
	}
	total := workflow.Now(ctx).Sub(startTime)
	totalLeft := fromStartTimeout - total
	if totalLeft <= 0 {
		return 0, nil
	}
	if signalToSignalTimeout > totalLeft {
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
	if greetings.startTime.IsZero() {
		greetings.startTime = workflow.Now(ctx)
	}
	exitRequested := false

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 100 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for !workflow.GetInfo(ctx).GetContinueAsNewSuggested() {

		timeout, err := a.GetNextTimeout(ctx, greetings.startTime, exitRequested)
		childCtx, cancelHandler := workflow.WithCancel(ctx)
		selector := workflow.NewSelector(ctx)

		if err != nil {
			log.Error("Error calculating timeout")
			return "", err
		}
		log.Debug("Awaiting for " + timeout.String())
		selector.AddReceive(workflow.GetSignalChannel(ctx, "greeting"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &a)
			unprocessedGreetings = append(unprocessedGreetings, a)
			log.Debug("Signal Received with text: " + a.GreetingText + ", more?: " + strconv.FormatBool(more) + "\n")
			cancelHandler() // cancel timer future
			a = AccumulateGreeting{}
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "exit"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			exitRequested = true
			cancelHandler() // cancel timer future
			log.Debug("Exit Signal Received, more?: " + strconv.FormatBool(more) + "\n")
		})

		timerFuture := workflow.NewTimer(childCtx, timeout)
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			log.Debug("Timer fired \n")
		})

		selector.Select(ctx)

		if len(unprocessedGreetings) == 0 { // timeout without a signal coming in, so let's process the greetings and wrap it up!
			log.Debug("Into final processing, received greetings count: " + strconv.Itoa(len(greetings.GreetingsList)) + "\n")
			allGreetings = ""
			err := workflow.ExecuteActivity(ctx, ComposeGreeting, greetings.GreetingsList).Get(ctx, &allGreetings)
			if err != nil {
				log.Error("ComposeGreeting activity failed.", "Error", err)
				return allGreetings, err
			}

			if !selector.HasPending() { // in case a signal came in while activity was running, check again
				return allGreetings, nil
			} else {
				log.Info("Received a signal while processing ComposeGreeting activity.")
			}
		}

		/* process latest signals
		 * Here is where we can process individual signals as they come in.
		 * It's ok to call activities here.
		 * This also validates an individual greeting:
		 * - check for duplicates
		 * - check for correct bucket
		 * Using update validation could improve this in the future
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
