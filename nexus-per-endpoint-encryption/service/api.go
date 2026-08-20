package service

const (
	Name = "per-endpoint-encryption-service"

	EndpointA = "nexus-encryption-endpoint-a"
	EndpointB = "nexus-encryption-endpoint-b"

	SyncOperationName  = "encrypted-echo"
	AsyncOperationName = "encrypted-workflow"
)

type OperationInput struct {
	Message   string
	RequestID string
}

type OperationOutput struct {
	Message string
}

type WorkflowOutput struct {
	EndpointASync  string
	EndpointAAsync string
	EndpointBSync  string
	EndpointBAsync string
}
