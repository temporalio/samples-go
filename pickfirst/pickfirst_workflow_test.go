package pickfirst

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(SampleActivity)
	env.OnActivity(SampleActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {
			// make branch 0 super fast so we don't have to wait sleep time in unit test
			if currentBranchID == 0 {
				totalDuration = time.Nanosecond
			}
			return SampleActivity(ctx, currentBranchID, totalDuration)
		})
	env.ExecuteWorkflow(SamplePickFirstWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal("Branch 0 done in 1ns.", result)
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_WorkflowBranchOne() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(SampleActivity)
	// Use .After to test what happens if branch 1 finished before branch 0
	env.OnActivity(SampleActivity, mock.Anything, 0, mock.Anything).After(5*time.Second).Return("Branch 0 done in 5s.", nil)
	env.OnActivity(SampleActivity, mock.Anything, 1, mock.Anything).After(1*time.Second).Return("Branch 1 done in 1s.", nil)
	env.ExecuteWorkflow(SamplePickFirstWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal("Branch 1 done in 1s.", result)
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_CancelActivity() {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	ctx, cancl := context.WithCancel(context.Background())
	env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: ctx,
	})
	env.RegisterActivity(SampleActivity)

	go func() {
		// Cancel the activity after 5s
		time.Sleep(5 * time.Second)
		cancl()
	}()
	_, err := env.ExecuteActivity(SampleActivity, 0, 10*time.Second)
	// Expect the activity to return a cancled error
	s.Error(err)
}

func (s *UnitTestSuite) Test_HeatbeatActivity() {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	env.RegisterActivity(SampleActivity)
	var heartBeatCount int
	env.SetOnActivityHeartbeatListener(func(activityInfo *activity.Info, details converter.EncodedValues) {
		var val string
		err := details.Get(&val)
		s.NoError(err)
		s.Equal("status-report-to-workflow", val)
		heartBeatCount++
	})

	_, err := env.ExecuteActivity(SampleActivity, 0, 5*time.Second)
	s.NoError(err)
	s.Equal(1, heartBeatCount)
}
