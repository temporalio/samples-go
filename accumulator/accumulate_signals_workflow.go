package accumulator

import (
	"fmt"
	"strconv"
	"time"
	

	//"go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/workflow"
)

/**
 * The sample demonstrates how to deal with multiple signals that can come out of order and require actions
 * if a certain signal not received in a specified time interval.
 *
 * This specific sample receives three signals: Signal1, Signal2, Signal3. They have to be processed in the
 * sequential order, but they can be received out of order.
 * There are two timeouts to enforce.
 * The first one is the maximum time between signals.
 * The second limits the total time since the first signal received.
 *
 * A naive implementation of such use case would use a single loop that contains a Selector to listen on three
 * signals and a timer. Something like:

 *	for {
 *		selector := workflow.NewSelector(ctx)
 *		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal1"), func(c workflow.ReceiveChannel, more bool) {
 *			// Process signal1
 *		})
 *		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal2"), func(c workflow.ReceiveChannel, more bool) {
 *			// Process signal2
 *		}
 *		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal3"), func(c workflow.ReceiveChannel, more bool) {
 *			// Process signal3
 *		}
 *		cCtx, cancel := workflow.WithCancel(ctx)
 *		timer := workflow.NewTimer(cCtx, timeToNextSignal)
 *		selector.AddFuture(timer, func(f workflow.Future) {
 *			// Process timeout
 *		})
 * 		selector.Select(ctx)
 *	    cancel()
 *      // break out of the loop on certain condition
 *	}
 *
 *  The above implementation works. But it quickly becomes pretty convoluted if the number of signals
 *  and rules around order of their arrivals and timeouts increases.
 *
 *  The following example demonstrates an alternative approach. It receives signals in a separate goroutine.
 *  Each signal handler just updates a correspondent shared variable with the signal data.
 *  The main workflow function awaits the next step using `workflow.AwaitWithTimeout` using condition composed of
 *  the shared variables. This makes the main workflow method free from signal callbacks and makes the business logic
 *  clear.
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

/* todo section
[x] listen for signals
[x] add to slice
[x] take fisrtsignaltime and exitrequested out of the struct
[x] test exit signal
[x] signal with start
[x] starter like java
[ ] tests like java
[x] consider checking for multiple messages in the signal wait loop
[x] "process" each greeting as they come in -- activity? no
[x] activity to combine all greetings
[ ] make GetNextTimeout not be a func on the struct
[ ] fix the extra listen
[ ] continue as new check and doing it
[ ] decide to use a separate goroutine function or keep the one you have
[ ] for fun race vs java
[ ] update readme
*/

// Listen to signals - greetings and exit
func Listen(ctx workflow.Context, a AccumulateGreeting, ExitRequested bool, FirstSignalTime time.Time) {
	log := workflow.GetLogger(ctx)
	for {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(workflow.GetSignalChannel(ctx, "greeting"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &a)
			log.Info("Signal Received")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "exit"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			ExitRequested = true
			log.Info("Exit Signal Received")
		})
		selector.Select(ctx)
		if FirstSignalTime.IsZero() {
			FirstSignalTime = workflow.Now(ctx)
		}
	}
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
// func AccumulateSignalsWorkflow(ctx workflow.Context, bucketKey string, greetings []AccumulateGreeting, greetingKeys map[string]bool) (greeting string, err error) {
func AccumulateSignalsWorkflow(ctx workflow.Context, bucketKey string) (allGreetings string, err error) {
	log := workflow.GetLogger(ctx)
	var a AccumulateGreeting
	greetings := []AccumulateGreeting{}
	unprocessedGreetings := []AccumulateGreeting{}
	uniqueGreetingKeysMap := make(map[string]bool)
	var FirstSignalTime time.Time
	ExitRequested := false

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	printGreetings(greetings)
	// Listen to signals in a different goroutine
	workflow.Go(ctx, func(gCtx workflow.Context) {
		for {
			log.Info("in workflow.go signals goroutine for{} \n")
			selector := workflow.NewSelector(gCtx)
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "greeting"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, &a)
				unprocessedGreetings = append(unprocessedGreetings, a)
				log.Info("Signal Received with text: " + a.GreetingText + ", more: " + strconv.FormatBool(more) + "\n")
				// initialize a
				a = AccumulateGreeting{} 
			})
			selector.AddReceive(workflow.GetSignalChannel(gCtx, "exit"), func(c workflow.ReceiveChannel, more bool) {
				c.Receive(gCtx, nil)
				ExitRequested = true
				log.Info("Exit Signal Received\n")
			})
			//log.Info("before select, greeting  text: " + a.GreetingText + "\n")
			selector.Select(gCtx)
			//log.Info("after select, greeting  text: " + a.GreetingText + "\n")
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
			log.Info("no time left for signals, checking one more time\n")
		}

		printGreetings(greetings)
		log.Info("Awaiting for " + timeout.String() + "\n")
		gotSignalBeforeTimeout, err := workflow.AwaitWithTimeout(ctx, timeout, func() bool {
			return len(unprocessedGreetings) > 0 || ExitRequested
		})
		printGreetings(greetings)

		// timeout
		if len(unprocessedGreetings) == 0 {
			log.Info("Into final processing, signal received? " + strconv.FormatBool(gotSignalBeforeTimeout)  + ", exit requested: " + strconv.FormatBool(ExitRequested) +"\n")
			// do final processing
			//printGreetings(greetings)
			allGreetings = ""
			// get token - retryable like normal, it's failure-prone and idempotent
			err := workflow.ExecuteActivity(ctx, ComposeGreeting, greetings).Get(ctx, &allGreetings)
			if err != nil {
				log.Error("ComposeGreeting activity failed.", "Error", err)
				return allGreetings, err
			}
			return allGreetings, nil
		}
		
		// process latest signal
		// Here is where we can process individual signals as they come in.
		// It's ok to call activities here.
		// This also validates an individual greeting:
		// - check for duplicates
		// - check for correct bucket 

		for len(unprocessedGreetings) > 0 {
			ug := unprocessedGreetings[0]
			unprocessedGreetings = unprocessedGreetings[1:] 
			//fmt.Printf("greetings slice info for unprocessedGreetings after taking out ug: len=%d cap=%d %v\n", len(unprocessedGreetings), cap(unprocessedGreetings), unprocessedGreetings)
			if ug.Bucket != bucketKey {
				log.Warn("Wrong bucket, something is wrong with your signal processing. WF Bucket: [" + bucketKey +"], greeting bucket: [" + ug.Bucket + "]");
			} else if(uniqueGreetingKeysMap[ug.GreetingKey]) {
				log.Warn("Duplicate Greeting Key. Key: [" + ug.GreetingKey +"]");
			} else {
				uniqueGreetingKeysMap[ug.GreetingKey] = true
				greetings = append(greetings, ug)
				//log.Info("Adding Greeting. Key: [" + ug.GreetingKey +"], Text: [" + ug.GreetingText + "]");
			}
		}
		
		//a = AccumulateGreeting{} 
	}

	return a.GreetingText, nil
}

func printGreetings(s []AccumulateGreeting) {
	fmt.Printf("greetings slice info: len=%d cap=%d %v\n", len(s), cap(s), s)
}