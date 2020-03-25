package helloworld

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/temporal-proto/workflowservicemock"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"
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
// "helloworld.json" can be downloaded from Temporal CLI:
//      tctl wf show -w helloworld_d002cd3a-aeee-4a11-aa30-1c62385b4d87 --output_filename ~/tmp/helloworld.json
// Or from Temporal Web UI. And you may need to change workflowType in the first event.
func (s *replayTestSuite) TestReplayWorkflowHistoryFromFile() {
	logger, _ := zap.NewDevelopment()
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(HelloworldWorkflow)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(logger, "helloworld.json")
	require.NoError(s.T(), err)
}
