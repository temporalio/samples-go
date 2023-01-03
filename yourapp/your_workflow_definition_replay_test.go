package yourapp

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/workflowservicemock/v1"
	"go.temporal.io/sdk/worker"
)

type replayTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service  *workflowservicemock.MockWorkflowServiceClient
}

func TestReplayTestSuite(t *testing.T) {
	s := new(replayTestSuite)
	suite.Run(t, s)
}

func (s *replayTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicemock.NewMockWorkflowServiceClient(s.mockCtrl)
}

func (s *replayTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

// This replay test is the recommended way to make sure changing workflow code is backward compatible without non-deterministic errors.
// "yourworkflow.json" can be downloaded from Temporal CLI:
//
//	tctl wf show -w your-workflow-id --output_filename ./your-workflow-id.json
//
// Or from Temporal Web UI. And you may need to change workflowType in the first event.
func (s *replayTestSuite) TestReplayWorkflowHistoryFromFile() {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(YourWorkflowDefinition)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "yourworkflow.json")
	require.NoError(s.T(), err)
}
