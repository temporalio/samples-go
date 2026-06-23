package service

const HelloServiceName = "my-hello-service"

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

