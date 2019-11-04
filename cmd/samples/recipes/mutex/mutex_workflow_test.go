package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
	"go.temporal.io/temporal/testsuite"
	"go.temporal.io/temporal/worker"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	var h common.SampleHelper
	s.env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), _sampleHelperContextKey, h),
	})
}

func (s *UnitTestSuite) Test_Workflow_Success() {
	env := s.NewTestWorkflowEnvironment()
	mockResourceID := "mockResourceID"
	MockMutexLock(env, mockResourceID, nil)
	env.ExecuteWorkflow(SampleWorkflowWithMutex, mockResourceID)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_Workflow_Error() {
	env := s.NewTestWorkflowEnvironment()
	mockResourceID := "mockResourceID"
	MockMutexLock(env, mockResourceID, errors.New("bad-error"))
	env.ExecuteWorkflow(SampleWorkflowWithMutex, mockResourceID)

	s.True(env.IsWorkflowCompleted())
	s.EqualError(env.GetWorkflowError(), "bad-error")
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_MutexWorkflow_Success() {
	mockNamespace := "mockNamespace"
	mockResourceID := "mockResourceID"
	mockUnlockTimeout := 10 * time.Minute
	mockSenderWorkflowID := "mockSenderWorkflowID"
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RequestLockSignalName, mockSenderWorkflowID)
	}, time.Millisecond*0)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unlock-event-mockSenderWorkflowID", "releaseLock")
	}, time.Millisecond*0)
	s.env.OnSignalExternalWorkflow(mock.Anything, mockSenderWorkflowID, "",
		AcquireLockSignalName, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(
		MutexWorkflow,
		mockNamespace,
		mockResourceID,
		mockUnlockTimeout,
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_MutexWorkflow_TimeoutSuccess() {
	mockNamespace := "mockNamespace"
	mockResourceID := "mockResourceID"
	mockUnlockTimeout := 10 * time.Minute
	mockSenderWorkflowID := "mockSenderWorkflowID"
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RequestLockSignalName, mockSenderWorkflowID)
	}, time.Millisecond*0)
	s.env.OnSignalExternalWorkflow(mock.Anything, mockSenderWorkflowID, "",
		AcquireLockSignalName, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(
		MutexWorkflow,
		mockNamespace,
		mockResourceID,
		mockUnlockTimeout,
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
