// @@@SNIPSTART yourapp-workflow-replay-test
package yourapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/worker"
)

// TestReplayWorkflowHistoryFromFile tests the code against the existing Worklow History saved to the JSON file.
// This Replay test is the recommended way to make sure changing workflow code is backward compatible without non-deterministic errors.
// "your_workflow_history.json" can be downloaded from the Web UI or the Temporal CLI:
//
//	tctl wf show -w your-workflow-id --output_filename ./your_workflow_history.json
func TestReplayWorkflowHistoryFromFile(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(YourWorkflowDefinition)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "your_workflow_history.json")
	require.NoError(t, err)
}

// @@@SNIPEND
