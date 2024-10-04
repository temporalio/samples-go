package safe_message_handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type updateCallback struct {
	accept   func()
	reject   func(error)
	complete func(interface{}, error)
}

func (uc *updateCallback) Accept() {
	if uc.accept == nil {
		return
	}
	uc.accept()
}

func (uc *updateCallback) Reject(err error) {
	if uc.reject == nil {
		return
	}
	uc.reject(err)
}

func (uc *updateCallback) Complete(success interface{}, err error) {
	if uc.complete == nil {
		return
	}
	uc.complete(success, err)
}

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(AssignNodesToJobsActivity)
	env.RegisterActivity(UnassignNodesForJobActivity)
	env.RegisterActivity(FindBadNodesActivity)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(StartCluster, nil)
	}, time.Second*2)

	for i := range 6 {
		env.RegisterDelayedCallback(func() {
			env.UpdateWorkflow(AssignNodesToJobs, fmt.Sprintf("TestUpdateID-%d", i), &updateCallback{
				complete: func(response interface{}, err error) {
					s.NoError(err)
					r := response.(ClusterManagerAssignNodesToJobResult)
					s.Equal(len(r.NodesAssigned), 2)
					env.RegisterDelayedCallback(func() {
						env.UpdateWorkflow(DeleteJob, fmt.Sprintf("TestUpdateID-%d", i), &updateCallback{}, ClusterManagerDeleteJobInput{
							JobName: fmt.Sprintf("TestJobID-%d", i),
						})
					}, 2*time.Duration(i)*time.Second)
				},
			}, ClusterManagerAssignNodesToJobInput{
				JobName:       fmt.Sprintf("TestJobID-%d", i),
				TotalNumNodes: 2,
			})
		}, time.Duration(i)*time.Second)
	}

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ShutdownCluster, nil)
	}, time.Minute)

	env.ExecuteWorkflow(ClusterManagerWorkflow, ClusterManagerInput{})

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_UpdateFailure() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(AssignNodesToJobsActivity)
	env.RegisterActivity(UnassignNodesForJobActivity)
	env.RegisterActivity(FindBadNodesActivity)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(StartCluster, nil)
	}, time.Second*2)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(AssignNodesToJobs, "TestUpdateID", &updateCallback{
			complete: func(response interface{}, err error) {
				s.NoError(err)
				env.RegisterDelayedCallback(func() {
					env.UpdateWorkflow(AssignNodesToJobs, "TestUpdateID2", &updateCallback{
						complete: func(response interface{}, err error) {
							s.Error(err)
							s.Contains(err.Error(), "not enough nodes to assign to job")
						},
					}, ClusterManagerAssignNodesToJobInput{
						JobName:       "little-task",
						TotalNumNodes: 3,
					})
				}, time.Second)
			},
		}, ClusterManagerAssignNodesToJobInput{
			JobName:       "big-task",
			TotalNumNodes: 24,
		})
	}, time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ShutdownCluster, nil)
	}, time.Minute)

	env.ExecuteWorkflow(ClusterManagerWorkflow, ClusterManagerInput{})

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}
