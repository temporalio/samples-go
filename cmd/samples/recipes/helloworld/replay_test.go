package main

import (
	"testing"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
)

type internalWorkerTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service  *workflowservicetest.MockClient
}

func TestInternalWorkferTestSuite(t *testing.T) {
	s := new(internalWorkerTestSuite)
	suite.Run(t, s)
}

func (s *internalWorkerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicetest.NewMockClient(s.mockCtrl)
}

func (s *internalWorkerTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *internalWorkerTestSuite) TestReplayWorkflowHistoryFromFile() {
	logger, _ := zap.NewDevelopment()
	err := worker.ReplayWorkflowHistoryFromJSONFile(logger, "helloworld.json")
	require.NoError(s.T(), err)
}
