package accumulator

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
)

// this activity will process all of the signals together
func ComposeGreeting(ctx context.Context, s []AccumulateGreeting) (string, error) {
	log := activity.GetLogger(ctx)
	log.Info("Compose Greetings Activity started. ")
	fmt.Printf("greetings slice info: len=%d cap=%d %v\n", len(s), cap(s), s)
	if(len(s) < 1) {
		log.Warn("No greetings found when trying to Compose Greetings. ")
		return "", nil
	}

	words := "Hello"
	for _, v:= range s {
		words += fmt.Sprintf(", " + v.GreetingText )
	}

	words += "!"
	return words, nil
	
	/*       List<String> greetingList =
	      greetings.stream().map(u -> u.greetingText).collect(Collectors.toList());
	  return "Hello (" + greetingList.size() + ") robots: " + greetingList + "!";
	}
	*/
}
