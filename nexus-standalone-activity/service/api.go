// @@@SNIPSTART samples-go-nexus-standalone-activity-service
package service

import "fmt"

// Nexus service definition shared by the caller and the handler. It declares a single operation
// whose backing execution is a Standalone Activity.

const GreetingServiceName = "NexusGreetingService"

const GreetOperationName = "greet"

type GreetingInput struct {
	Name string
}

type GreetingOutput struct {
	Message string
}

func GreetingActivityID(input GreetingInput) string {
	return fmt.Sprintf("greeting-%s", input.Name)
}

// @@@SNIPEND
