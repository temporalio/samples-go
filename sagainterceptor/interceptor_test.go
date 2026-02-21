package sagainterceptor

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type orderInfo struct {
	ID       string
	IsDelete bool
}

var (
	orders = map[string]*orderInfo{}
	amount = 1
	suite  testsuite.WorkflowTestSuite
)

func createOrder(ctx context.Context, amount int) (string, error) {
	log.Println("enter createOrder")
	id := "abc"
	orders[id] = &orderInfo{
		ID: id,
	}
	return id, nil
}

func deleteOrder(ctx context.Context, id string) error {
	log.Println("enter deleteOrder, id:", id)
	orders[id].IsDelete = true
	return nil
}

func stockDeduct(ctx context.Context, in int) error {
	log.Println("enter stockDeduct")
	amount -= in
	return nil
}

func stockInc(ctx context.Context, in int) error {
	log.Println("enter stockInc")
	amount += in
	return nil
}

func createPay(ctx context.Context, in int) error {
	return errors.New("must fail")
}

func testConvertor(ctx workflow.Context, f workflow.Future, req ...interface{}) (rsp []interface{}, err error) {
	log.Println("convert req:", req[0].(int))
	var id string
	if err := f.Get(ctx, &id); err != nil {
		return nil, err
	}
	return []interface{}{id}, nil
}

func testWorkflow(ctx workflow.Context, a int) error {
	log.Println("enter workflow")
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	var id string
	if err := workflow.ExecuteActivity(ctx, "createOrder", a).Get(ctx, &id); err != nil {
		return err
	}
	log.Println("create order, id:", id)
	if err := workflow.ExecuteActivity(ctx, "stockDeduct", a).Get(ctx, nil); err != nil {
		return err
	}
	if err := workflow.ExecuteActivity(ctx, "createPay", a).Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

func TestWorkflow(t *testing.T) {
	env := suite.NewTestWorkflowEnvironment()
	intercept, _ := NewInterceptor(InterceptorOptions{
		WorkflowRegistry: map[string]SagaOptions{
			"testWorkflow": {},
		},
		ActivityRegistry: map[string]CompensationOptions{
			"createOrder": {
				ActivityType: "deleteOrder",
				Converter:    testConvertor,
			},
			"stockDeduct": {
				ActivityType: "stockInc",
			},
		},
	})
	env.SetWorkerOptions(worker.Options{Interceptors: []interceptor.WorkerInterceptor{intercept}})
	env.RegisterWorkflowWithOptions(testWorkflow, workflow.RegisterOptions{
		Name: "testWorkflow",
	})
	env.RegisterActivityWithOptions(createOrder, activity.RegisterOptions{
		Name: "createOrder",
	})
	env.RegisterActivityWithOptions(deleteOrder, activity.RegisterOptions{
		Name: "deleteOrder",
	})
	env.RegisterActivityWithOptions(stockDeduct, activity.RegisterOptions{
		Name: "stockDeduct",
	})
	env.RegisterActivityWithOptions(stockInc, activity.RegisterOptions{
		Name: "stockInc",
	})
	env.RegisterActivityWithOptions(createPay, activity.RegisterOptions{
		Name: "createPay",
	})

	env.ExecuteWorkflow(testWorkflow, 1)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	require.Equal(t, 1, len(orders))
	for _, order := range orders {
		require.True(t, order.IsDelete)
	}
	require.Equal(t, 1, amount)
	env.AssertExpectations(t)
}
