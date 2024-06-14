package accumulator

import (
	"context"
	"fmt"
	"strconv"
	"go.temporal.io/sdk/activity"
)

// this activity will process all of the signals together
func ComposeGreeting(ctx context.Context, s []AccumulateGreeting) (string, error) {
	log := activity.GetLogger(ctx)
	if(len(s) < 1) {
		log.Warn("No greetings found when trying to Compose Greetings. ")
	}

	words := "Hello (" + strconv.Itoa(len(s)) + ") Robots"
	for _, v:= range s {
		words += fmt.Sprintf(", " + v.GreetingText )
	}

	words += "!"
	return words, nil
	
}
