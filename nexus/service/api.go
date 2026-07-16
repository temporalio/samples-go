// @@@SNIPSTART samples-go-nexus-service
package service

import (
	"fmt"
	"strings"
)

const HelloServiceName = "my-hello-service"

// Echo operation
const EchoOperationName = "echo"

type EchoInput struct {
	Message string
}

type EchoOutput EchoInput

// Hello operation
const HelloOperationName = "say-hello"

type Language string

const (
	EN Language = "en"
	FR Language = "fr"
	DE Language = "de"
	ES Language = "es"
	TR Language = "tr"
)

type HelloInput struct {
	Name     string
	Language Language
}

type HelloOutput struct {
	Message string
}

func HelloWorkflowID(input HelloInput) string {
	name := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(input.Name), " ", "-"))
	return fmt.Sprintf("hello-%s-%s", input.Language, name)
}

// @@@SNIPEND
