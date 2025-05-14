// @@@SNIPSTART samples-go-nexus-service
package service

const HelloServiceName = "my-hello-service"

// Echo operation
const EchoOperationName = "echo"

type EchoInput struct {
	Message string
}

type EchoOutput EchoInput

// Hello operation
const HelloOperationName = "say-hello"
const CancelOperationName = "say-cancel"

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

// @@@SNIPEND

