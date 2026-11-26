package requestcancelexternalworkflow

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"testing"
	"time"
)

const childWorkflowID string = "luke"
const signalName string = "mysignal"
const signalValueDone string = "done"

func MyWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	info := workflow.GetInfo(ctx)

	logger := log.With(workflow.GetLogger(ctx), "type", info.WorkflowType, "wid", info.WorkflowExecution.ID)
	logger.Info("workflow started", "info", workflow.GetInfo(ctx))

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:          childWorkflowID,
		WaitForCancellation: true,
	})

	var signalValue string
	var childFuture workflow.Future
	childFuture = workflow.ExecuteChildWorkflow(childCtx, MyChild)
	sigChan := workflow.GetSignalChannel(ctx, signalName)
	sel := workflow.NewSelector(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		sel.AddReceive(sigChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &signalValue)
			logger.Info("signal received", "signal", signalName, signalValue)
		})
		sel.Select(ctx)
	})

	if err := workflow.Await(ctx, func() bool {
		return signalValue == signalValueDone
	}); err != nil {
		logger.Error("await err", "err", err)
	}
	logger.Info("released from block")
	if err := workflow.RequestCancelExternalWorkflow(childCtx, childWorkflowID, "").Get(ctx, nil); err != nil {
		logger.Error("request cancel err", "err", err)
	}
	if err := childFuture.Get(ctx, nil); err != nil {
		logger.Error("child GET err", "err", err)
	}
	return nil
}
func MyChild(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	info := workflow.GetInfo(ctx)
	logger := log.With(workflow.GetLogger(ctx), "type", info.WorkflowType, "wid", info.WorkflowExecution.ID)
	logger.Info("workflow started", "info", workflow.GetInfo(ctx))

	var signalValue string
	sigChan := workflow.GetSignalChannel(ctx, signalName)
	sel := workflow.NewSelector(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		sel.AddReceive(sigChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &signalValue)
			logger.Info("signal received", "signal", signalName, signalValue)
		})
		sel.Select(ctx)
	})

	if err := workflow.Await(ctx, func() bool {
		return signalValue == signalValueDone
	}); err != nil {
		logger.Error("await err", "err", err)
	}
	logger.Info("released from block")
	return nil
}

// https://docs.temporal.io/docs/go/testing/
type CancelTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *CancelTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *CancelTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *CancelTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *CancelTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *CancelTestSuite) Test_Cancel_NoMocky() {

	s.env.RegisterWorkflow(MyWorkflow)
	s.env.RegisterWorkflow(MyChild)
	//var cancelRequestCalled bool
	var childCancelInfo *workflow.Info
	var childCompletedInfo *workflow.Info
	s.env.SetOnChildWorkflowCanceledListener(func(workflowInfo *workflow.Info) {
		childCancelInfo = workflowInfo
	})
	s.env.SetOnChildWorkflowCompletedListener(func(workflowInfo *workflow.Info, result converter.EncodedValue, err error) {
		childCompletedInfo = workflowInfo
	})
	//s.env.OnRequestCancelExternalWorkflow(mock.Anything, childWorkflowID, "").Run(func(args mock.Arguments) {
	//	cancelRequestCalled = true
	//}).Return(nil).Once()

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflowByID("default-test-workflow-id", signalName, signalValueDone)
	}, time.Second*1)
	s.env.ExecuteWorkflow(MyWorkflow)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	//s.True(cancelRequestCalled)
	s.NotNil(childCancelInfo)
	s.NotNil(childCompletedInfo)

}
func (s *CancelTestSuite) Test_Cancel_WithMocky() {

	s.env.RegisterWorkflow(MyWorkflow)
	s.env.RegisterWorkflow(MyChild)
	//var cancelRequestCalled bool
	var childCancelInfo *workflow.Info
	var childCompletedInfo *workflow.Info
	// this mock is broken...MyChild does not require any arguments but the workflowInterceptor in our test suite panics without args
	s.env.OnWorkflow(MyChild, mock.Anything).Return(func(ctx workflow.Context) error {
		return nil
	})
	s.env.SetOnChildWorkflowCanceledListener(func(workflowInfo *workflow.Info) {
		childCancelInfo = workflowInfo
	})
	s.env.SetOnChildWorkflowCompletedListener(func(workflowInfo *workflow.Info, result converter.EncodedValue, err error) {
		childCompletedInfo = workflowInfo
	})
	s.env.OnRequestCancelExternalWorkflow(mock.Anything, childWorkflowID, "").Return(nil).Once()

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflowByID("default-test-workflow-id", signalName, signalValueDone)
	}, time.Second*1)
	s.env.ExecuteWorkflow(MyWorkflow)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	//s.True(cancelRequestCalled)
	s.NotNil(childCancelInfo)
	s.NotNil(childCompletedInfo)

}
func TestWorkflow(t *testing.T) {
	suite.Run(t, &CancelTestSuite{})
}
