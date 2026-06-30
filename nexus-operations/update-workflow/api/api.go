package api

const (
	// Nexus service exposing the Update operation
	CounterUpdateServiceName = "counter-update-service"
	// Name of the Nexus Operation- backed by the UpdateWorkflow- to bump counter
	IncrOperationName = "incr"
	// Name of the Update receiver on the handler
	IncrUpdateName = "incr"
	// Signal for the long running counter workflow(handler) to complete
	DoneSignalName = "done"
	// Task queue for the handler worker to process counter updates
	HandlerTaskQueueName = "counter-update-handler-tq"
	// Workflow ID of the handler workflow. Required to be known by the caller
	CounterWorkflowID = "counter-workflow-1"
	// Nexus endpoint
	EndpointName = "counter-update-endpoint"
)

type Input struct {
	WorkflowID string
	Incr       int
}

type Output struct {
	NewCount int
}
