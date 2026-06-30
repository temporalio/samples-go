package api

// Nexus service exposing the Update operation
const CounterUpdateServiceName = "counter-update-service"

// Name of the Nexus Operation- backed by the UpdateWorkflow- to bump counter
const IncrOperationName = "incr"

// Name of the Update receiver on the handler
const IncrUpdateName = "incr"

// Signal for the long running counter workflow(handler) to complete
const DoneSignalName = "done"

// Task queue for the handler worker to process counter updates
const HandlerTaskQueueName = "counter-update-handler-tq"

// Workflow ID of the handler workflow. Required to be known by the caller
const CounterWorkflowID = "counter-workflow-1"

type Input struct {
	WorkflowID string
}

type Output struct {
	NewCount int
}
