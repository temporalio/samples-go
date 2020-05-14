package dynamic

import (
	"fmt"
)

type Activities struct{}

// Get Name Activity.
func (a *Activities) GetName() (string, error) {
	return "Temporal", nil
}

// Get Greeting Activity.
func (a *Activities) GetGreeting() (string, error) {
	return "Hello", nil
}

// Say Greeting Activity.
func (a *Activities) SayGreeting(greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}
