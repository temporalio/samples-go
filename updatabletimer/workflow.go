package updatabletimer

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	QueryType  = "GetWakeUpTime"
	SignalType = "UpdateWakeUpTime"
)

// UpdatableTimer is an example of a timer that can have its wake time updated
type UpdatableTimer struct {
	wakeUpTime time.Time
}

// SleepUntil sleeps until the provided wake-up time.
// The wake-up time can be updated at any time by sending a new time over updateWakeUpTimeCh.
// Supports ctx cancellation.
// Returns temporal.CanceledError if ctx was canceled.
func (u *UpdatableTimer) SleepUntil(ctx workflow.Context, wakeUpTime time.Time, updateWakeUpTimeCh workflow.ReceiveChannel) (err error) {
	u.wakeUpTime = wakeUpTime
	timerFired := false
	for !timerFired && ctx.Err() == nil {
		ctx, timerCancel := workflow.WithCancel(ctx)
		duration := u.wakeUpTime.Sub(workflow.Now(ctx))
		workflow.NewSelector(ctx).
			AddFuture(workflow.NewTimer(ctx, duration), func(f workflow.Future) {
				err := f.Get(ctx, nil)
				// if a timer returned an error then it was canceled
				if err == nil {
					timerFired = true
				}
			}).
			AddReceive(updateWakeUpTimeCh, func(c workflow.ReceiveChannel, more bool) {
				timerCancel()                 // cancel outstanding timer
				c.Receive(ctx, &u.wakeUpTime) // update wake-up time
			}).
			Select(ctx)
	}
	return ctx.Err()
}

func (u *UpdatableTimer) GetWakeUpTime() time.Time {
	return u.wakeUpTime
}

// Workflow that sleeps initialWakeUpTime unless the new wake-up time is received through "UpdateWakeUpTime" signal.
func Workflow(ctx workflow.Context, initialWakeUpTime time.Time) error {
	workflow.GetLogger(ctx).Info("Start Workflow")
	timer := UpdatableTimer{}
	err := workflow.SetQueryHandler(ctx, QueryType, func() (time.Time, error) {
		return timer.GetWakeUpTime(), nil
	})
	if err != nil {
		return err
	}
	workflow.GetLogger(ctx).Info("Query Handler registered", QueryType)

	return timer.SleepUntil(ctx, initialWakeUpTime, workflow.GetSignalChannel(ctx, SignalType))
}
