package metrics

import (
	"time"

	"github.com/uber-go/tally"
)

const (
	activityLatency        = "activity_latency"
	scheduleToStartLatency = "schedule_to_start_latency"

	activityStartedCount = "activity_started"
	activityFailedCount  = "activity_failed"
	activitySuccessCount = "activity_succeeded"
)

func recordActivityStart(scope tally.Scope, activityType string, scheduledTimeNanos int64) (tally.Scope, tally.Stopwatch) {
	scope = scope.Tagged(map[string]string{"operation": activityType})
	scheduleToStartDuration := time.Duration(time.Now().UnixNano() - scheduledTimeNanos)
	scope.Timer(scheduleToStartLatency).Record(scheduleToStartDuration)
	scope.Counter(activityStartedCount).Inc(1)
	sw := scope.Timer(activityLatency).Start()
	return scope, sw
}

// recordActivityEnd emits metrics at the end of an activity function
func recordActivityEnd(scope tally.Scope, sw tally.Stopwatch, err error) {
	sw.Stop()
	if err != nil {
		scope.Counter(activityFailedCount).Inc(1)
		return
	}
	scope.Counter(activitySuccessCount).Inc(1)
}
