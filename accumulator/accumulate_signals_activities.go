package accumulator

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/activity"
)

// this activity will process all of the signals together
func ComposeGreeting(ctx context.Context, s []AccumulateGreeting) (string, error) {
	log := activity.GetLogger(ctx)
	if len(s) == 0 {
		log.Warn("No greetings found when trying to Compose Greetings.")
	}

	words := fmt.Sprintf("Hello (%v) Robots", len(s))
	for _, v := range s {
		words += ", " + v.GreetingText
	}
	words += "!"
	return words, nil

}
