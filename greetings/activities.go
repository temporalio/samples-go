package greetings

import "fmt"

// @@@SNIPSTART samples-go-dependency-sharing-activities
type Activities struct {
	Name     string
	Greeting string
}

// SayGreeting Activity.
func (a *Activities) SayGreeting(greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}

// GetName Activity.
func (a *Activities) GetName() (string, error) {
	return a.Name, nil
}

// GetGreeting Activity.
func (a *Activities) GetGreeting() (string, error) {
	return a.Greeting, nil
}
// @@@SNIPEND
