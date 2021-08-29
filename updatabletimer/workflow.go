package updatabletimer

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue  = "updatable-timer"
	WorkflowID = "updatable-timer"
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
	logger := workflow.GetLogger(ctx)
	u.wakeUpTime = wakeUpTime
	timerFired := false
	for !timerFired && ctx.Err() == nil {
		timerCtx, timerCancel := workflow.WithCancel(ctx)
		duration := u.wakeUpTime.Sub(workflow.Now(timerCtx))
		timer := workflow.NewTimer(timerCtx, duration)
		logger.Info("SleepUntil", "WakeUpTime", u.wakeUpTime)
		workflow.NewSelector(timerCtx).
			AddFuture(timer, func(f workflow.Future) {
				err := f.Get(timerCtx, nil)
				// if a timer returned an error then it was canceled
				if err == nil {
					logger.Info("Timer fired")
					timerFired = true
				} else if ctx.Err() != nil { // Only log on root ctx cancellation, not on timerCancel function call.
					logger.Info("SleepUntil canceled")
				}
			}).
			AddReceive(updateWakeUpTimeCh, func(c workflow.ReceiveChannel, more bool) {
				timerCancel()                      // cancel outstanding timer
				c.Receive(timerCtx, &u.wakeUpTime) // update wake-up time
				logger.Info("Wake up time update requested")
			}).
			Select(timerCtx)
	}
	return ctx.Err()
}

func (u *UpdatableTimer) GetWakeUpTime() time.Time {
	return u.wakeUpTime
}

// Workflow that sleeps initialWakeUpTime unless the new wake-up time is received through "UpdateWakeUpTime" signal.
func Workflow(ctx workflow.Context, initialWakeUpTime time.Time) error {
	timer := UpdatableTimer{}
	err := workflow.SetQueryHandler(ctx, QueryType, func() (time.Time, error) {
		return timer.GetWakeUpTime(), nil
	})
	if err != nil {
		return err
	}
	return timer.SleepUntil(ctx, initialWakeUpTime, workflow.GetSignalChannel(ctx, SignalType))
}
