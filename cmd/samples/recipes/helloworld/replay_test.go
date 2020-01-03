package main

import (
	"testing"

	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-proto-go/workflowservicemock"
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
// "helloworld.json" can be downloaded from cadence CLI:
//      cadence --do samples-domain wf show -w helloworld_d002cd3a-aeee-4a11-aa30-1c62385b4d87 --output_filename ~/tmp/helloworld.json
// Or from Cadence Web UI. And you may need to change workflowType in the first event.
func (s *replayTestSuite) TestReplayWorkflowHistoryFromFile() {
	logger, _ := zap.NewDevelopment()
	err := worker.ReplayWorkflowHistoryFromJSONFile(logger, "helloworld.json")
	require.NoError(s.T(), err)
}
