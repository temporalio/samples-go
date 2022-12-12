package await_signals

import (
	"go.temporal.io/sdk/temporal"
	"time"

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

type AwaitSignals struct {
	FirstSignalTime time.Time
	Signal1Received bool
	Signal2Received bool
	Signal3Received bool
}

// Listen to signals Signal1, Signal2, and Signal3
func (a *AwaitSignals) Listen(ctx workflow.Context) {
	log := workflow.GetLogger(ctx)
	for {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal1"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			a.Signal1Received = true
			log.Info("Signal1 Received")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal2"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			a.Signal2Received = true
			log.Info("Signal2 Received")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal3"), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			a.Signal3Received = true
			log.Info("Signal3 Received")
		})
		selector.Select(ctx)
		if a.FirstSignalTime.IsZero() {
			a.FirstSignalTime = workflow.Now(ctx)
		}
	}
}

// GetNextTimeout returns the maximum time allowed to wait for the next signal.
func (a *AwaitSignals) GetNextTimeout(ctx workflow.Context) (time.Duration, error) {
	if a.FirstSignalTime.IsZero() {
		panic("FirstSignalTime is not yet set")
	}
	total := workflow.Now(ctx).Sub(a.FirstSignalTime)
	totalLeft := FromFirstSignalTimeout - total
	if totalLeft <= 0 {
		return 0, temporal.NewApplicationError("FromFirstSignalTimeout", "timeout")
	}
	if SignalToSignalTimeout < totalLeft {
		return SignalToSignalTimeout, nil
	}
	return totalLeft, nil
}

// AwaitSignalsWorkflow workflow definition
func AwaitSignalsWorkflow(ctx workflow.Context) (err error) {
	log := workflow.GetLogger(ctx)
	var a AwaitSignals
	// Listen to signals in a different goroutine
	workflow.Go(ctx, a.Listen)

	// Wait for Signal1
	err = workflow.Await(ctx, func() bool {
		return a.Signal1Received
	})
	// Cancellation
	if err != nil {
		return
	}
	log.Info("Signal1 Processed")

	// Wait for Signal2
	timeout, err := a.GetNextTimeout(ctx)
	// No time left. At this point this cannot really happen.
	if err != nil {
		return
	}
	ok, err := workflow.AwaitWithTimeout(ctx, timeout, func() bool {
		return a.Signal2Received
	})
	// Cancellation
	if err != nil {
		return
	}
	// timeout
	if !ok {
		return temporal.NewApplicationError("Timed out waiting for signal2", "timeout")
	}
	log.Info("Signal2 Processed")

	// Wait for Signal3
	timeout, err = a.GetNextTimeout(ctx)
	// No time left.
	if err != nil {
		return
	}
	ok, err = workflow.AwaitWithTimeout(ctx, timeout, func() bool {
		return a.Signal3Received
	})
	// Cancellation
	if err != nil {
		return
	}
	// timeout
	if !ok {
		return temporal.NewApplicationError("Timed out waiting for signal3", "timeout")
	}
	log.Info("Signal3 Processed")
	return nil
}
