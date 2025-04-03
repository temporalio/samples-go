package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

// 测试资源获取和释放流程
func (s *UnitTestSuite) Test_ResourceAcquisitionAndRelease() {
	// 设置测试环境
	env := s.NewTestWorkflowEnvironment()

	// 模拟SignalWithStartResourcePoolWorkflowActivity的行为
	execution := &workflow.Execution{ID: "mockResourcePool", RunID: "mockRunID"}
	env.OnActivity(SignalWithStartResourcePoolWorkflowActivity,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(execution, nil)

	// 模拟发送资源获取信号
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ResourceAcquiredSignalName, "mockResourceChannelName")
	}, time.Millisecond*100)

	// 模拟发送到外部工作流的信号
	env.OnSignalExternalWorkflow(mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil)

	// 执行示例工作流
	env.ExecuteWorkflow(SampleWorkflowWithResourcePool, "test-resource-1", 2*time.Second)

	// 验证工作流执行完成且无错误
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	// 验证活动和信号
	env.AssertExpectations(s.T())
}

// 测试资源池中多个工作流争抢资源
func (s *UnitTestSuite) Test_ResourcePoolWorkflow() {
	// 设置测试环境
	env := s.NewTestWorkflowEnvironment()

	// 执行资源池工作流
	env.ExecuteWorkflow(ResourcePoolWorkflow, "TestNamespace", "test-resource-2", 1, 10*time.Second)

	// 模拟第一个工作流请求资源
	request1 := ResourceRequest{
		WorkflowID: "workflow-1",
		Priority:   1,
	}
	env.SignalWorkflow(RequestResourceSignalName, request1)

	// 等待处理第一个请求
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return workflow.Sleep(ctx, 100*time.Millisecond)
	})

	// 模拟第二个工作流请求资源（此时资源已被占用）
	request2 := ResourceRequest{
		WorkflowID: "workflow-2",
		Priority:   2,
	}
	env.SignalWorkflow(RequestResourceSignalName, request2)

	// 等待处理第二个请求
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return workflow.Sleep(ctx, 100*time.Millisecond)
	})

	// 模拟第一个工作流释放资源
	env.SignalWorkflow("resource-channel-workflow-1", "release")

	// 等待处理资源释放和分配给第二个工作流
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return workflow.Sleep(ctx, 100*time.Millisecond)
	})

	// 验证工作流仍在运行（资源池工作流永远不会自行完成）
	s.False(env.IsWorkflowCompleted())
}

// 运行所有测试
func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
